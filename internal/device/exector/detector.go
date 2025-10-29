package exector

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Trendyol/go-triton-client/base"
	tritonGrpc "github.com/Trendyol/go-triton-client/client/grpc"
	"github.com/minio/minio-go/v7"
	"github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"
	"gocv.io/x/gocv"

	"lumina/internal/dao"
	"lumina/internal/device/config"
	"lumina/internal/device/metadata"
	"lumina/internal/model"
	"lumina/internal/utils"
	"lumina/pkg/log"
)

type Detector struct {
	tritonCli       base.Client
	ctx             context.Context
	cancel          context.CancelFunc
	wg              *sync.WaitGroup
	job             *dao.JobSpec
	logger          *logrus.Entry
	status          model.ExectorStatus
	workDir         string
	conf            *config.Config
	nsqProducer     *nsq.Producer
	minioCli        *minio.Client
	deviceInfo      *metadata.DeviceInfo
	triggerCount    int
	lastTriggerTime time.Time
}

func NewDetector(conf *config.Config, deviceInfo *metadata.DeviceInfo, parentCtx context.Context,
	minioCli *minio.Client, nsqProducer *nsq.Producer, job *dao.JobSpec) (*Detector, error) {
	if job.Detect == nil {
		return nil, fmt.Errorf("job %s detect is nil", job.Uuid)
	}

	tritonCli, err := tritonGrpc.NewClient(
		conf.Triton.ServerAddr,
		false, // verbose logging
		30,    // connection timeout in seconds
		30,    // network timeout in seconds
		false, // use ssl
		true,  // insecure connection
		nil,   // existing grpc connection
		nil,   // logger
	)
	if err != nil {
		return nil, err
	}

	workDir := path.Join(conf.JobDir(), job.Uuid)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(parentCtx)
	return &Detector{
		tritonCli:       tritonCli,
		ctx:             ctx,
		cancel:          cancel,
		wg:              &sync.WaitGroup{},
		job:             job,
		status:          model.ExectorStatusStopped,
		logger:          log.GetLogger(ctx).WithField("job", job.Uuid),
		workDir:         workDir,
		conf:            conf,
		nsqProducer:     nsqProducer,
		minioCli:        minioCli,
		deviceInfo:      deviceInfo,
		lastTriggerTime: time.Now(),
	}, nil
}

func (e *Detector) Job() *dao.JobSpec {
	return e.job
}

func (e *Detector) Status() model.ExectorStatus {
	return e.status
}

func (e *Detector) Start() error {
	if isLive, err := e.tritonCli.IsServerLive(e.ctx, nil); err != nil {
		return err
	} else if !isLive {
		return errors.New("triton server is not live")
	}

	if isReady, err := e.tritonCli.IsServerReady(e.ctx, nil); err != nil {
		return err
	} else if !isReady {
		return errors.New("triton server is not ready")
	}

	if isReady, err := e.tritonCli.IsModelReady(e.ctx, e.job.Detect.ModelName, "1", nil); err != nil {
		return err
	} else if !isReady {
		return errors.New("triton model is not ready")
	}

	video, err := gocv.VideoCaptureFile(e.job.Input())
	if err != nil {
		return fmt.Errorf("failed to open input video: %v", err)
	}

	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		e.uploadRoutine()
	}()

	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		e.logger.Info("detect job started")
		e.status = model.ExectorStatusRunning
		e.runJob(video)
		e.logger.Info("detect job stopped")
	}()

	return nil
}

func (e *Detector) Stop() {
	e.cancel()
	e.wg.Wait()
	e.status = model.ExectorStatusStopped
}

func (e *Detector) inferRoutine(frameCh <-chan gocv.Mat) {
	frameCount := 0
	totalInferenceTime := time.Duration(0)
	labelMap := e.job.Detect.GetLabelMap()
	lastLogTime := time.Now()

	for frame := range frameCh {
		frameCount++

		start := time.Now()
		processedFrame, boxes, err := performInference(e.tritonCli, &frame, e.job.Detect.ModelName, labelMap)
		if err != nil {
			e.logger.WithError(err).Errorf("inference error")
			processedFrame = frame.Clone()
		}
		inferenceTime := time.Since(start)
		totalInferenceTime += inferenceTime

		needSave := false
		if len(boxes) > 0 {
			e.triggerCount += 1
			if time.Since(e.lastTriggerTime) > time.Duration(e.job.Detect.TriggerInterval)*time.Second &&
				e.triggerCount >= e.job.Detect.TriggerCount {
				needSave = true
				e.lastTriggerTime = time.Now()
			}
		} else {
			e.triggerCount = 0
		}

		if needSave {
			if err := e.saveResult(&frame, boxes); err != nil {
				e.logger.WithError(err).Errorf("save result error")
			}
		}

		processedFrame.Close()

		if time.Since(lastLogTime) > 5*time.Second {
			e.logger.Infof("processed %d frames in %v, avg inference time: %v", frameCount, totalInferenceTime, totalInferenceTime/time.Duration(frameCount))
			lastLogTime = time.Now()
			frameCount = 0
			totalInferenceTime = time.Duration(0)
		}

		frame.Close()
	}
}

