import numpy as np
import cv2
import triton_python_backend_utils as pb_utils
import sys
import os
import time
sys.path.append(os.path.dirname(os.path.dirname(os.path.dirname(__file__))))
from triton_model_base import TritonPythonModelBase

IMG_SIZE = 640

class TritonPythonModel(TritonPythonModelBase):
    def initialize(self, args):
        # Initialize parent class and setup logging
        super().__init__()
        self._setup_logging('preproc')
        self.logger.info(f"Preproc initialized with target size: {IMG_SIZE}x{IMG_SIZE}")

    def execute(self, requests):
        start_time = time.time()
        responses = []
        for request in requests:
            frame = pb_utils.get_input_tensor_by_name(request, "FRAME").as_numpy()
            # HWC BGR uint8
            ih, iw, _ = frame.shape
            self.logger.debug(f"Input frame shape: {frame.shape}, dtype: {frame.dtype}")
            
            scale = min(IMG_SIZE / iw, IMG_SIZE / ih)
            ow, oh = int(round(iw * scale)), int(round(ih * scale))
            self.logger.debug(f"Scaling: {iw}x{ih} -> {ow}x{oh} (scale: {scale:.3f})")
            
            resized = cv2.resize(frame, (ow, oh), interpolation=cv2.INTER_LINEAR)
            canvas = np.full((IMG_SIZE, IMG_SIZE, 3), 114, dtype=np.uint8)
            pad_t = (IMG_SIZE - oh) // 2
            pad_l = (IMG_SIZE - ow) // 2
            self.logger.debug(f"Padding: top={pad_t}, left={pad_l}")
            
            canvas[pad_t:pad_t+oh, pad_l:pad_l+ow] = resized
            # to NCHW float32, 0-1
            img = canvas.astype(np.float32) / 255.0
            img = img.transpose(2,0,1)[None, ...]
            self.logger.debug(f"Output tensor shape: {img.shape}, dtype: {img.dtype}")

            meta = np.zeros((12,), dtype=np.float32)
            meta[:10] = [ih, iw, oh, ow, scale, pad_t, pad_l, IMG_SIZE - oh - pad_t, IMG_SIZE - ow - pad_l, IMG_SIZE]

            out_img = pb_utils.Tensor("IMAGES", img)
            out_meta = pb_utils.Tensor("META", meta)
            responses.append(pb_utils.InferenceResponse(output_tensors=[out_img, out_meta]))
        
        execute_time = time.time() - start_time
        self.logger.debug(f"Preproc execute time: {execute_time:.4f}s for {len(requests)} requests")
        return responses
