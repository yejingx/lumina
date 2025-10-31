package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"

	"lumina/internal/dao"
	"lumina/internal/model"
	"lumina/pkg/log"
)

type Consumer struct {
	conf            *Config
	ctx             context.Context
	cancel          context.CancelFunc
	consumer        *nsq.Consumer
	wg              sync.WaitGroup
	logger          *logrus.Entry
	workflowManager *WorkflowManager
	// influx
	influxClient influxdb2.Client
	writeAPI     api.WriteAPIBlocking
}

func NewConsumer(conf *Config) (*Consumer, error) {
	ctx, cancel := context.WithCancel(context.Background())

	logger := log.GetLogger(ctx).WithField("component", "consumer")

	config := nsq.NewConfig()
	config.MsgTimeout = time.Minute
	config.MaxInFlight = 10
	config.MaxAttempts = 2

	consumer, err := nsq.NewConsumer(conf.NSQ.Topic, "lumina-consumer", config)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create NSQ consumer: %w", err)
	}

	c := &Consumer{
		conf:            conf,
		ctx:             ctx,
		cancel:          cancel,
		consumer:        consumer,
		logger:          logger,
		workflowManager: NewWorkflowManager(ctx),
	}

	// init influxdb client if enabled
	if conf.InfluxDB.Enabled {
		client := influxdb2.NewClient(conf.InfluxDB.URL, conf.InfluxDB.Token)
		c.influxClient = client
		c.writeAPI = client.WriteAPIBlocking(conf.InfluxDB.Org, conf.InfluxDB.Bucket)
	}

	consumer.AddHandler(c)

	return c, nil
}

func (c *Consumer) HandleMessage(message *nsq.Message) error {
	c.logger.Debugf("Received NSQ message: %s", string(message.Body))
	message.DisableAutoResponse()

	var msg dao.DeviceMessage
	if err := json.Unmarshal(message.Body, &msg); err != nil {
		c.logger.WithError(err).Error("Failed to unmarshal NSQ message")
		return err
	}

	c.logger.WithFields(logrus.Fields{
		"jobUuid":   msg.JobUuid,
		"timestamp": msg.Timestamp,
		"imagePath": msg.ImagePath,
		"boxCount":  len(msg.DetectBoxes),
	}).Info("Processing detection result message")

	job, err := model.GetJobByUuid(msg.JobUuid)
	if err != nil {
		c.logger.WithError(err).Errorf("Failed to get job by uuid %s", msg.JobUuid)
		return err
	} else if job == nil {
		message.Finish()
		return nil
	}

	wf, err := job.Workflow()
	if err != nil {
		c.logger.WithError(err).Errorf("Failed to get workflow for job %s", msg.JobUuid)
		return err
	} else if wf == nil {
		message.Finish()
		return nil
	}

	var resp *OpenAIResponse
	if job.Kind == model.JobKindVideoSegment {
		resp, err = c.workflowManager.VideoCompletion(wf, c.conf.S3.UrlPrefix()+msg.VideoPath)
	} else {
		resp, err = c.workflowManager.ImageCompletion(wf, c.conf.S3.UrlPrefix()+msg.ImagePath, msg.DetectBoxes)
	}
	if err != nil {
		c.logger.WithError(err).Errorf("Failed to call WorkflowManager API for job %s", msg.JobUuid)
		return err
	}
	answer := parseResponseContent(resp.Choices[0].Message.Content)
	c.logger.Infof("workflow response for job %s: %+v", msg.JobUuid, answer)

	m := msg.ToModel(job)
	m.WorkflowResp = &model.WorkflowResp{
		TotalTokens: resp.Usage.TotalTokens,
		Answer:      answer.Reason,
		Confidence:  answer.Confidence,
		Match:       answer.Match,
		RawContent:  resp.Choices[0].Message.Content,
	}
	// if wf.ResultFilter != nil && wf.ResultFilter.Match(answer) {
	// 	m.Alerted = true
	// }
	if answer.Match {
		m.Alerted = true
	}
	if err := model.AddMessage(m); err != nil {
		c.logger.WithError(err).Errorf("Failed to add message to DB for job %s", msg.JobUuid)
		return err
	}

	// write event to influxdb
	c.writeInfluxEvents(job, &msg)

	message.Finish()
	c.logger.Debugf("Successfully processed message for job %s", msg.JobUuid)
	return nil
}

func (c *Consumer) writeInfluxEvents(job *model.Job, msg *dao.DeviceMessage) {
	if c.writeAPI == nil || !c.conf.InfluxDB.Enabled {
		return
	}

	// Event time derived from message timestamp (see dao conversion logic)
	// Note: original timestamp unit may be microseconds; follow existing model conversion
	evtTime := time.Unix(msg.Timestamp/1000000000, msg.Timestamp%1000000000)
	if evtTime.IsZero() {
		evtTime = time.Now()
	}

	// Base message event
	baseTags := map[string]string{
		"job_uuid": string(job.Uuid),
		"job_kind": string(job.Kind),
	}
	baseFields := map[string]any{
		"count": 1,
	}
	p := influxdb2.NewPoint(influxMeasurementMessage, baseTags, baseFields, evtTime)
	if err := c.writeAPI.WritePoint(c.ctx, p); err != nil {
		c.logger.WithError(err).Warn("Failed to write message event to InfluxDB")
	}

	// Detection boxes events if detection job
	if job.Kind == model.JobKindDetect && len(msg.DetectBoxes) > 0 {
		for _, box := range msg.DetectBoxes {
			if box == nil {
				continue
			}
			tags := map[string]string{
				"job_uuid": job.Uuid,
				"label":    box.Label,
			}
			fields := map[string]any{
				"confidence": box.Confidence,
				"x1":         box.X1,
				"y1":         box.Y1,
				"x2":         box.X2,
				"y2":         box.Y2,
			}
			pbox := influxdb2.NewPoint(influxMeasurementDetection, tags, fields, evtTime)
			if err := c.writeAPI.WritePoint(c.ctx, pbox); err != nil {
				c.logger.WithError(err).Warn("Failed to write detection event to InfluxDB")
			}
		}
	}
}

func (c *Consumer) Start() error {
	c.logger.Info("Starting NSQ consumer...")

	err := c.consumer.ConnectToNSQDs(c.conf.NSQ.NSQDAddrs)
	if err != nil {
		return fmt.Errorf("failed to connect to NSQs: %w", err)
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		<-c.ctx.Done()
		c.consumer.Stop()
	}()

	return nil
}

func (c *Consumer) Stop() {
	c.cancel()
	c.wg.Wait()
	if c.influxClient != nil {
		c.influxClient.Close()
	}
}

const influxMeasurementMessage = "lumina_message"
const influxMeasurementDetection = "lumina_detection"
