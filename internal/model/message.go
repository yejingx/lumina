package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

type DetectionBox struct {
	X1         int     `json:"x1,omitempty"`
	Y1         int     `json:"y1,omitempty"`
	X2         int     `json:"x2,omitempty"`
	Y2         int     `json:"y2,omitempty"`
	Confidence float32 `json:"confidence,omitempty"`
	ClassId    int     `json:"classId,omitempty"`
	Label      string  `json:"label,omitempty"`
}

// Value implements driver.Valuer interface for JSON serialization
func (d DetectionBox) Value() (driver.Value, error) {
	return json.Marshal(d)
}

// Scan implements sql.Scanner interface for JSON deserialization
func (d *DetectionBox) Scan(value any) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, d)
}

// DetectionBoxSlice is a custom type for handling []*DetectionBox serialization
type DetectionBoxSlice []*DetectionBox

// Value implements driver.Valuer interface for JSON serialization
func (d DetectionBoxSlice) Value() (driver.Value, error) {
	if d == nil {
		return nil, nil
	}
	return json.Marshal(d)
}

// Scan implements sql.Scanner interface for JSON deserialization
func (d *DetectionBoxSlice) Scan(value any) error {
	if value == nil {
		*d = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, d)
}

type WorkflowResp struct {
	Answer string `json:"answer,omitempty" gorm:"type:text"`
}

func (w WorkflowResp) Value() (driver.Value, error) {
	return json.Marshal(w)
}

func (w *WorkflowResp) Scan(value any) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, w)
}

type Message struct {
	Id           int               `json:"id" gorm:"primaryKey"`
	JobId        int               `json:"jobId" gorm:"type:int;index"`
	Timestamp    time.Time         `json:"timestamp" gorm:"type:datetime;index"`
	ImagePath    string            `json:"imagePath,omitempty" gorm:"type:varchar(255)"`
	DetectBoxes  DetectionBoxSlice `json:"detectBoxes,omitempty" gorm:"type:json"`
	VideoPath    string            `json:"videoPath,omitempty" gorm:"type:varchar(255)"`
	CreateTime   time.Time         `json:"createTime" gorm:"type:datetime;autoCreateTime"`
	WorkflowResp *WorkflowResp     `json:"workflowResp,omitempty" gorm:"type:json"`
	Alerted      bool              `json:"alerted,omitempty" gorm:"type:bool;default:false"`
}

func AddMessage(m *Message) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(m).Error; err != nil {
			return err
		}
		if m.Alerted {
			alert := &AlertMessage{MessageId: m.Id}
			if err := tx.Create(alert).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func DeleteMessage(id int) error {
	return DB.Delete(&Message{}, id).Error
}

func GetMessage(id int) (*Message, error) {
	var m *Message
	if err := DB.Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m, nil
}

func GetMessagesByJobId(jobId, start, limit int) ([]*Message, int64, error) {
	var ms []*Message
	if err := DB.Where("job_id = ?", jobId).Order("id desc").Offset(start).Limit(limit).Find(&ms).Error; err != nil {
		return nil, 0, err
	}
	var count int64
	if err := DB.Model(&Message{}).Where("job_id = ?", jobId).Count(&count).Error; err != nil {
		return nil, 0, err
	}
	return ms, count, nil
}

func GetMessages(start, limit int) ([]*Message, int64, error) {
	var total int64
	var ms []*Message
	if err := DB.Order("id desc").Offset(start).Limit(limit).Find(&ms).Error; err != nil {
		return nil, 0, err
	}
	if err := DB.Model(&Message{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	return ms, total, nil
}

type AlertMessage struct {
	Id         int       `gorm:"primaryKey"`
	MessageId  int       `gorm:"type:int;index"`
	Message    Message   `gorm:"foreignKey:MessageId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CreateTime time.Time `gorm:"type:datetime;autoCreateTime"`
}

func GetAlertMessagesByJobId(jobId, start, limit int) ([]*Message, int64, error) {
	var alerts []*AlertMessage
	base := DB.Model(&AlertMessage{}).
		Joins("JOIN messages ON messages.id = alert_messages.message_id").
		Where("messages.job_id = ?", jobId)

	if err := base.Preload("Message").
		Order("alert_messages.id desc").
		Offset(start).
		Limit(limit).
		Find(&alerts).Error; err != nil {
		return nil, 0, err
	}

	var count int64
	if err := base.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	var ms []*Message
	for _, a := range alerts {
		ms = append(ms, &a.Message)
	}

	return ms, count, nil
}

func GetAlertMessages(start, limit int) ([]*Message, int64, error) {
	var total int64
	var alerts []*AlertMessage
	if err := DB.Preload("Message").Model(&AlertMessage{}).Order("id desc").Offset(start).Limit(limit).Find(&alerts).Error; err != nil {
		return nil, 0, err
	}
	if err := DB.Model(&AlertMessage{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var ms []*Message
	for _, a := range alerts {
		ms = append(ms, &a.Message)
	}

	return ms, total, nil
}
