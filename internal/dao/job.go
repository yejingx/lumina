package dao

import (
	"errors"
	"strings"
	"time"

	"lumina/internal/model"
	"lumina/pkg/str"
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
	Device       *DeviceSpec          `json:"device,omitempty"`
}

func FromJobModel(job *model.Job) (*JobSpec, error) {
	if job == nil {
		return nil, errors.New("job is nil")
	}
	j := &JobSpec{
		Uuid:       job.Uuid,
		Kind:       job.Kind,
		Status:     job.Status,
		Input:      job.Input,
		CreateTime: job.CreateTime.Format(time.RFC3339),
		UpdateTime: job.UpdateTime.Format(time.RFC3339),
		WorkflowId: job.WorkflowId,
		Query:      job.Query,
	}

	if job.Detect != nil {
		j.Detect = &DetectOptions{
			ModelName:     job.Detect.ModelName,
			Labels:        job.Detect.Labels,
			ConfThreshold: job.Detect.ConfThreshold,
			IoUThreshold:  job.Detect.IoUThreshold,
			Interval:      job.Detect.Interval,
		}
	}

	if job.VideoSegment != nil {
		j.VideoSegment = &VideoSegmentOptions{
			Interval: job.VideoSegment.Interval,
		}
	}

	if job.DeviceId != 0 {
		device, err := model.GetDeviceById(job.DeviceId)
		if err != nil {
			return nil, err
		} else if device != nil {
			j.Device = FromDeviceModel(device)
		}
	}
	return j, nil
}

type CreateJobRequest struct {
	Kind         model.JobKind        `json:"kind" binding:"required"`
	Input        string               `json:"input" binding:"required"`
	Detect       *DetectOptions       `json:"detect,omitempty"`
	VideoSegment *VideoSegmentOptions `json:"videoSegment,omitempty"`
	WorkflowId   int                  `json:"workflowId,omitempty"`
	Query        string               `json:"query,omitempty"`
}

func (req *CreateJobRequest) ToModel() *model.Job {
	job := &model.Job{
		Uuid:       str.GenDeviceId(16),
		Kind:       req.Kind,
		Input:      req.Input,
		Status:     model.JobStatusStopped,
		WorkflowId: req.WorkflowId,
		Query:      req.Query,
	}

	// 设置检测选项
	if req.Detect != nil {
		job.Detect = &model.DetectOptions{
			ModelName:     req.Detect.ModelName,
			Labels:        req.Detect.Labels,
			ConfThreshold: req.Detect.ConfThreshold,
			IoUThreshold:  req.Detect.IoUThreshold,
			Interval:      req.Detect.Interval,
		}
		// 设置默认值
		if job.Detect.Interval == 0 {
			job.Detect.Interval = 3
		}
		if job.Detect.ConfThreshold == 0 {
			job.Detect.ConfThreshold = 0.25
		}
		if job.Detect.IoUThreshold == 0 {
			job.Detect.IoUThreshold = 0.45
		}
	}

	// 设置视频分割选项
	if req.VideoSegment != nil {
		job.VideoSegment = &model.VideoSegmentOptions{
			Interval: req.VideoSegment.Interval,
		}
		// 设置默认值
		if job.VideoSegment.Interval == 0 {
			job.VideoSegment.Interval = 30
		}
	}
	return job
}

type CreateJobResponse struct {
	Uuid string `json:"uuid"`
}

type UpdateJobRequest struct {
	Input        *string              `json:"input,omitempty"`
	Detect       *DetectOptions       `json:"detect,omitempty"`
	VideoSegment *VideoSegmentOptions `json:"videoSegment,omitempty"`
	WorkflowId   *int                 `json:"workflowId,omitempty"`
	Query        *string              `json:"query,omitempty"`
}

func (req *UpdateJobRequest) UpdateModel(job *model.Job) {
	if req.Input != nil {
		job.Input = *req.Input
	}
	if req.WorkflowId != nil {
		job.WorkflowId = *req.WorkflowId
	}
	if req.Query != nil {
		job.Query = *req.Query
	}
	if req.Detect != nil {
		job.Detect = &model.DetectOptions{
			ModelName:     req.Detect.ModelName,
			Labels:        req.Detect.Labels,
			ConfThreshold: req.Detect.ConfThreshold,
			IoUThreshold:  req.Detect.IoUThreshold,
			Interval:      req.Detect.Interval,
		}
	}
	if req.VideoSegment != nil {
		job.VideoSegment = &model.VideoSegmentOptions{
			Interval: req.VideoSegment.Interval,
		}
	}
}

type ListJobsRequest struct {
	Start int `json:"start,omitempty" binding:"required,min=0"`
	Limit int `json:"limit,omitempty" binding:"required,min=1,max=50"`
}

type ListJobsResponse struct {
	Items []JobSpec `json:"items"`
	Total int64     `json:"total"`
}
