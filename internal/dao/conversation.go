package dao

import (
	"errors"
	"time"

	"lumina/internal/model"
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

type ToolCallSpec struct {
	Name string `json:"name"`
	Args string `json:"args"`
}

type AgentThoughtSpec struct {
	Thought     string        `json:"thought,omitempty"`
	Observation string        `json:"observation,omitempty"`
	ToolCall    *ToolCallSpec `json:"toolCall,omitempty"`
}

type ChatMessageSpec struct {
	Id             int                 `json:"id"`
	ConversationId int                 `json:"conversationId"`
	Query          string              `json:"query"`
	Answer         string              `json:"answer,omitempty"`
	AgentThoughts  []*AgentThoughtSpec `json:"agentThoughts,omitempty"`
	CreateTime     string              `json:"createTime"`
}

func (m *ChatMessageSpec) ToModel() *model.ChatMessage {
	if m == nil {
		return nil
	}
	return &model.ChatMessage{
		Id:             m.Id,
		ConversationId: m.ConversationId,
		Query:          m.Query,
		Answer:         m.Answer,
	}
}

func FromChatMessageModel(m *model.ChatMessage) *ChatMessageSpec {
	if m == nil {
		return nil
	}
	spec := &ChatMessageSpec{
		Id:             m.Id,
		ConversationId: m.ConversationId,
		Query:          m.Query,
		Answer:         m.Answer,
		CreateTime:     m.CreateTime.Format(time.RFC3339),
	}
	if len(m.AgentThoughts) > 0 {
		spec.AgentThoughts = make([]*AgentThoughtSpec, 0, len(m.AgentThoughts))
		for _, t := range m.AgentThoughts {
			spec.AgentThoughts = append(spec.AgentThoughts, &AgentThoughtSpec{
				Thought:     t.Thought,
				Observation: t.Observation,
				ToolCall: &ToolCallSpec{
					Name: t.ToolCall.ToolName,
					Args: t.ToolCall.Args,
				},
			})
		}
	}
	return spec
}

type ListChatMessagesRequest struct {
	Start int `json:"start" form:"start" binding:"min=0"`
	Limit int `json:"limit" form:"limit" binding:"min=0,max=50"`
}

// ListChatMessagesResponse represents response for listing chat messages
type ListChatMessagesResponse struct {
	Items []*ChatMessageSpec `json:"items"`
	Total int64              `json:"total"`
}

type ChatRequest struct {
	Query string `json:"query" binding:"required"`
}
