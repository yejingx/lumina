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
	CreateTime time.Time `gorm:"type:timestamp;autoCreateTime"`
}

func CreateWorkflow(wf *Workflow) error {
	return DB.Create(wf).Error
}

func GetWorkflow(id int) (*Workflow, error) {
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