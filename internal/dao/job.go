package dao

import (
	"lumina/internal/model"
)

type DetectOptions struct {
	ModelName     string  `json:"modelName" binding:"required"`
	Input         string  `json:"input" binding:"required"`
	Interval      int     `json:"interval,omitempty"`
	Labels        string  `json:"labels,omitempty"`
	ConfThreshold float32 `json:"confThreshold,omitempty"`
	IoUThreshold  float32 `json:"iouThreshold,omitempty"`
}

type JobSpec struct {
	Uuid        string          `json:"uuid" binding:"required"`
	Status      model.JobStatus `json:"status" binding:"required"`
	CreatedTime string          `json:"createdTime" binding:"required,datetime=RFC3339"`
	Detect      *DetectOptions  `json:"detect,omitempty"`
}

type GetJobListResp struct {
	Items []JobSpec `json:"items"`
	Total int64     `json:"total"`
}
