package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
	Role       LLMMessageRole `json:"role"`
	Content    string         `json:"content,omitempty"`
	ToolCalls  []ToolCall     `json:"tool_calls,omitempty"`
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
	Stream      bool            `json:"stream,omitempty"`
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

// OpenAI streaming response structures
type OpenAIStreamResponse struct {
	ID      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Model   string               `json:"model"`
	Choices []OpenAIStreamChoice `json:"choices"`
}

type OpenAIStreamChoice struct {
	Index        int               `json:"index"`
	Delta        OpenAIStreamDelta `json:"delta"`
	FinishReason *string           `json:"finish_reason"`
}

type OpenAIStreamDelta struct {
	Role      string           `json:"role,omitempty"`
	Content   string           `json:"content,omitempty"`
	ToolCalls []OpenAIToolCall `json:"tool_calls,omitempty"`
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

// ChatCompletionStream performs streaming chat completion and returns the complete response
func (llm *LLM) ChatCompletionStream(ctx context.Context, messages []*LLMMessage, tools []*Tool, writer io.Writer) (*LLMMessage, error) {
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

	// Create OpenAI request with streaming enabled
	request := OpenAIRequest{
		Model:       llm.conf.Model,
		Messages:    openAIMessages,
		Tools:       openAITools,
		Temperature: llm.conf.Temperature,
		Stream:      true,
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
	req.Header.Set("Accept", "text/event-stream")

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

	// Variables to accumulate the complete response
	var completeContent strings.Builder
	var toolCalls []ToolCall
	var currentRole string
	var currentCallId string

	// Process streaming response
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		data, ok := strings.CutPrefix(line, "data: ")
		if !ok {
			continue
		} else if data == "[DONE]" {
			break
		}

		// Parse JSON response
		var streamResp OpenAIStreamResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			continue // Skip malformed JSON
		}

		if len(streamResp.Choices) == 0 {
			continue
		}

		delta := streamResp.Choices[0].Delta

		if delta.Role != "" {
			currentRole = delta.Role
		}

		// Handle content
		if delta.Content != "" {
			completeContent.WriteString(delta.Content)
			// Write content to output stream
			_, err := writer.Write([]byte(delta.Content))
			if err != nil {
				return nil, fmt.Errorf("failed to write content: %w", err)
			}
		}

		// Handle tool calls
		if len(delta.ToolCalls) > 0 {
			for _, openAIToolCall := range delta.ToolCalls {
				if openAIToolCall.ID != "" {
					currentCallId = openAIToolCall.ID
				}
				// Find or create tool call entry
				var targetToolCall *ToolCall
				for i := range toolCalls {
					if toolCalls[i].Id == currentCallId {
						targetToolCall = &toolCalls[i]
						break
					}
				}

				if targetToolCall == nil {
					toolCall := ToolCall{
						Id:       currentCallId,
						ToolName: openAIToolCall.Function.Name,
						Args:     openAIToolCall.Function.Arguments,
					}
					toolCalls = append(toolCalls, toolCall)
				} else {
					if openAIToolCall.Function.Name != "" {
						targetToolCall.ToolName = openAIToolCall.Function.Name
					}
					if openAIToolCall.Function.Arguments != "" {
						targetToolCall.Args += openAIToolCall.Function.Arguments
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading stream: %w", err)
	}

	// Return the complete message
	return &LLMMessage{
		Role:      LLMMessageRole(currentRole),
		Content:   completeContent.String(),
		ToolCalls: toolCalls,
	}, nil
}
