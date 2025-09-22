package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Workflow struct {
	Id         int       `gorm:"primaryKey"`
	Uuid       string    `gorm:"type:char(96);unique"`
	Key        string    `gorm:"type:varchar(255)"`
	Endpoint   string    `gorm:"type:varchar(255)"`
	Name       string    `gorm:"type:varchar(255)"`
	Timeout    int       `gorm:"type:integer;default:30"`
	CreateTime time.Time `gorm:"type:timestamp;autoCreateTime"`
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
