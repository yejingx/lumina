package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Workflow struct {
	Id           int              `gorm:"primaryKey"`
	Uuid         string           `gorm:"type:char(96);unique"`
	Key          string           `gorm:"type:varchar(255)"`
	ModelName    string           `gorm:"type:varchar(255)"`
	Endpoint     string           `gorm:"type:varchar(255)"`
	Name         string           `gorm:"type:varchar(255)"`
	Timeout      int              `gorm:"type:integer;default:30"`
	CreateTime   time.Time        `gorm:"type:timestamp;autoCreateTime"`
	Query        string           `json:"query" gorm:"type:text"`
	ResultFilter *FilterCondition `json:"result_filter" gorm:"type:json"`
}

func CreateWorkflow(wf *Workflow) error {
	return DB.Create(wf).Error
}

func GetWorkflowById(id int) (*Workflow, error) {
	var wf Workflow
	err := DB.First(&wf, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &wf, err
}

func GetWorkflowByUuid(uuid string) (*Workflow, error) {
	var wf Workflow
	err := DB.Where("uuid = ?", uuid).First(&wf).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &wf, err
}

func DeleteWorkflow(id int) error {
	return DB.Delete(&Workflow{}, id).Error
}

func ListWorkflows(start, limit int) ([]Workflow, int64, error) {
	var workflows []Workflow
	var total int64
	if err := DB.Model(&Workflow{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := DB.Model(&Workflow{}).Offset(start).Limit(limit).Find(&workflows).Error; err != nil {
		return nil, 0, err
	}
	return workflows, total, nil
}

func UpdateWorkflow(wf *Workflow) error {
	return DB.Save(wf).Error
}

type Operator string

const (
	OperatorEqual       = "eq"
	OperatorNotEqual    = "ne"
	OperatorIn          = "in"
	OperatorNotIn       = "not_in"
	OperatorContains    = "contains"
	OperatorNotContains = "not_contains"
	OperatorStartsWith  = "starts_with"
	OperatorEndsWith    = "ends_with"
	OperatorEmpty       = "empty"
	OperatorNotEmpty    = "not_empty"
)

type CombineOperator string

const (
	CombineOperatorAnd = "and"
	CombineOperatorOr  = "or"
)

type Condition struct {
	Field string   `json:"field,omitempty"`
	Op    Operator `json:"op,omitempty"`
	Value string   `json:"value,omitempty"`
}

func (c Condition) Match(s string) bool {
	switch c.Op {
	case OperatorEqual:
		return c.Value == s
	case OperatorNotEqual:
		return c.Value != s
	case OperatorIn:
		return strings.Contains(c.Value, s)
	case OperatorNotIn:
		return !strings.Contains(c.Value, s)
	case OperatorContains:
		return strings.Contains(s, c.Value)
	case OperatorNotContains:
		return !strings.Contains(s, c.Value)
	case OperatorStartsWith:
		return strings.HasPrefix(s, c.Value)
	case OperatorEndsWith:
		return strings.HasSuffix(s, c.Value)
	case OperatorEmpty:
		return s == ""
	case OperatorNotEmpty:
		return s != ""
	default:
		return false
	}
}

type FilterCondition struct {
	CombineOperator CombineOperator `json:"combine_op,omitempty"`
	Conditions      []*Condition    `json:"conditions,omitempty"`
}

func (f *FilterCondition) Value() (driver.Value, error) {
	return json.Marshal(f)
}

func (f *FilterCondition) Scan(value any) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, f)
}

func (f FilterCondition) Match(s string) bool {
	switch f.CombineOperator {
	case CombineOperatorAnd:
		for _, cond := range f.Conditions {
			if !cond.Match(s) {
				return false
			}
		}
		return true
	case CombineOperatorOr:
		for _, cond := range f.Conditions {
			if cond.Match(s) {
				return true
			}
		}
		return false
	}
	return false
}
