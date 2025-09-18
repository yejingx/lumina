import argparse, cv2, numpy as np, time
from itertools import count
from tritonclient.grpc import InferenceServerClient, InferInput, InferRequestedOutput

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

def run(in_mp4, out_mp4, url, model="pipeline"):
    cap = cv2.VideoCapture(in_mp4)
    fps = cap.get(cv2.CAP_PROP_FPS) or 25.0
    w  = int(cap.get(cv2.CAP_PROP_FRAME_WIDTH))
    h  = int(cap.get(cv2.CAP_PROP_FRAME_HEIGHT))
    writer = cv2.VideoWriter(out_mp4, cv2.VideoWriter_fourcc(*"mp4v"), fps, (w,h))
    
    cli = InferenceServerClient(url=url)

    for fid in count(0):
        ok, frame = cap.read()
        if not ok: break
        inp_frame = InferInput("FRAME", [h, w, 3], "UINT8")
        inp_frame.set_data_from_numpy(frame)

        out_detections = InferRequestedOutput("DETECTIONS")
        start_time = time.time()
        res = cli.infer(model_name=model, inputs=[inp_frame], outputs=[out_detections])
        infer_time = time.time() - start_time
        
        detections = res.as_numpy("DETECTIONS")
        print(f"Frame {fid}: Inference time: {infer_time:.4f}s, Detections: {len(detections)}")
        
        # Draw detections on frame
        annotated_frame = draw_detections(frame, detections)
        writer.write(annotated_frame)

    cap.release(); writer.release()

if __name__ == "__main__":
    ap = argparse.ArgumentParser()
    ap.add_argument("--in", dest="in_mp4", default="in.mp4")
    ap.add_argument("--out", dest="out_mp4", default="out.mp4")
    ap.add_argument("--url", default="localhost:8001")
    ap.add_argument("--model", default="pipeline")
    args = ap.parse_args()
    run(args.in_mp4, args.out_mp4, args.url, args.model)
