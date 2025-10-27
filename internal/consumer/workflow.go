package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"lumina/internal/dao"
	"lumina/internal/model"
)

type WorkflowManager struct {
	ctx     context.Context
	httpCli *http.Client
}

func NewWorkflowManager(ctx context.Context) *WorkflowManager {
	httpCli := &http.Client{}
	return &WorkflowManager{
		ctx:     ctx,
		httpCli: httpCli,
	}
}

type OpenAIRequest struct {
	Model       string                 `json:"model"`
	Messages    []OpenAIRequestMessage `json:"messages"`
	Temperature float64                `json:"temperature,omitempty"`
}

type OpenAIRequestMessage struct {
	Role    string                        `json:"role"`
	Content []OpenAIRequestMessageContent `json:"content"`
}

type OpenAIRequestMessageContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageUrl struct {
		Url string `json:"url"`
	} `json:"image_url,omitempty"`
	VideoUrl struct {
		Url string `json:"url"`
	} `json:"video_url,omitempty"`
}

const systemPrompt = `You are a helpful assistant that can analyze images and videos.`

func (v *WorkflowManager) ImageCompletion(wf *model.Workflow, imageURL string, boxes []*dao.DetectionBox) (*OpenAIResponse, error) {
	prompt := systemPrompt
	if len(boxes) > 0 {
		prompt += "\n\nDetected Objects in the image:\n"
		for i, box := range boxes {
			prompt += fmt.Sprintf("%d. label: %s, confidence: %.2f, position: (%d,%d) - (%d,%d)\n",
				i+1, box.Label, box.Confidence, box.X1, box.Y1, box.X2, box.Y2)
		}
		prompt += "\n\n"
	}

	req := OpenAIRequest{
		Model: wf.ModelName,
		Messages: []OpenAIRequestMessage{
			{Role: "system", Content: []OpenAIRequestMessageContent{{Type: "text", Text: prompt}}},
			{Role: "user", Content: []OpenAIRequestMessageContent{
				{Type: "text", Text: wf.Query},
				{Type: "image_url", ImageUrl: struct {
					Url string `json:"url"`
				}{Url: imageURL}},
			}},
		},
		Temperature: 0.3,
	}

	return v.chatCompletion(wf, req)
}

func (v *WorkflowManager) VideoCompletion(wf *model.Workflow, videoURL string) (*OpenAIResponse, error) {
	req := OpenAIRequest{
		Model: wf.ModelName,
		Messages: []OpenAIRequestMessage{
			{Role: "system", Content: []OpenAIRequestMessageContent{{Type: "text", Text: systemPrompt}}},
			{Role: "user", Content: []OpenAIRequestMessageContent{
				{Type: "text", Text: wf.Query},
				{Type: "video_url", VideoUrl: struct {
					Url string `json:"url"`
				}{Url: videoURL}},
			}},
		},
		Temperature: 0.3,
	}

	return v.chatCompletion(wf, req)
}

type OpenAIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   OpenAIUsage    `json:"usage"`
}

type OpenAIResponseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIChoice struct {
	Index        int                   `json:"index"`
	Message      OpenAIResponseMessage `json:"message"`
	FinishReason string                `json:"finish_reason"`
}

type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func (v *WorkflowManager) chatCompletion(wf *model.Workflow, r OpenAIRequest) (*OpenAIResponse, error) {
	jsonData, _ := json.Marshal(r)

	ctx, cancel := context.WithTimeout(v.ctx, time.Duration(wf.Timeout)*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", wf.Endpoint+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if wf.Key != "" {
		req.Header.Set("Authorization", "Bearer "+wf.Key)
	}

	resp, err := v.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response OpenAIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Choices) > 0 {
		return &response, nil
	}

	return nil, fmt.Errorf("no choices found in response")
}
