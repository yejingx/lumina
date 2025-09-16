package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"log"
	"time"

	"github.com/Trendyol/go-triton-client/base"
	"github.com/Trendyol/go-triton-client/client/grpc"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gocv.io/x/gocv"
)

// COCO数据集的80个类别名称
var cocoClassNames = []string{
	"person", "bicycle", "car", "motorcycle", "airplane", "bus", "train", "truck", "boat", "traffic light",
	"fire hydrant", "stop sign", "parking meter", "bench", "bird", "cat", "dog", "horse", "sheep", "cow",
	"elephant", "bear", "zebra", "giraffe", "backpack", "umbrella", "handbag", "tie", "suitcase", "frisbee",
	"skis", "snowboard", "sports ball", "kite", "baseball bat", "baseball glove", "skateboard", "surfboard", "tennis racket", "bottle",
	"wine glass", "cup", "fork", "knife", "spoon", "bowl", "banana", "apple", "sandwich", "orange",
	"broccoli", "carrot", "hot dog", "pizza", "donut", "cake", "chair", "couch", "potted plant", "bed",
	"dining table", "toilet", "tv", "laptop", "mouse", "remote", "keyboard", "cell phone", "microwave", "oven",
	"toaster", "sink", "refrigerator", "book", "clock", "vase", "scissors", "teddy bear", "hair drier", "toothbrush",
}

var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Tools for lumina",
	Long:  `Various tools and utilities for lumina application.`,
}

var (
	inputVideo    string
	outputVideo   string
	tritonURL     string
	modelName     string
	confThreshold float64
	iouThreshold  float64
	trackClasses  string
)

var inferCmd = &cobra.Command{
	Use:   "infer",
	Short: "Perform video inference using Triton server",
	Long:  `Process video files using Triton inference server for object detection and tracking`,
	Run: func(cmd *cobra.Command, args []string) {
		if inputVideo == "" || outputVideo == "" {
			logrus.Fatal("Both input and output video paths are required")
		}

		// Create Triton client
		client, err := grpc.NewClient(
			tritonURL,
			false, // verbose logging
			30,    // connection timeout in seconds
			30,    // network timeout in seconds
			false, // use SSL
			true,  // insecure connection
			nil,   // existing gRPC connection
			nil,   // logger
		)
		if err != nil {
			logrus.Fatalf("Failed to create Triton client: %v", err)
		}

		// Process video
		err = processVideo(client, inputVideo, outputVideo, modelName, confThreshold, iouThreshold, trackClasses)
		if err != nil {
			logrus.Fatalf("Error processing video: %v", err)
		}

		logrus.Info("Video processing completed successfully")
	},
}

// processVideo processes the input video frame by frame using Triton inference
func processVideo(client base.Client, inputPath, outputPath, modelName string,
	confThreshold, iouThreshold float64, trackClasses string) error {
	video, err := gocv.VideoCaptureFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input video: %v", err)
	}
	defer video.Close()

	fps := video.Get(gocv.VideoCaptureFPS)
	width := int(video.Get(gocv.VideoCaptureFrameWidth))
	height := int(video.Get(gocv.VideoCaptureFrameHeight))

	logrus.Infof("Video properties: %dx%d @ %.2f FPS", width, height, fps)

	writer, err := gocv.VideoWriterFile(outputPath, "mp4v", fps, width, height, true)
	if err != nil {
		return fmt.Errorf("failed to create output video writer: %v", err)
	}
	defer writer.Close()

	frame := gocv.NewMat()
	defer frame.Close()

	frameCount := 0
	totalInferenceTime := time.Duration(0)

	for {
		if ok := video.Read(&frame); !ok {
			break
		}

		if frame.Empty() {
			continue
		}

		frameCount++

		start := time.Now()
		processedFrame, err := performInference(client, frame, modelName, confThreshold, iouThreshold, trackClasses)
		if err != nil {
			log.Printf("Error in inference for frame %d: %v", frameCount, err)
			processedFrame = frame.Clone()
		}
		inferenceTime := time.Since(start)
		totalInferenceTime += inferenceTime

		// Write processed frame to output video
		writer.Write(processedFrame)
		processedFrame.Close()

		if frameCount%30 == 0 {
			logrus.Infof("Processed %d frames, avg inference time: %.2fms",
				frameCount, float64(totalInferenceTime.Nanoseconds())/float64(frameCount)/1e6)
		}
	}

	logrus.Infof("Total frames processed: %d", frameCount)
	logrus.Infof("Average inference time: %.2fms",
		float64(totalInferenceTime.Nanoseconds())/float64(frameCount)/1e6)

	return nil
}