func (e *Detector) runJob(input *gocv.VideoCapture) {
	fps := input.Get(gocv.VideoCaptureFPS)
	width := int(input.Get(gocv.VideoCaptureFrameWidth))
	height := int(input.Get(gocv.VideoCaptureFrameHeight))
	logrus.Infof("Video properties: %dx%d @ %.2f FPS", width, height, fps)

	frameChan := make(chan gocv.Mat, 10)

	defer func() {
		input.Close()
		close(frameChan)
	}()

	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		e.inferRoutine(frameChan)
	}()

	lastFrameTime := time.Now()
	var interval time.Duration
	if e.job.Detect.Interval <= 0 {
		interval = 3 * time.Second
	} else {
		interval = time.Duration(e.job.Detect.Interval) * time.Millisecond
	}

	for {
		select {
		case <-e.ctx.Done():
			return
		default:
		}

		frame := gocv.NewMat()
		if ok := input.Read(&frame); !ok {
			frame.Close()
			e.status = model.ExectorStatusFinished
			break
		}

		if frame.Empty() {
			frame.Close()
			continue
		}

		if time.Since(lastFrameTime) < interval {
			frame.Close()
			continue
		}
		lastFrameTime = time.Now()

		select {
		case frameChan <- frame:
		default:
			e.logger.Warnf("frame dropped, frame pool is full")
			frame.Close()
		}
	}
}

func (e *Detector) saveResult(frame *gocv.Mat, boxes []*dao.DetectionBox) error {
	ts := time.Now().UnixNano()
	imagePath := path.Join(e.workDir, fmt.Sprintf("%d.jpg", ts))
	jsonPath := path.Join(e.workDir, fmt.Sprintf("%d.json", ts))

	result := &dao.DetectionResult{
		JobId:     e.job.Uuid,
		Timestamp: ts,
		ImagePath: imagePath,
		JsonPath:  jsonPath,
		Boxes:     boxes,
	}
	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal detection result error: %w", err)
	}

	if !gocv.IMWrite(imagePath, *frame) {
		return fmt.Errorf("write image file error")
	}

	tmpPath := jsonPath + ".tmp"
	if err := os.WriteFile(tmpPath, jsonData, 0644); err != nil {
		os.Remove(imagePath)
		return fmt.Errorf("write json file error")
	}
	if err := os.Rename(tmpPath, jsonPath); err != nil {
		return fmt.Errorf("rename json file error")
	}
	return nil
}

func drawDetections(frame *gocv.Mat, boxes []*dao.DetectionBox) gocv.Mat {
	annotatedFrame := frame.Clone()

	if len(boxes) == 0 {
		return annotatedFrame
	}

	// Process detections in groups of 6 (x1, y1, x2, y2, confidence, class_id)
	for _, box := range boxes {
		label := fmt.Sprintf("%s: %.2f", box.Label, box.Confidence)
		labelSize := gocv.GetTextSize(label, gocv.FontHersheySimplex, 0.5, 2)

		gocv.Rectangle(&annotatedFrame, image.Rect(box.X1, box.Y1, box.X2, box.Y2), color.RGBA{0, 255, 0, 255}, 2)
		gocv.Rectangle(&annotatedFrame, image.Rect(box.X1, box.Y1-labelSize.Y-10, box.X1+labelSize.X, box.Y1), color.RGBA{0, 255, 0, 255}, -1)
		gocv.PutText(&annotatedFrame, label, image.Pt(box.X1, box.Y1-5), gocv.FontHersheySimplex, 0.5, color.RGBA{0, 0, 0, 255}, 2)
	}

	return annotatedFrame
}

