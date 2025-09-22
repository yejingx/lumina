package model

import "time"

type DetectionBox struct {
	X1         int     `json:"x1,omitempty"`
	Y1         int     `json:"y1,omitempty"`
	X2         int     `json:"x2,omitempty"`
	Y2         int     `json:"y2,omitempty"`
	Confidence float32 `json:"confidence,omitempty"`
	ClassId    int     `json:"classId,omitempty"`
	Label      string  `json:"label,omitempty"`
}

type WorkflowResp struct {
	Answer string `json:"answer,omitempty" gorm:"type:text"`
}

type Message struct {
	Id           int             `json:"id" gorm:"primaryKey"`
	JobId        int             `json:"jobId" gorm:"type:int;index"`
	Timestamp    time.Time       `json:"timestamp" gorm:"type:timestamp;index"`
	ImagePath    string          `json:"imagePath,omitempty" gorm:"type:varchar(255)"`
	DetectBoxes  []*DetectionBox `json:"detectBoxes,omitempty" gorm:"type:json"`
	VideoPath    string          `json:"videoPath,omitempty" gorm:"type:varchar(255)"`
	CreateTime   time.Time       `json:"createTime" gorm:"type:timestamp;autoCreateTime"`
	WorkflowResp *WorkflowResp   `json:"workflowResp,omitempty" gorm:"type:json"`
}

func AddMessage(m *Message) error {
	return DB.Create(m).Error
}

func DeleteMessage(id int) error {
	return DB.Delete(&Message{}, id).Error
}

func GetMessage(id int) (*Message, error) {
	var m *Message
	if err := DB.Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

func GetMessagesByJobId(jobId, start, limit int) ([]*Message, error) {
	var ms []*Message
	if err := DB.Where("job_id = ?", jobId).Offset(start).Limit(limit).Find(&ms).Error; err != nil {
		return nil, err
	}
	return ms, nil
}

func CountMessagesByJobId(jobId int) (int64, error) {
	var count int64
	if err := DB.Model(&Message{}).Where("job_id = ?", jobId).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
