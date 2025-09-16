package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"os"
	"path"
	"sync"
	"time"

	"github.com/Trendyol/go-triton-client/base"
	"github.com/Trendyol/go-triton-client/client/grpc"
	"github.com/sirupsen/logrus"
	"gocv.io/x/gocv"

	"lumina/internal/dao"
	"lumina/pkg/log"
)

type JobExecutor struct {
	tritonCli base.Client
	ctx       context.Context
	cancel    context.CancelFunc
	job       *dao.JobSpec
	logger    *logrus.Entry
	doneChan  chan struct{}
	workDir   string
}

func NewJobExecutor(tritonCli base.Client, workDir string, parentCtx context.Context, job *dao.JobSpec) (*JobExecutor, error) {
	workDir = path.Join(workDir, job.Uuid)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(parentCtx)
	return &JobExecutor{
		tritonCli: tritonCli,
		ctx:       ctx,
		cancel:    cancel,
		job:       job,
		logger:    log.GetLogger(ctx).WithField("job", job.Uuid),
		doneChan:  make(chan struct{}),
		workDir:   workDir,
	}, nil
}

func (e *JobExecutor) Job() *dao.JobSpec {
	return e.job
}

func (e *JobExecutor) Start() error {
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

	if e.job.Detect == nil {
		e.logger.Info("empty job finished")
		return nil
	}

	video, err := gocv.VideoCaptureFile(e.job.Detect.Input)
	if err != nil {
		return fmt.Errorf("failed to open input video: %v", err)
	}

	go e.runJob(video)

	return nil
}

func (e *JobExecutor) Stop() {
	e.cancel()
	<-e.Done()
}

func (e *JobExecutor) Done() <-chan struct{} {
	return e.doneChan
}

func (e *JobExecutor) inferRoutine(frameCh <-chan gocv.Mat) {
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

		if err := e.saveResult(&frame, boxes); err != nil {
			e.logger.WithError(err).Errorf("save result error")
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

func (e *JobExecutor) runJob(input *gocv.VideoCapture) error {
	e.logger.Info("job started")

	fps := input.Get(gocv.VideoCaptureFPS)
	width := int(input.Get(gocv.VideoCaptureFrameWidth))
	height := int(input.Get(gocv.VideoCaptureFrameHeight))
	logrus.Infof("Video properties: %dx%d @ %.2f FPS", width, height, fps)

	frameChan := make(chan gocv.Mat, 10)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		e.inferRoutine(frameChan)
	}()

	defer func() {
		e.logger.Info("job stopped")
		input.Close()
		close(frameChan)
		wg.Wait()
		close(e.doneChan)
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
			return nil
		default:
		}

		frame := gocv.NewMat()
		if ok := input.Read(&frame); !ok {
			frame.Close()
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

	return nil
}

func (e *JobExecutor) saveResult(frame *gocv.Mat, boxes []*DetectionBox) error {
	ts := time.Now().UnixNano()
	result := &DetectionResult{
		JobId:     e.job.Uuid,
		Timestamp: ts,
		Boxes:     boxes,
	}
	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal detection result error: %w", err)
	}
	imagePath := path.Join(e.workDir, fmt.Sprintf("%d.jpg", ts))
	if !gocv.IMWrite(imagePath, *frame) {
		return fmt.Errorf("write image file error")
	}

	jsonPath := path.Join(e.workDir, fmt.Sprintf("%d.json", ts))
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

type DetectionBox struct {
	X1         int     `json:"x1,omitempty"`
	Y1         int     `json:"y1,omitempty"`
	X2         int     `json:"x2,omitempty"`
	Y2         int     `json:"y2,omitempty"`
	Confidence float32 `json:"confidence,omitempty"`
	ClassId    int     `json:"classId,omitempty"`
	Label      string  `json:"label,omitempty"`
}

type DetectionResult struct {
	JobId     string          `json:"jobId"`
	Timestamp int64           `json:"timestamp"`
	Boxes     []*DetectionBox `json:"boxes,omitempty"`
}

func drawDetections(frame *gocv.Mat, boxes []*DetectionBox) gocv.Mat {
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
func performInference(client base.Client, frame *gocv.Mat, modelName string, labelMap map[int]string) (gocv.Mat, []*DetectionBox, error) {
	frameBytes := frame.ToBytes()

	// Create input tensors
	// FRAME input - image data
	frameInput := grpc.NewInferInput("FRAME", "BYTES", []int64{int64(frame.Rows()), int64(frame.Cols()), 3}, nil)
	err := frameInput.SetData(frameBytes, true)
	if err != nil {
		return gocv.NewMat(), nil, fmt.Errorf("failed to set FRAME input data: %v", err)
	}
	frameInput.SetDatatype("UINT8")

	outputs := []base.InferOutput{
		grpc.NewInferOutput("DETECTIONS", map[string]any{"binary_data": false}),
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

	var boxes []*DetectionBox
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

		boxes = append(boxes, &DetectionBox{
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
