import json
import numpy as np
import triton_python_backend_utils as pb_utils
import sys
import os
import time
sys.path.append(os.path.dirname(os.path.dirname(os.path.dirname(__file__))))
from triton_model_base import TritonPythonModelBase

def letterbox_to_original(boxes, meta):
    """
    Convert bounding boxes from letterbox coordinates to original image coordinates
    Args:
        boxes: numpy array with shape [N, 4] containing [x1, y1, x2, y2]
        meta: metadata array containing [ih, iw, oh, ow, scale, pad_t, pad_l, pad_b, pad_r, ...]
    Returns:
        boxes in original image coordinates
    """
    ih, iw, oh, ow, scale, pad_t, pad_l, pad_b, pad_r = meta[:9]
    
    out = boxes.copy()
    # Remove padding
    out[:, [0,2]] -= pad_l
    out[:, [1,3]] -= pad_t
    # Scale back to original size
    out /= scale
    # Clip to image boundaries
    out[:, 0::2] = np.clip(out[:, 0::2], 0, iw-1)
    out[:, 1::2] = np.clip(out[:, 1::2], 0, ih-1)
    return out

class TritonPythonModel(TritonPythonModelBase):
    def initialize(self, args):
        # Initialize parent class and setup logging
        super().__init__()
        self._setup_logging('postproc')
        
        # Parse model config
        self.model_config = json.loads(args['model_config'])
        
        # Get configuration parameters
        params = self.model_config.get('parameters', {})
        self.conf_threshold = float(params.get('conf_threshold', {}).get('string_value', '0.25'))
        self.iou_threshold = float(params.get('iou_threshold', {}).get('string_value', '0.45'))
        self.track_classes = [int(i) for i in json.loads(params.get('track_classes', {}).get('string_value', '[0, 1, 2, 3, 4]'))]
        
        self.logger.info(f"Postproc initialized with conf_threshold={self.conf_threshold}, iou_threshold={self.iou_threshold}, track_classes={self.track_classes}")

    def execute(self, requests):
        start_time = time.time()
        responses = []
        for request in requests:
            # Get YOLO raw output and META
            yolo_raw = pb_utils.get_input_tensor_by_name(request, "YOLO_RAW").as_numpy()
            meta = pb_utils.get_input_tensor_by_name(request, "META").as_numpy()
            
            # Remove batch dimension if present
            if len(yolo_raw.shape) == 3:
                yolo_raw = yolo_raw[0]  # Shape should be [8400, 84] for YOLOv8/11
            
            # Log YOLO raw output info
            self.logger.debug(f"YOLO raw shape: {yolo_raw.shape}")
            
            # Transpose if needed - YOLO output should be [4+class_logits..., features]
            if yolo_raw.shape[0] < yolo_raw.shape[1]:
                yolo_raw = yolo_raw.T
                self.logger.debug(f"Transposed to shape: {yolo_raw.shape}")
            
            # No detections
            if yolo_raw.shape[0] == 0:
                output_tensor = pb_utils.Tensor("DETECTIONS", np.zeros((0, 6), dtype=np.float32))
                responses.append(pb_utils.InferenceResponse(output_tensors=[output_tensor]))
                continue

            # YOLOv8 output format: [cx, cy, w, h, class_probs...]
            # Shape: [num_anchors, 4 + num_classes]
            # No separate objectness confidence in YOLOv8
            # Extract box coordinates (cx, cy, w, h)
            boxes_cxcywh = yolo_raw[:, :4]
            class_probs = yolo_raw[:, 4:]
            
            # Get max class probability and class ID
            confidences = np.max(class_probs, axis=1)
            class_ids = np.argmax(class_probs, axis=1)
            
            self.logger.debug(f"Confidences range: {confidences.min():.3f}-{confidences.max():.3f}")

            class_mask = np.isin(class_ids, self.track_classes)
            confidence_mask = confidences > self.conf_threshold
            valid_mask = class_mask & confidence_mask
            if np.sum(valid_mask) == 0:
                output_tensor = pb_utils.Tensor("DETECTIONS", np.zeros((0, 6), dtype=np.float32))
                responses.append(pb_utils.InferenceResponse(output_tensors=[output_tensor]))
                continue
            
            boxes_cxcywh = boxes_cxcywh[valid_mask]
            class_ids = class_ids[valid_mask]
            confidences = confidences[valid_mask]

            self.logger.debug(f"Class IDs after filtering: {len(class_ids)}")

            # Convert from center format to corner format
            cx, cy, w, h = boxes_cxcywh[:, 0], boxes_cxcywh[:, 1], boxes_cxcywh[:, 2], boxes_cxcywh[:, 3]
            x1 = cx - w / 2
            y1 = cy - h / 2
            x2 = cx + w / 2
            y2 = cy + h / 2
            boxes_filtered = np.stack([x1, y1, x2, y2], axis=1)
            
            keep_indices = self._nms(boxes_filtered, confidences, iou_threshold=self.iou_threshold)

            self.logger.debug(f"Keep indices: {len(keep_indices)}")
            
            if len(keep_indices) == 0:
                output = np.zeros((0, 6), dtype=np.float32)
            else:
                # Convert boxes to original image coordinates
                boxes_original = letterbox_to_original(boxes_filtered[keep_indices], meta)
                
                # Final output format: [x1, y1, x2, y2, confidence, class_id]
                output = np.column_stack([
                    boxes_original,
                    confidences[keep_indices],
                    class_ids[keep_indices]
                ]).astype(np.float32)
            
            # Create output tensor
            output_tensor = pb_utils.Tensor("DETECTIONS", output)
            responses.append(pb_utils.InferenceResponse(output_tensors=[output_tensor]))
        
        execute_time = time.time() - start_time
        self.logger.debug(f"Postproc execute time: {execute_time:.4f}s for {len(requests)} requests")
        return responses
    
    def _nms(self, boxes, scores, iou_threshold):
        """Simple Non-Maximum Suppression"""
        if len(boxes) == 0:
            return []
        
        # Sort by confidence score (descending)
        indices = np.argsort(scores)[::-1]
        keep = []
        
        while len(indices) > 0:
            # Pick the detection with highest confidence
            current = indices[0]
            keep.append(current)
            
            if len(indices) == 1:
                break
            
            # Calculate IoU with remaining boxes
            current_box = boxes[current]
            remaining_boxes = boxes[indices[1:]]
            
            ious = self._calculate_iou(current_box, remaining_boxes)
            
            # Keep only boxes with IoU below threshold
            indices = indices[1:][ious < iou_threshold]
        
        return keep
    
    def _calculate_iou(self, box, boxes):
        """Calculate IoU between one box and multiple boxes"""
        x1, y1, x2, y2 = box
        xx1, yy1, xx2, yy2 = boxes[:, 0], boxes[:, 1], boxes[:, 2], boxes[:, 3]
        
        # Calculate intersection
        inter_x1 = np.maximum(x1, xx1)
        inter_y1 = np.maximum(y1, yy1)
        inter_x2 = np.minimum(x2, xx2)
        inter_y2 = np.minimum(y2, yy2)
        
        inter_w = np.maximum(0, inter_x2 - inter_x1)
        inter_h = np.maximum(0, inter_y2 - inter_y1)
        inter_area = inter_w * inter_h
        
        # Calculate union
        box_area = (x2 - x1) * (y2 - y1)
        boxes_area = (xx2 - xx1) * (yy2 - yy1)
        union_area = box_area + boxes_area - inter_area
        
        # Calculate IoU
        iou = inter_area / (union_area + 1e-6)
        return iou