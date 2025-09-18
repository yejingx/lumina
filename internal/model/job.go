package model

import "time"

type JobStatus int

const (
	JobStatusStopped JobStatus = iota
	JobStatusRunning
)

type JobKind int

func (k JobKind) String() string {
	switch k {
	case JobKindDetect:
		return "detect"
	case JobKindVideoSegment:
		return "video_segment"
	default:
		return "unknown"
	}
}

const (
	JobKindDetect JobKind = iota
	JobKindVideoSegment
)

type DetectOptions struct {
	ModelName     string  `json:"model_name" gorm:"NOT NULL"`
	Interval      int     `json:"interval" gorm:"default:3"`
	Labels        string  `json:"labels" gorm:"default:''"`
	ConfThreshold float32 `json:"conf_threshold" gorm:"default:0.25"`
	IoUThreshold  float32 `json:"iou_threshold" gorm:"default:0.45"`
}

type VideoSegmentOptions struct {
	Interval int `json:"interval" gorm:"default:30"`
}

type Job struct {
	Id           int                  `json:"id" gorm:"primaryKey"`
	Uuid         string               `json:"uuid" gorm:"unique"`
	Kind         JobKind              `json:"kind" gorm:"default:0"`
	Input        string               `json:"input" gorm:"NOT NULL"`
	Status       JobStatus            `json:"status" gorm:"default:0"`
	CreateTime   time.Time            `json:"create_time" gorm:"datetime;autoCreateTime"`
	UpdateTime   time.Time            `json:"update_time" gorm:"datetime;autoCreateTime;autoUpdateTime"`
	Detect       *DetectOptions       `json:"detect" gorm:"type:json"`
	VideoSegment *VideoSegmentOptions `json:"video_segment" gorm:"type:json"`
	WorkflowId   int                  `json:"workflow_id" gorm:"default:0"`
	Query        string               `json:"query" gorm:"type:text;default:''"`
}
