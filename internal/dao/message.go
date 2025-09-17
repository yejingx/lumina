package dao

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
	ImagePath string          `json:"imagePath"`
	JsonPath  string          `json:"jsonPath"`
	Boxes     []*DetectionBox `json:"boxes,omitempty"`
}

type Message struct {
	JobUuid     string          `json:"jobUuid"`
	Timestamp   int64           `json:"timestamp"`
	ImagePath   string          `json:"imagePath,omitempty"`
	DetectBoxes []*DetectionBox `json:"detectBoxes,omitempty"`
}
