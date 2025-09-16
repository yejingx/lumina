package dao

import (
	"lumina/internal/model"
	"strings"
)

type DetectOptions struct {
	ModelName     string  `json:"modelName" binding:"required"`
	Input         string  `json:"input" binding:"required"`
	Interval      int     `json:"interval,omitempty"`
	Labels        string  `json:"labels,omitempty"`
	ConfThreshold float32 `json:"confThreshold,omitempty"`
	IoUThreshold  float32 `json:"iouThreshold,omitempty"`
}

func (d *DetectOptions) GetLabelMap() map[int]string {
	labels := strings.Split(d.Labels, ",")
	labelMap := make(map[int]string)
	for i, label := range labels {
		labelMap[i] = label
	}
	return labelMap
}

type JobSpec struct {
	Uuid       string          `json:"uuid" binding:"required"`
	Status     model.JobStatus `json:"status" binding:"required"`
	CreateTime string          `json:"createTime" binding:"required,datetime=RFC3339"`
	UpdateTime string          `json:"updateTime" binding:"required,datetime=RFC3339"`
	Detect     *DetectOptions  `json:"detect,omitempty"`
}

type GetJobListResp struct {
	Items []JobSpec `json:"items"`
	Total int64     `json:"total"`
}
