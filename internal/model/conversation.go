package model

import (
	"errors"
	"time"

	"lumina/internal/agent"

	"gorm.io/gorm"
)

type Conversation struct {
	Id         int       `gorm:"primaryKey"`
	Uuid       string    `gorm:"unique"`
	Title      string    `gorm:"default:''"`
	CreateTime time.Time `gorm:"datetime;autoCreateTime"`
}

func (c *Conversation) GetChatMessages(start, limit int) ([]*ChatMessage, int64, error) {
	return GetChatMessagesByConversationId(c.Id, start, limit)
}

func CreateConversation(c *Conversation) error {
	return DB.Create(c).Error
}

func GetConversationByUuid(uuid string) (*Conversation, error) {
	var c Conversation
	err := DB.Where("uuid = ?", uuid).First(&c).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func ListConversations(start, limit int) ([]*Conversation, int64, error) {
	var conversations []*Conversation
	var total int64
	err := DB.Model(&Conversation{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = DB.Model(&Conversation{}).Order("id desc").Offset(start).Limit(limit).Find(&conversations).Error
	if err != nil {
		return nil, 0, err
	}
	return conversations, total, nil
}

func DeleteConversationByUuid(uuid string) error {
	return DB.Where("uuid = ?", uuid).Delete(&Conversation{}).Error
}

type ChatMessage struct {
    Id             int                     `gorm:"primaryKey"`
    ConversationId int                     `gorm:"index"`
    Query          string                  `gorm:"default:''"`
    Answer         string                  `gorm:"type:longtext"`
    AgentThoughts  agent.AgentThoughtSlice `gorm:"type:json"`
    CreateTime     time.Time               `gorm:"datetime;autoCreateTime"`
}

func CreateChatMessage(m *ChatMessage) error {
	return DB.Create(m).Error
}

func DeleteChatMessage(id int) error {
	return DB.Where("id = ?", id).Delete(&ChatMessage{}).Error
}

func GetChatMessage(id int) (*ChatMessage, error) {
	var m ChatMessage
	err := DB.Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func GetChatMessagesByConversationId(id int, start, limit int) ([]*ChatMessage, int64, error) {
	var ms []*ChatMessage
	var total int64
	err := DB.Model(&ChatMessage{}).Where("conversation_id = ?", id).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = DB.Where("conversation_id = ?", id).Order("id desc").Offset(start).Limit(limit).Find(&ms).Error
	if err != nil {
		return nil, 0, err
	}
	return ms, total, nil
}
