package dao

import (
	"errors"
	"lumina/internal/model"
	"time"
)

type ConversationSpec struct {
	Id         int    `json:"id"`
	Uuid       string `json:"uuid" binding:"required"`
	Title      string `json:"title"`
	CreateTime string `json:"createTime" binding:"required,datetime=2006-01-02T15:04:05Z07:00"`
}

func (c *ConversationSpec) ToModel() *model.Conversation {
	if c == nil {
		return nil
	}
	return &model.Conversation{
		Id:    c.Id,
		Uuid:  c.Uuid,
		Title: c.Title,
	}
}

func FromConversationModel(c *model.Conversation) (*ConversationSpec, error) {
	if c == nil {
		return nil, errors.New("conversation is nil")
	}
	return &ConversationSpec{
		Id:         c.Id,
		Uuid:       c.Uuid,
		Title:      c.Title,
		CreateTime: c.CreateTime.Format(time.RFC3339),
	}, nil
}

type CreateConversationRequest struct {
	Title string `json:"title"`
}

type CreateConversationResponse struct {
	Uuid string `json:"uuid" binding:"required"`
}

type ListConversationsRequest struct {
	Start int `json:"start" form:"start" binding:"min=0"`
	Limit int `json:"limit" form:"limit" binding:"min=0,max=50"`
}

type ListConversationsResponse struct {
	Items []*ConversationSpec `json:"items"`
	Total int64               `json:"total"`
}

type ToolCall struct {
	Name string `json:"name"`
	Args string `json:"args"`
}

func (t *ToolCall) ToModel() *model.ToolCall {
	if t == nil {
		return nil
	}
	return &model.ToolCall{
		Name: t.Name,
		Args: t.Args,
	}
}

func FromToolCallModel(t *model.ToolCall) *ToolCall {
	if t == nil {
		return nil
	}
	return &ToolCall{
		Name: t.Name,
		Args: t.Args,
	}
}

type LLMMessageSpec struct {
	Id             int       `json:"id"`
	ConversationId int       `json:"conversationId"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	ToolCall       *ToolCall `json:"toolCall,omitempty"`
	CreateTime     string    `json:"createTime"`
}

func (m *LLMMessageSpec) ToModel() *model.LLMMessage {
	if m == nil {
		return nil
	}
	return &model.LLMMessage{
		Id:             m.Id,
		ConversationId: m.ConversationId,
		Role:           m.Role,
		Content:        m.Content,
		ToolCall:       m.ToolCall.ToModel(),
	}
}

func FromLLMMessageModel(m *model.LLMMessage) *LLMMessageSpec {
	if m == nil {
		return nil
	}
	return &LLMMessageSpec{
		Id:             m.Id,
		ConversationId: m.ConversationId,
		Role:           m.Role,
		Content:        m.Content,
		ToolCall:       FromToolCallModel(m.ToolCall),
		CreateTime:     m.CreateTime.Format(time.RFC3339),
	}
}

type ListLLMMessagesRequest struct {
	Start int `json:"start" form:"start" binding:"min=0"`
	Limit int `json:"limit" form:"limit" binding:"min=0,max=50"`
}

// ListLLMMessagesResponse represents response for listing LLM messages
type ListLLMMessagesResponse struct {
	Items []*LLMMessageSpec `json:"items"`
	Total int64             `json:"total"`
}

type ChatMessage struct {
	Query string `json:"query" binding:"required"`
}
