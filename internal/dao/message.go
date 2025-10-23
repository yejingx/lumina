package dao

import (
	"time"

	"lumina/internal/model"
)

type DetectionBox struct {
	X1         int     `json:"x1,omitempty"`
	Y1         int     `json:"y1,omitempty"`
	X2         int     `json:"x2,omitempty"`
	Y2         int     `json:"y2,omitempty"`
	Confidence float32 `json:"confidence,omitempty"`
	ClassId    int     `json:"classId,omitempty"`
	Label      string  `json:"label,omitempty"`
}

func (b DetectionBox) ToModel() *model.DetectionBox {
	return &model.DetectionBox{
		X1:         b.X1,
		Y1:         b.Y1,
		X2:         b.X2,
		Y2:         b.Y2,
		Confidence: b.Confidence,
		ClassId:    b.ClassId,
		Label:      b.Label,
	}
}

type DetectionResult struct {
	JobId     string          `json:"jobId"`
	Timestamp int64           `json:"timestamp"`
	ImagePath string          `json:"imagePath"`
	JsonPath  string          `json:"jsonPath"`
	Boxes     []*DetectionBox `json:"boxes,omitempty"`
}

type DeviceMessage struct {
	JobUuid     string          `json:"jobUuid"`
	Timestamp   int64           `json:"timestamp"` // us
	ImagePath   string          `json:"imagePath,omitempty"`
	DetectBoxes []*DetectionBox `json:"detectBoxes,omitempty"`
	VideoPath   string          `json:"videoPath,omitempty"`
}

func (m DeviceMessage) ToModel(job *model.Job) *model.Message {
	mdl := &model.Message{
		JobId:     job.Id,
		Timestamp: time.Unix(m.Timestamp/1000000000, m.Timestamp%1000000000),
		ImagePath: m.ImagePath,
		VideoPath: m.VideoPath,
	}
	if m.DetectBoxes != nil {
		mdl.DetectBoxes = make(model.DetectionBoxSlice, len(m.DetectBoxes))
		for i, box := range m.DetectBoxes {
			mdl.DetectBoxes[i] = box.ToModel()
		}
	}
	return mdl
}

type WorkflowResp struct {
	Answer string `json:"answer,omitempty"`
}

func (w WorkflowResp) ToModel() *model.WorkflowResp {
	return &model.WorkflowResp{
		Answer: w.Answer,
	}
}

type MessageSpec struct {
	Id           int             `json:"id"`
	JobId        int             `json:"jobId"`
	Timestamp    string          `json:"timestamp"`
	ImagePath    string          `json:"imagePath,omitempty"`
	DetectBoxes  []*DetectionBox `json:"detectBoxes,omitempty"`
	VideoPath    string          `json:"videoPath,omitempty"`
	CreateTime   string          `json:"createTime"`
	WorkflowResp *WorkflowResp   `json:"workflowResp,omitempty"`
}

func FromMessageModel(msg *model.Message) *MessageSpec {
	if msg == nil {
		return nil
	}
	m := &MessageSpec{}
	m.Id = msg.Id
	m.JobId = msg.JobId
	m.Timestamp = msg.Timestamp.Format(time.RFC3339)
	m.ImagePath = msg.ImagePath
	m.VideoPath = msg.VideoPath
	m.CreateTime = msg.CreateTime.Format(time.RFC3339)

	if msg.DetectBoxes != nil {
		m.DetectBoxes = make([]*DetectionBox, len(msg.DetectBoxes))
		for i, box := range msg.DetectBoxes {
			m.DetectBoxes[i] = &DetectionBox{
				X1:         box.X1,
				Y1:         box.Y1,
				X2:         box.X2,
				Y2:         box.Y2,
				Confidence: box.Confidence,
				ClassId:    box.ClassId,
				Label:      box.Label,
			}
		}
	}

	if msg.WorkflowResp != nil {
		m.WorkflowResp = &WorkflowResp{
			Answer: msg.WorkflowResp.Answer,
		}
	}

	return m
}

type CreateMessageRequest struct {
	JobId        int             `json:"jobId" binding:"required"`
	Timestamp    int64           `json:"timestamp" binding:"required"`
	ImagePath    string          `json:"imagePath,omitempty"`
	DetectBoxes  []*DetectionBox `json:"detectBoxes,omitempty"`
	VideoPath    string          `json:"videoPath,omitempty"`
	WorkflowResp *WorkflowResp   `json:"workflowResp,omitempty"`
}

func (req *CreateMessageRequest) ToModel() *model.Message {
	msg := &model.Message{
		JobId:     req.JobId,
		Timestamp: time.Unix(req.Timestamp, 0),
		ImagePath: req.ImagePath,
		VideoPath: req.VideoPath,
	}

	if req.DetectBoxes != nil {
		msg.DetectBoxes = make(model.DetectionBoxSlice, len(req.DetectBoxes))
		for i, box := range req.DetectBoxes {
			msg.DetectBoxes[i] = box.ToModel()
		}
	}

	if req.WorkflowResp != nil {
		msg.WorkflowResp = req.WorkflowResp.ToModel()
	}

	return msg
}

type CreateMessageResponse struct {
	Id int `json:"id"`
}

type ListMessagesRequest struct {
	JobId int `json:"jobId" form:"jobId"`
	Start int `json:"start" form:"start" binding:"min=0"`
	Limit int `json:"limit" form:"limit" binding:"min=0,max=50"`
}

type ListMessagesResponse struct {
	Items []MessageSpec `json:"items"`
	Total int64         `json:"total"`
}

type AlertMessageSpec struct {
	Id         int         `json:"id"`
	Message    MessageSpec `json:"message"`
	CreateTime string      `json:"createTime"`
}

func FromAlertMessageModel(am *model.AlertMessage) *AlertMessageSpec {
	if am == nil {
		return nil
	}
	return &AlertMessageSpec{
		Id:         am.Id,
		Message:    *FromMessageModel(&am.Message),
		CreateTime: am.CreateTime.Format(time.RFC3339),
	}
}

type ListAlertMessagesRequest struct {
	JobId int `json:"jobId" form:"jobId"`
	Start int `json:"start" form:"start" binding:"min=0"`
	Limit int `json:"limit" form:"limit" binding:"min=0,max=50"`
}

type ListAlertMessagesResponse struct {
	Items []AlertMessageSpec `json:"items"`
	Total int64              `json:"total"`
}
