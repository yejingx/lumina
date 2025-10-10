# Triton YOLOE Video Pipeline

This repository contains a Triton model-repo layout and a client for processing MP4 videos through a pipeline that performs:

- YOLOE instance segmentation (person + vehicle classes)
- SORT-like tracking (assigns track IDs per SEQ_ID)
- Rendering masks, bounding boxes and track IDs on frames
- Outputting annotated MP4

## Quickstart
1. Put your exported YOLOE ONNX or TensorRT plan into `model-repo/yoloe/1/` and ensure `yoloe/config.pbtxt` matches exported output names.

2. Start Triton:
   docker run --gpus all -it --rm   -p8000:8000 -p8001:8001 -p8002:8002   -v $(pwd)/model-repo:/models triton tritonserver --model-repository=/models

3. Run client:
   python3 client/video_pipeline_client.py --in input.mp4 --out output.mp4 --url http://localhost:8000

## Notes
- If your YOLOE export uses different output names/format, adapt the ensemble and painter/tracker accordingly.
- For better tracking you can replace tracker with ByteTrack/OC-SORT implementations.


perf 

```
perf_analyzer -m pipeline --async --concurrency-range 1:1   --shape FRAME:1920,1080,3   --shape SEQ_ID:1 --measurement-request-count 2000
```