// performInference performs inference on a single frame using Triton
func performInference(client base.Client, frame *gocv.Mat, modelName string, labelMap map[int]string) (gocv.Mat, []*dao.DetectionBox, error) {
	frameBytes := frame.ToBytes()

	// Create input tensors
	// FRAME input - image data
	frameInput := tritonGrpc.NewInferInput("FRAME", "BYTES", []int64{int64(frame.Rows()), int64(frame.Cols()), 3}, nil)
	err := frameInput.SetData(frameBytes, true)
	if err != nil {
		return gocv.NewMat(), nil, fmt.Errorf("failed to set FRAME input data: %v", err)
	}
	frameInput.SetDatatype("UINT8")

	outputs := []base.InferOutput{
		tritonGrpc.NewInferOutput("DETECTIONS", map[string]any{"binary_data": false}),
	}

	response, err := client.Infer(
		context.Background(),
		modelName,
		"1",
		[]base.InferInput{frameInput},
		outputs,
		nil,
	)
	if err != nil {
		return gocv.NewMat(), nil, fmt.Errorf("inference failed: %v", err)
	}

	detections, err := response.AsFloat32Slice("DETECTIONS")
	if err != nil {
		return gocv.NewMat(), nil, fmt.Errorf("failed to get detection data: %v", err)
	}

	var boxes []*dao.DetectionBox
	// detections: slice of float32 values with shape [N, 6] containing [x1, y1, x2, y2, confidence, class_id]
	for i := 0; i < len(detections); i += 6 {
		if i+5 >= len(detections) {
			break
		}

		x1 := int(detections[i])
		y1 := int(detections[i+1])
		x2 := int(detections[i+2])
		y2 := int(detections[i+3])
		confidence := detections[i+4]
		classID := int(detections[i+5])

		className, exists := labelMap[classID]
		if !exists || className == "" {
			continue
		}

		boxes = append(boxes, &dao.DetectionBox{
			X1:         x1,
			Y1:         y1,
			X2:         x2,
			Y2:         y2,
			Confidence: confidence,
			ClassId:    classID,
			Label:      className,
		})
	}

	processedFrame := drawDetections(frame, boxes)

	return processedFrame, boxes, nil
}

func (e *Detector) uploadRoutine() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		if err := e.listAndUpload(); err != nil {
			e.logger.WithError(err).Errorf("list and upload failed")
		}

		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (e *Detector) listAndUpload() error {
	return filepath.WalkDir(e.workDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".json") {
			return nil
		}

		jsonData, err := os.ReadFile(path)
		if err != nil {
			e.logger.WithError(err).Errorf("read JSON file %s failed", path)
			return nil
		}

		var result dao.DetectionResult
		if err := json.Unmarshal(jsonData, &result); err != nil {
			e.logger.WithError(err).Errorf("unmarshal JSON file %s failed", path)
			return nil
		}

		fileName := strings.TrimSuffix(d.Name(), ".json")
		imgPath := filepath.Join(e.workDir, fileName+".jpg")

		var ts time.Time
		if result.Timestamp != 0 {
			ts = time.Unix(result.Timestamp/1000000000, result.Timestamp%1000000000)
		} else {
			if jsonInfo, err := d.Info(); err == nil {
				ts = jsonInfo.ModTime()
			} else {
				ts = time.Now()
			}
		}
		minioPath := fmt.Sprintf("/%s/%04d/%02d/%02d/%s/%s.jpg",
			*e.deviceInfo.Uuid, ts.Year(), ts.Month(), ts.Day(), result.JobId, fileName)

		ctx, cancel := context.WithTimeout(e.ctx, 30*time.Second)
		defer cancel()
		if err := utils.UploadFileToMinio(ctx, e.minioCli, e.conf.S3.Bucket, imgPath, minioPath); err != nil {
			e.logger.WithError(err).Errorf("upload image %s to minio failed", imgPath)
			return nil
		}

		msg := &dao.DeviceMessage{
			JobUuid:     result.JobId,
			Timestamp:   ts.UnixNano(),
			ImagePath:   minioPath,
			DetectBoxes: result.Boxes,
		}
		msgData, _ := json.Marshal(msg)
		if err := e.nsqProducer.Publish(e.conf.NSQ.Topic, msgData); err != nil {
			e.logger.WithError(err).Errorf("publish to NSQ failed for %s", path)
			return nil
		}

		os.Remove(path)
		os.Remove(imgPath)

		e.logger.Infof("successfully processed %s: uploaded image to %s and sent to NSQ", path, minioPath)
		return nil
	})
}
