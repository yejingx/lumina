#!/usr/bin/env python3
"""
Image Pipeline Client for Triton Inference Server
Processes a single image through the YOLOv8 detection pipeline
"""

import argparse
import cv2
import numpy as np
from tritonclient.http import InferenceServerClient, InferInput, InferRequestedOutput
import os
import time

def draw_detections(frame, detections):
    """
    Draw detection boxes on frame
    Args:
        frame: input frame
        detections: numpy array with shape [N, 6] containing [x1, y1, x2, y2, confidence, class_id]
    Returns:
        annotated frame
    """
    annotated_frame = frame.copy()
    
    if len(detections) == 0:
        return annotated_frame
    
    for detection in detections:
        x1, y1, x2, y2, confidence, class_id = detection
        x1, y1, x2, y2 = int(x1), int(y1), int(x2), int(y2)
        
        # Draw bounding box
        cv2.rectangle(annotated_frame, (x1, y1), (x2, y2), (0, 255, 0), 2)
        
        # Draw label
        label = f"Class {int(class_id)}: {confidence:.2f}"
        label_size = cv2.getTextSize(label, cv2.FONT_HERSHEY_SIMPLEX, 0.5, 2)[0]
        cv2.rectangle(annotated_frame, (x1, y1 - label_size[1] - 10), 
                     (x1 + label_size[0], y1), (0, 255, 0), -1)
        cv2.putText(annotated_frame, label, (x1, y1 - 5), 
                   cv2.FONT_HERSHEY_SIMPLEX, 0.5, (0, 0, 0), 2)
    
    return annotated_frame

def process_image(input_path, output_path, url="localhost:8000", model="pipeline"):
    """
    Process a single image through the detection pipeline
    
    Args:
        input_path (str): Path to input image
        output_path (str): Path to save output image
        url (str): Triton server URL
        model (str): Model name to use
    
    Returns:
        bool: True if successful, False otherwise
    """
    # Load input image
    frame = cv2.imread(input_path)
    if frame is None:
        print(f"错误：无法加载图片 {input_path}")
        return False
    
    h, w, _ = frame.shape
    print(f"输入图片尺寸: {w}x{h}")
    
    # Remove protocol scheme from URL if present
    if url.startswith('http://'):
        url = url[7:]
    elif url.startswith('https://'):
        url = url[8:]
    
    try:
        # Create Triton client
        client = InferenceServerClient(url=url)
        
        # Create input tensors
        inp_frame = InferInput("FRAME", [h, w, 3], "UINT8")
        inp_frame.set_data_from_numpy(frame)
        
        # Create output tensor
        out_detections = InferRequestedOutput("DETECTIONS")
        
        print("正在运行推理...")
        
        # Run inference
        start_time = time.time()
        result = client.infer(
            model_name=model,
            inputs=[inp_frame],
            outputs=[out_detections]
        )
        end_time = time.time()
        infer_time = end_time - start_time
        print(f"推理时间: {infer_time:.4f}s")
        
        # Get detection results
        detections = result.as_numpy("DETECTIONS")
        print(f"检测结果数量: {len(detections)}")
        
        # Draw detections on frame
        output_frame = draw_detections(frame, detections)
        
        # Save result
        cv2.imwrite(output_path, output_frame)
        print(f"结果已保存到: {output_path}")
        
        # Print detection details
        if len(detections) > 0:
            print("✅ 检测成功！发现以下目标:")
            for i, detection in enumerate(detections):
                x1, y1, x2, y2, confidence, class_id = detection
                print(f"  目标 {i+1}: 类别 {int(class_id)}, 置信度 {confidence:.3f}, 位置 ({int(x1)},{int(y1)}) - ({int(x2)},{int(y2)})")
            return True
        else:
            print("⚠️  未检测到目标。")
            return True  # Still successful inference, just no detections
            
    except Exception as e:
        print(f"❌ 推理过程中出现错误: {e}")
        return False

def main():
    parser = argparse.ArgumentParser(description="图片检测管道客户端")
    parser.add_argument("--input", "-i", default="in.jpg", help="输入图片路径")
    parser.add_argument("--output", "-o", default="out.jpg", help="输出图片路径")
    parser.add_argument("--url", default="localhost:8000", help="Triton服务器URL")
    parser.add_argument("--model", default="pipeline", help="模型名称")
    
    args = parser.parse_args()
    
    # Check if input file exists
    if not os.path.exists(args.input):
        print(f"错误：输入文件不存在: {args.input}")
        return 1
    
    # Process the image
    success = process_image(args.input, args.output, args.url, args.model)
    
    return 0 if success else 1

if __name__ == "__main__":
    exit(main())