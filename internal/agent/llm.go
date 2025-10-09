package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type LLMConfig struct {
	BaseUrl     string
	ApiKey      string
	Model       string
	Temperature float64
	Timeout     time.Duration
}

type LLMMessageRole string

const (
	RoleSystem    LLMMessageRole = "system"
	RoleUser      LLMMessageRole = "user"
	RoleAssistant LLMMessageRole = "assistant"
	RoleTool      LLMMessageRole = "tool"
)

type LLMMessage struct {
	Role      LLMMessageRole `json:"role"`
	Content   string         `json:"content,omitempty"`
	ToolCalls []ToolCall     `json:"tool_calls,omitempty"`
	ToolCallId string         `json:"tool_call_id,omitempty"`
}

// OpenAI compatible request structures
type OpenAIToolCall struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Function OpenAIToolCallFunction `json:"function"`
}

type OpenAIToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type OpenAIMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content,omitempty"`
	ToolCalls []OpenAIToolCall `json:"tool_calls,omitempty"`
}

type OpenAITool struct {
	Type     string             `json:"type"`
	Function OpenAIToolFunction `json:"function"`
}

type OpenAIToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Tools       []OpenAITool    `json:"tools,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
}

// OpenAI compatible response structures
type OpenAIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   OpenAIUsage    `json:"usage"`
}

type OpenAIChoice struct {
	Index        int           `json:"index"`
	Message      OpenAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type LLM struct {
	httpCli *http.Client
	conf    LLMConfig
}

func NewLLM(conf LLMConfig) *LLM {
	return &LLM{
		httpCli: &http.Client{
			Timeout: conf.Timeout,
		},
		conf: conf,
	}
}

// func (llm *LLM) Completion(ctx context.Context, prompt string) (*LLMMessage, error) {

// }

func (llm *LLM) ChatCompletion(ctx context.Context, messages []*LLMMessage, tools []*Tool) (*LLMMessage, error) {
	// Convert messages to OpenAI format
	openAIMessages := make([]OpenAIMessage, len(messages))
	for i, msg := range messages {
		openAIMessages[i] = OpenAIMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
	}

	// Convert tools to OpenAI format
	var openAITools []OpenAITool
	if len(tools) > 0 {
		openAITools = make([]OpenAITool, len(tools))
		for i, tool := range tools {
			params, err := tool.GetParametersSchema()
			if err != nil {
				return nil, fmt.Errorf("failed to get parameters schema for tool %s: %w", tool.Name, err)
			}

			openAITools[i] = OpenAITool{
				Type: "function",
				Function: OpenAIToolFunction{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  params,
				},
			}
		}
	}

	// Create OpenAI request
	request := OpenAIRequest{
		Model:       llm.conf.Model,
		Messages:    openAIMessages,
		Tools:       openAITools,
		Temperature: llm.conf.Temperature,
	}

	// Marshal request to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := llm.conf.BaseUrl + "/chat/completions"
	ctx, cancel := context.WithTimeout(ctx, llm.conf.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+llm.conf.ApiKey)

	// Send HTTP request
	resp, err := llm.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check if response has choices
	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// Convert response back to LLMMessage
	choice := openAIResp.Choices[0]

	// Convert OpenAI tool calls to internal ToolCall format
	var toolCalls []ToolCall
	for _, openAIToolCall := range choice.Message.ToolCalls {
		toolCall := ToolCall{
			Id:       openAIToolCall.ID,
			ToolName: openAIToolCall.Function.Name,
			Args:     openAIToolCall.Function.Arguments,
		}
		toolCalls = append(toolCalls, toolCall)
	}

	return &LLMMessage{
		Role:      LLMMessageRole(choice.Message.Role),
		Content:   choice.Message.Content,
		ToolCalls: toolCalls,
	}, nil
}
