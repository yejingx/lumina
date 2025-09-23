package dao

import (
	"time"

	"lumina/internal/model"
	"lumina/pkg/str"
)

type WorkflowSpec struct {
	Id         int    `json:"id"`
	Uuid       string `json:"uuid" binding:"required"`
	Key        string `json:"key" binding:"required"`
	Endpoint   string `json:"endpoint" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Timeout    int    `json:"timeout"`
	CreateTime string `json:"createTime" binding:"datetime=2006-01-02T15:04:05Z07:00"`
}

func FromWorkflowModel(m *model.Workflow) *WorkflowSpec {
	if m == nil {
		return nil
	}
	w := &WorkflowSpec{}
	w.Id = m.Id
	w.Uuid = m.Uuid
	w.Key = m.Key
	w.Endpoint = m.Endpoint
	w.Name = m.Name
	w.Timeout = m.Timeout
	w.CreateTime = m.CreateTime.Format(time.RFC3339)
	return w
}

type CreateWorkflowRequest struct {
	Key      string `json:"key" binding:"required"`
	Endpoint string `json:"endpoint" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Timeout  int    `json:"timeout" binding:"min=1"`
}

func (req *CreateWorkflowRequest) ToModel() *model.Workflow {
	if req.Timeout == 0 {
		req.Timeout = 30
	}

	workflow := &model.Workflow{
		Uuid:     str.GenDeviceId(16),
		Key:      req.Key,
		Endpoint: req.Endpoint,
		Name:     req.Name,
		Timeout:  req.Timeout,
	}

	return workflow
}

type CreateWorkflowResponse struct {
	Id   int    `json:"id"`
	Uuid string `json:"uuid"`
}

type UpdateWorkflowRequest struct {
	Key      *string `json:"key"`
	Endpoint *string `json:"endpoint"`
	Name     *string `json:"name"`
	Timeout  *int    `json:"timeout"`
}

func (req *UpdateWorkflowRequest) UpdateModel(w *model.Workflow) {
	if req.Key != nil {
		w.Key = *req.Key
	}
	if req.Endpoint != nil {
		w.Endpoint = *req.Endpoint
	}
	if req.Name != nil {
		w.Name = *req.Name
	}
	if req.Timeout != nil {
		w.Timeout = *req.Timeout
	}
}

type ListWorkflowRequest struct {
	Start int `json:"start" form:"start" binding:"min=0"`
	Limit int `json:"limit" form:"limit" binding:"min=0,max=100"`
}

type ListWorkflowResponse struct {
	Workflows []WorkflowSpec `json:"workflows"`
	Total     int64          `json:"total"`
}
