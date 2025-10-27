package dao

import (
	"time"

	"lumina/internal/model"
	"lumina/pkg/str"
)

type WorkflowSpec struct {
	Id           int              `json:"id"`
	Uuid         string           `json:"uuid" binding:"required"`
	Key          string           `json:"key" binding:"required"`
	Endpoint     string           `json:"endpoint" binding:"required"`
	ModelName    string           `json:"modelName" binding:"required"`
	Name         string           `json:"name" binding:"required"`
	Timeout      int              `json:"timeout"`
	CreateTime   string           `json:"createTime" binding:"datetime=2006-01-02T15:04:05Z07:00"`
	Query        string           `json:"query" binding:"required"`
	ResultFilter *FilterCondition `json:"resultFilter,omitempty"`
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
	w.ModelName = m.ModelName
	w.Name = m.Name
	w.Timeout = m.Timeout
	w.CreateTime = m.CreateTime.Format(time.RFC3339)
	w.Query = m.Query
	w.ResultFilter = FromFilterConditionModel(m.ResultFilter)
	return w
}

type CreateWorkflowRequest struct {
	Key          string           `json:"key" binding:"required"`
	Endpoint     string           `json:"endpoint" binding:"required"`
	ModelName    string           `json:"modelName" binding:"required"`
	Name         string           `json:"name" binding:"required"`
	Timeout      int              `json:"timeout" binding:"min=1"`
	Query        string           `json:"query,omitempty"`
	ResultFilter *FilterCondition `json:"resultFilter,omitempty"`
}

func (req *CreateWorkflowRequest) ToModel() *model.Workflow {
	if req.Timeout == 0 {
		req.Timeout = 30
	}

	workflow := &model.Workflow{
		Uuid:         str.GenDeviceId(16),
		Key:          req.Key,
		Endpoint:     req.Endpoint,
		ModelName:    req.ModelName,
		Name:         req.Name,
		Timeout:      req.Timeout,
		Query:        req.Query,
		ResultFilter: req.ResultFilter.ToModel(),
	}

	return workflow
}

type CreateWorkflowResponse struct {
	Id   int    `json:"id"`
	Uuid string `json:"uuid"`
}

type UpdateWorkflowRequest struct {
	Key          *string          `json:"key"`
	Endpoint     *string          `json:"endpoint"`
	ModelName    *string          `json:"modelName"`
	Name         *string          `json:"name"`
	Timeout      *int             `json:"timeout"`
	Query        *string          `json:"query,omitempty"`
	ResultFilter *FilterCondition `json:"resultFilter,omitempty"`
}

func (req *UpdateWorkflowRequest) UpdateModel(w *model.Workflow) {
	if req.Key != nil {
		w.Key = *req.Key
	}
	if req.Endpoint != nil {
		w.Endpoint = *req.Endpoint
	}
	if req.ModelName != nil {
		w.ModelName = *req.ModelName
	}
	if req.Name != nil {
		w.Name = *req.Name
	}
	if req.Timeout != nil {
		w.Timeout = *req.Timeout
	}
	if req.Query != nil {
		w.Query = *req.Query
	}
	if req.ResultFilter != nil {
		w.ResultFilter = req.ResultFilter.ToModel()
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

type Condition struct {
	Field string         `json:"field,omitempty"`
	Op    model.Operator `json:"op,omitempty"`
	Value string         `json:"value,omitempty"`
}

func (c *Condition) ToModel() *model.Condition {
	if c == nil {
		return nil
	}
	return &model.Condition{
		Field: c.Field,
		Op:    c.Op,
		Value: c.Value,
	}
}

type FilterCondition struct {
	CombineOperator model.CombineOperator `json:"combineOp,omitempty"`
	Conditions      []*Condition          `json:"conditions,omitempty"`
}

func (c *FilterCondition) ToModel() *model.FilterCondition {
	if c == nil {
		return nil
	}
	fc := &model.FilterCondition{
		CombineOperator: c.CombineOperator,
		Conditions:      make([]*model.Condition, 0, len(c.Conditions)),
	}
	for _, cond := range c.Conditions {
		fc.Conditions = append(fc.Conditions, cond.ToModel())
	}
	return fc
}

func FromFilterConditionModel(filter *model.FilterCondition) *FilterCondition {
	if filter == nil {
		return nil
	}
	fc := &FilterCondition{
		CombineOperator: filter.CombineOperator,
		Conditions:      make([]*Condition, 0, len(filter.Conditions)),
	}
	for _, cond := range filter.Conditions {
		fc.Conditions = append(fc.Conditions, FromConditionModel(cond))
	}
	return fc
}

func FromConditionModel(condition *model.Condition) *Condition {
	if condition == nil {
		return nil
	}
	return &Condition{
		Field: condition.Field,
		Op:    condition.Op,
		Value: condition.Value,
	}
}
