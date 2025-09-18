import numpy as np
import triton_python_backend_utils as pb_utils
from collections import defaultdict
import json
import sys
import os
import time
sys.path.append(os.path.dirname(os.path.dirname(os.path.dirname(__file__))))
from triton_model_base import TritonPythonModelBase

# Simple SORT-like tracker (IoU + linear assignment via greedy)
# Good enough for demo; replace with ByteTrack/OC-SORT if needed.
# Note: Filtering is now handled in postproc module

class Track:
    __slots__ = ("box","id","age","miss")
    def __init__(self, box, tid):
        self.box = box
        self.id = tid
        self.age = 0
        self.miss = 0

class TritonPythonModel(TritonPythonModelBase):
    def initialize(self, args):
        # Initialize parent class and setup logging
        super().__init__()
        self._setup_logging('tracker')
        
        # Parse model config
        self.model_config = json.loads(args['model_config'])
        
        # Get configuration parameters
        params = self.model_config.get('parameters', {})
        self.iou_threshold = float(params.get('iou_threshold', {}).get('string_value', '0.45'))
        self.max_age = int(params.get('max_age', {}).get('string_value', '30'))
        
        self.seq_tracks = defaultdict(dict)  # seq_id -> {tid: Track}
        self.next_id = defaultdict(int)      # seq_id -> next tid
        
        self.logger.info(f"Tracker initialized with iou_threshold={self.iou_threshold}, max_age={self.max_age}")

    def _iou(self, a, b):
        ax1, ay1, ax2, ay2 = a
        bx1, by1, bx2, by2 = b
        inter_x1 = max(ax1, bx1); inter_y1 = max(ay1, by1)
        inter_x2 = min(ax2, bx2); inter_y2 = min(ay2, by2)
        iw = max(0.0, inter_x2 - inter_x1); ih = max(0.0, inter_y2 - inter_y1)
        inter = iw * ih
        if inter == 0: return 0.0
        area_a = (ax2-ax1)*(ay2-ay1); area_b = (bx2-bx1)*(by2-by1)
        return inter / (area_a + area_b - inter + 1e-6)

    def execute(self, requests):
        start_time = time.time()
        responses = []
        for req in requests:
            yolo_output = pb_utils.get_input_tensor_by_name(req, "YOLO_OUTPUT").as_numpy()
            seq_id_array = pb_utils.get_input_tensor_by_name(req, "SEQ_ID").as_numpy()
            
            # Handle different input shapes
            seq_id = int(seq_id_array.flatten()[0])

            # Parse YOLO output: assume format [detections, 6] where 6 = [x1,y1,x2,y2,conf,cls]
            boxes = yolo_output[:, :4]  # x1,y1,x2,y2
            scores = yolo_output[:, 4]  # confidence
            clsid = yolo_output[:, 5].astype(np.int64)  # class id
            tracks = self.seq_tracks[seq_id]

            self.logger.debug(f"Processing {len(boxes)} detections for tracking")

            used = set()
            out_ids = []
            out_boxes = []
            out_cls = []
            out_conf = []

            # Greedy match current tracks to detections
            for tid, tr in list(tracks.items()):
                tr.age += 1
                best_iou, best_j = 0.0, -1
                for j, db in enumerate(boxes):
                    if j in used: continue
                    iou = self._iou(tr.box, db)
                    if iou > best_iou:
                        best_iou, best_j = iou, j
                if best_iou >= self.iou_threshold and best_j >= 0:
                    tr.box = boxes[best_j]
                    tr.miss = 0
                    used.add(best_j)
                    out_ids.append(tid)
                    out_boxes.append(tr.box)
                    out_cls.append(clsid[best_j])
                    out_conf.append(scores[best_j])
                else:
                    tr.miss += 1

            # Create new tracks for unmatched detections
            new_tracks = 0
            for j, db in enumerate(boxes):
                if j in used: continue
                self.next_id[seq_id] += 1
                tid = self.next_id[seq_id]
                tracks[tid] = Track(db, tid)
                out_ids.append(tid)
                out_boxes.append(db)
                out_cls.append(clsid[j])
                out_conf.append(scores[j])
                new_tracks += 1
            
            if new_tracks > 0:
                self.logger.debug(f"Created {new_tracks} new tracks for seq_id={seq_id}")

            # Prune old tracks
            pruned_tracks = 0
            for tid in list(tracks.keys()):
                if tracks[tid].miss > self.max_age:
                    del tracks[tid]
                    pruned_tracks += 1
            
            if pruned_tracks > 0:
                self.logger.debug(f"Pruned {pruned_tracks} old tracks for seq_id={seq_id}")
            
            # Log tracking statistics
            active_tracks = len(tracks)
            self.logger.debug(f"Tracking stats - Active: {active_tracks}, Output: {len(out_ids)}, New: {new_tracks}, Pruned: {pruned_tracks}")

            if len(out_boxes)==0:
                tboxes = np.zeros((0,4), dtype=np.float32)
                tids = np.zeros((0,), dtype=np.int64)
                tcls = np.zeros((0,), dtype=np.int64)
                tconf = np.zeros((0,), dtype=np.float32)
            else:
                tboxes = np.stack(out_boxes).astype(np.float32)
                tids = np.asarray(out_ids, dtype=np.int64)
                tcls = np.asarray(out_cls, dtype=np.int64)
                tconf = np.asarray(out_conf, dtype=np.float32)

            responses.append(pb_utils.InferenceResponse(output_tensors=[
                pb_utils.Tensor("TBOXES", tboxes),
                pb_utils.Tensor("TIDS", tids),
                pb_utils.Tensor("TCLSID", tcls),
                pb_utils.Tensor("CONFIDENCE", tconf)
            ]))
        
        execute_time = time.time() - start_time
        self.logger.debug(f"Tracker execute time: {execute_time:.4f}s for {len(requests)} requests")
        return responses
