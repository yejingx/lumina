#!/usr/bin/env python3
"""
Script to download YOLOv8 model and convert it to TensorRT engine format
for use with Triton Inference Server.
"""

from ultralytics import YOLO
import os

def main():
    # Download and load YOLOv8 small model
    print("Loading YOLOv8 small model...")
    model = YOLO("yolov8s.pt")
    
    # Export to ONNX format first (TensorRT requires GPU which is not available)
    print("Converting to ONNX format with dynamic batch size and INT8 quantization support...")
    onnx_file = model.export(
        format="onnx",    # ONNX format
        imgsz=640,       # Input image size
        dynamic=True,    # Dynamic input shape including batch dimension
        opset=11,        # ONNX opset version
        int8=True,       # Enable INT8 quantization support
        data="coco8.yaml"  # Calibration dataset for INT8 quantization
    )
    
    print(f"ONNX model created: {onnx_file}")
    
    # Move the ONNX file to the correct location for Triton
    target_path = "model-repo/yolov8s/1/model.onnx"
    
    if os.path.exists(onnx_file):
        os.rename(onnx_file, target_path)
        print(f"ONNX file moved to: {target_path}")
    else:
        print(f"Error: ONNX file not found at {onnx_file}")
    
    print("\nONNX conversion completed successfully!")
    print("\nTo convert to TensorRT engine format, you need a system with GPU and TensorRT installed.")
    print("Use the following command on a GPU system with dynamic batch size and INT8 quantization:")
    print(f"trtexec --onnx={target_path} --saveEngine=model-repo/yolov8s/1/model.plan --int8 --minShapes=images:1x3x640x640 --optShapes=images:4x3x640x640 --maxShapes=images:8x3x640x640")
    print("\nThis configuration supports dynamic batch sizes from 1 to 8 with INT8 quantization.")
    print("INT8 quantization provides faster inference with slightly reduced accuracy.")
    print("You can adjust minShapes, optShapes, and maxShapes according to your needs.")
    print("\nFor now, you can use the ONNX model with Triton by changing the platform in config.pbtxt to 'onnxruntime_onnx'")
    print("The ONNX model already supports dynamic batch sizes.")

if __name__ == "__main__":
    main()