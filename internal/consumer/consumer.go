package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"

	"lumina/internal/dao"
	"lumina/pkg/log"
)

type Consumer struct {
	conf     *Config
	ctx      context.Context
	cancel   context.CancelFunc
	consumer *nsq.Consumer
	wg       sync.WaitGroup
	logger   *logrus.Entry
	dify     *Dify
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
		conf:     conf,
		ctx:      ctx,
		cancel:   cancel,
		consumer: consumer,
		logger:   logger,
		dify:     NewDify(ctx, conf.Dify),
	}

	consumer.AddHandler(c)

	return c, nil
}

func (c *Consumer) HandleMessage(message *nsq.Message) error {
	c.logger.Debugf("Received NSQ message: %s", string(message.Body))
	message.DisableAutoResponse()

	var msg dao.Message
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

	resp, err := c.dify.ChatCompletion(c.conf.S3.UrlPrefix()+msg.ImagePath, &msg, "图中有没有出现带手提包的人")
	if err != nil {
		c.logger.WithError(err).Errorf("Failed to call Dify API for job %s", msg.JobUuid)
		return err
	}
	c.logger.Infof("DifyChatCompletion response for job %s: %s", msg.JobUuid, resp)

	message.Finish()
	c.logger.Debugf("Successfully processed message for job %s", msg.JobUuid)
	return nil
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
	c.wg.Wait() // 等待所有goroutine完成
}