// getClassName 根据类别ID获取COCO类别名称
func getClassName(classID int) string {
	if classID >= 0 && classID < len(cocoClassNames) {
		return cocoClassNames[classID]
	}
	return fmt.Sprintf("Class %d", classID)
}

// drawDetections draws detection boxes on frame
// detections: slice of float32 values with shape [N, 6] containing [x1, y1, x2, y2, confidence, class_id]
func drawDetections(frame gocv.Mat, detections []float32) gocv.Mat {
	annotatedFrame := frame.Clone()

	if len(detections) == 0 {
		return annotatedFrame
	}

	// Process detections in groups of 6 (x1, y1, x2, y2, confidence, class_id)
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

		gocv.Rectangle(&annotatedFrame, image.Rect(x1, y1, x2, y2), color.RGBA{0, 255, 0, 255}, 2)

		className := getClassName(classID)
		label := fmt.Sprintf("%s: %.2f", className, confidence)
		labelSize := gocv.GetTextSize(label, gocv.FontHersheySimplex, 0.5, 2)

		gocv.Rectangle(&annotatedFrame, image.Rect(x1, y1-labelSize.Y-10, x1+labelSize.X, y1), color.RGBA{0, 255, 0, 255}, -1)
		gocv.PutText(&annotatedFrame, label, image.Pt(x1, y1-5), gocv.FontHersheySimplex, 0.5, color.RGBA{0, 0, 0, 255}, 2)
	}

	return annotatedFrame
}

// performInference performs inference on a single frame using Triton
func performInference(client base.Client, frame gocv.Mat, modelName string,
	confThreshold, iouThreshold float64, trackClasses string) (gocv.Mat, error) {
	frameBytes := frame.ToBytes()

	// Create input tensors
	// FRAME input - image data
	frameInput := grpc.NewInferInput("FRAME", "BYTES", []int64{int64(frame.Rows()), int64(frame.Cols()), 3}, nil)
	err := frameInput.SetData(frameBytes, true)
	if err != nil {
		return gocv.NewMat(), fmt.Errorf("failed to set FRAME input data: %v", err)
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
		return gocv.NewMat(), fmt.Errorf("inference failed: %v", err)
	}

	detections, err := response.AsFloat32Slice("DETECTIONS")
	if err != nil {
		return gocv.NewMat(), fmt.Errorf("failed to get detection data: %v", err)
	}

	processedFrame := drawDetections(frame, detections)

	return processedFrame, nil
}

func init() {
	inferCmd.Flags().StringVarP(&inputVideo, "input", "i", "in.mp4", "Input video file path")
	inferCmd.Flags().StringVarP(&outputVideo, "output", "o", "out.mp4", "Output video file path")
	inferCmd.Flags().StringVar(&tritonURL, "triton-url", "localhost:8001", "Triton server URL")
	inferCmd.Flags().StringVar(&modelName, "model", "pipeline", "Model name")
	inferCmd.Flags().Float64Var(&confThreshold, "conf-threshold", 0.25, "Confidence threshold for detection")
	inferCmd.Flags().Float64Var(&iouThreshold, "iou-threshold", 0.45, "IoU threshold for NMS")
	inferCmd.Flags().StringVar(&trackClasses, "track-classes", "0,1,2,3,4,5", "Comma-separated list of class IDs to track")

	toolsCmd.AddCommand(inferCmd)
}
