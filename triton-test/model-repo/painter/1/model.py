import numpy as np
import cv2
import triton_python_backend_utils as pb_utils
import sys
import os
import time
sys.path.append(os.path.dirname(os.path.dirname(os.path.dirname(__file__))))
from triton_model_base import TritonPythonModelBase

PALETTE = np.array([
    [255, 56, 56], [255, 159, 56], [255, 255, 56],
    [56, 255, 56], [56, 255, 255], [56, 56, 255],
    [255, 56, 255], [128, 128, 128]
], dtype=np.uint8)

# COCO class names
COCO_CLASSES = [
    'person', 'bicycle', 'car', 'motorcycle', 'airplane', 'bus', 'train', 'truck', 'boat', 'traffic light',
    'fire hydrant', 'stop sign', 'parking meter', 'bench', 'bird', 'cat', 'dog', 'horse', 'sheep', 'cow',
    'elephant', 'bear', 'zebra', 'giraffe', 'backpack', 'umbrella', 'handbag', 'tie', 'suitcase', 'frisbee',
    'skis', 'snowboard', 'sports ball', 'kite', 'baseball bat', 'baseball glove', 'skateboard', 'surfboard',
    'tennis racket', 'bottle', 'wine glass', 'cup', 'fork', 'knife', 'spoon', 'bowl', 'banana', 'apple',
    'sandwich', 'orange', 'broccoli', 'carrot', 'hot dog', 'pizza', 'donut', 'cake', 'chair', 'couch',
    'potted plant', 'bed', 'dining table', 'toilet', 'tv', 'laptop', 'mouse', 'remote', 'keyboard', 'cell phone',
    'microwave', 'oven', 'toaster', 'sink', 'refrigerator', 'book', 'clock', 'vase', 'scissors', 'teddy bear',
    'hair drier', 'toothbrush'
]

def letterbox_to_original(boxes, meta):
    ih, iw, oh, ow, scale, pad_t, pad_l, pad_b, pad_r, _ = meta[:10]
    
    out = boxes.copy()
    out[:, [0,2]] -= pad_l
    out[:, [1,3]] -= pad_t
    out /= scale
    out[:, 0::2] = np.clip(out[:, 0::2], 0, iw-1)
    out[:, 1::2] = np.clip(out[:, 1::2], 0, ih-1)
    return out

class TritonPythonModel(TritonPythonModelBase):
    def initialize(self, args):
        # Initialize parent class and setup logging
        super().__init__()
        self._setup_logging('painter')
        self.logger.info(f"Painter initialized with {len(COCO_CLASSES)} COCO classes, {len(PALETTE)} colors")

    def execute(self, requests):
        start_time = time.time()
        responses = []
        for req in requests:
            frame = pb_utils.get_input_tensor_by_name(req, "FRAME").as_numpy()
            meta  = pb_utils.get_input_tensor_by_name(req, "META").as_numpy()
            boxes = pb_utils.get_input_tensor_by_name(req, "TBOXES").as_numpy()
            clsid = pb_utils.get_input_tensor_by_name(req, "TCLSID").as_numpy()
            tids  = pb_utils.get_input_tensor_by_name(req, "TIDS").as_numpy()
            confidence = pb_utils.get_input_tensor_by_name(req, "CONFIDENCE").as_numpy()
            
            self.logger.debug(f"Input shapes - frame: {frame.shape}, boxes: {boxes.shape}, clsid: {clsid.shape}, tids: {tids.shape}, confidence: {confidence.shape}")

            out = frame.copy()

            # Project to original image
            boxes_orig = letterbox_to_original(boxes.astype(np.float32), meta)

            for i, box in enumerate(boxes_orig):
                cid = int(clsid[i])
                tid = int(tids[i])
                conf = float(confidence[i])
                color = PALETTE[tid % len(PALETTE)].tolist()
                class_name = COCO_CLASSES[cid] if 0 <= cid < len(COCO_CLASSES) else f"class_{cid}"

                x1,y1,x2,y2 = box.astype(int)
                cv2.rectangle(out, (x1,y1), (x2,y2), color, 2)
                label = f"{class_name} ID:{tid} {conf:.2f}"
                (tw, th), _ = cv2.getTextSize(label, cv2.FONT_HERSHEY_SIMPLEX, 0.6, 2)
                cv2.rectangle(out, (x1, y1-th-6), (x1+tw+6, y1), color, -1)
                cv2.putText(out, label, (x1+3, y1-4), cv2.FONT_HERSHEY_SIMPLEX, 0.6, (0,0,0), 2, cv2.LINE_AA)

            responses.append(pb_utils.InferenceResponse(output_tensors=[pb_utils.Tensor("OUT_FRAME", out)]))
        
        execute_time = time.time() - start_time
        self.logger.debug(f"Painter execute time: {execute_time:.4f}s for {len(requests)} requests")
        return responses
