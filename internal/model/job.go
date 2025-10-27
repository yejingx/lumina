package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

type JobStatus int

const (
	JobStatusStopped JobStatus = iota
	JobStatusRunning
)

func (j JobStatus) String() string {
	switch j {
	case JobStatusStopped:
		return "stopped"
	case JobStatusRunning:
		return "running"
	default:
		return "unknown"
	}
}

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

	TriggerCount    int `json:"trigger_count" gorm:"default:1"`
	TriggerInterval int `json:"trigger_interval" gorm:"default:30"`
}

// Value implements driver.Valuer interface for JSON serialization
func (d DetectOptions) Value() (driver.Value, error) {
	return json.Marshal(d)
}

// Scan implements sql.Scanner interface for JSON deserialization
func (d *DetectOptions) Scan(value any) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, d)
}

type VideoSegmentOptions struct {
	Interval int `json:"interval" gorm:"default:30"`
}

// Value implements driver.Valuer interface for JSON serialization
func (v VideoSegmentOptions) Value() (driver.Value, error) {
	return json.Marshal(v)
}

// Scan implements sql.Scanner interface for JSON deserialization
func (v *VideoSegmentOptions) Scan(value any) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, v)
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
