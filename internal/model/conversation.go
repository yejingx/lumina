package model

import "time"

type Conversation struct {
	Id         int       `gorm:"primaryKey"`
	Uuid       string    `gorm:"unique"`
	Title      string    `gorm:"default:''"`
	CreateTime time.Time `gorm:"datetime;autoCreateTime"`
}

func (c *Conversation) GetLLMMessages() ([]*LLMMessage, error) {
	return GetLLMMessagesByConversationId(c.Id)
}

func CreateConversation(c *Conversation) error {
	return DB.Create(c).Error
}

func GetConversationByUuid(uuid string) (*Conversation, error) {
	var c Conversation
	err := DB.Where("uuid = ?", uuid).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func DeleteConversationByUuid(uuid string) error {
	return DB.Where("uuid = ?", uuid).Delete(&Conversation{}).Error
}

type ToolCall struct {
	Name string `json:"name" gorm:"default:''"`
	Args string `json:"args" gorm:"default:''"`
}

type LLMMessage struct {
	Id             int       `gorm:"primaryKey"`
	ConversationId int       `gorm:"index"`
	Role           string    `gorm:"default:''"`
	Content        string    `gorm:"default:''"`
	ToolCall       *ToolCall `gorm:"type:json"`
	CreateTime     time.Time `gorm:"datetime;autoCreateTime"`
}

func CreateLLMMessage(m *LLMMessage) error {
	return DB.Create(m).Error
}

func DeleteLLMMessage(id int) error {
	return DB.Where("id = ?", id).Delete(&LLMMessage{}).Error
}

func GetLLMMessage(id int) (*LLMMessage, error) {
	var m LLMMessage
	err := DB.Where("id = ?", id).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func GetLLMMessagesByConversationId(id int) ([]*LLMMessage, error) {
	var ms []*LLMMessage
	err := DB.Where("conversation_id = ?", id).Find(&ms).Error
	if err != nil {
		return nil, err
	}
	return ms, nil
}
