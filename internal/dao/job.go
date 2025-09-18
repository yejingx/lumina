package dao

import (
	"strings"

	"lumina/internal/model"
)

type DetectOptions struct {
	ModelName     string  `json:"modelName" binding:"required"`
	Labels        string  `json:"labels,omitempty"`
	ConfThreshold float32 `json:"confThreshold,omitempty"`
	IoUThreshold  float32 `json:"iouThreshold,omitempty"`
	Interval      int     `json:"interval,omitempty"`
}

func (d *DetectOptions) GetLabelMap() map[int]string {
	labels := strings.Split(d.Labels, ",")
	labelMap := make(map[int]string)
	for i, label := range labels {
		labelMap[i] = label
	}
	return labelMap
}

type VideoSegmentOptions struct {
	Interval int `json:"interval,omitempty"`
}

type JobSpec struct {
	Uuid         string               `json:"uuid" binding:"required"`
	Kind         model.JobKind        `json:"kind" binding:"required"`
	Status       model.JobStatus      `json:"status" binding:"required"`
	Input        string               `json:"input" binding:"required"`
	CreateTime   string               `json:"createTime" binding:"required,datetime=RFC3339"`
	UpdateTime   string               `json:"updateTime" binding:"required,datetime=RFC3339"`
	Detect       *DetectOptions       `json:"detect,omitempty"`
	VideoSegment *VideoSegmentOptions `json:"videoSegment,omitempty"`
	WorkflowId   int                  `json:"workflowId,omitempty"`
	Query        string               `json:"query,omitempty"`
}

type GetJobListResp struct {
	Items []JobSpec `json:"items"`
	Total int64     `json:"total"`
}
