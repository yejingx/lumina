package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type JobStatus int

const (
	JobStatusStopped JobStatus = iota
	JobStatusRunning
)

type JobKind string

const (
	JobKindDetect       JobKind = "detect"
	JobKindVideoSegment JobKind = "video_segment"
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
	DeviceId     int                  `json:"device_id" gorm:"index"`
	Uuid         string               `json:"uuid" gorm:"unique"`
	Kind         JobKind              `json:"kind" gorm:"default:0"`
	Input        string               `json:"input" gorm:"NOT NULL"`
	Status       JobStatus            `json:"status" gorm:"default:0"`
	CreateTime   time.Time            `json:"create_time" gorm:"datetime;autoCreateTime"`
	UpdateTime   time.Time            `json:"update_time" gorm:"datetime;autoCreateTime;autoUpdateTime"`
	Detect       *DetectOptions       `json:"detect" gorm:"type:json"`
	VideoSegment *VideoSegmentOptions `json:"video_segment" gorm:"type:json"`
	WorkflowId   int                  `json:"workflow_id" gorm:"default:0"`
	Query        string               `json:"query" gorm:"type:text"`
}

func (j *Job) Device() (*Device, error) {
	if j.DeviceId == 0 {
		return nil, nil
	}
	device, err := GetDeviceById(j.DeviceId)
	if err != nil {
		return nil, err
	}
	return device, nil
}

func (j *Job) Workflow() (*Workflow, error) {
	if j.WorkflowId == 0 {
		return nil, nil
	}
	wf, err := GetWorkflowById(j.WorkflowId)
	if err != nil {
		return nil, err
	}
	return wf, nil
}

func AddJob(job *Job) error {
	return DB.Create(job).Error
}

func DeleteJob(job *Job) error {
	return DB.Delete(job).Error
}

func GetJobByUuid(uuid string) (*Job, error) {
	var job Job
	if err := DB.Where("uuid = ?", uuid).First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func GetJobById(id int) (*Job, error) {
	var job Job
	if err := DB.Where("id = ?", id).First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func ListJobs(start, limit int) ([]Job, int64, error) {
	var jobs []Job
	var total int64
	if err := DB.Model(&Job{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := DB.Model(&Job{}).Offset(start).Limit(limit).Find(&jobs).Error; err != nil {
		return nil, 0, err
	}
	return jobs, total, nil
}

func ListJobsByDeviceId(deviceId int, start, limit int) ([]Job, int64, error) {
	var jobs []Job
	var total int64
	if err := DB.Model(&Job{}).Where("device_id = ?", deviceId).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := DB.Model(&Job{}).Where("device_id = ?", deviceId).Offset(start).Limit(limit).Find(&jobs).Error; err != nil {
		return nil, 0, err
	}
	return jobs, total, nil
}

func UpdateJob(job *Job) error {
	return DB.Save(job).Error
}
