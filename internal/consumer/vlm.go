package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"lumina/internal/dao"
)

// OpenAI API 请求和响应结构体
type ChatMessage struct {
	Role    string        `json:"role"`
	Content []ContentPart `json:"content"`
}

type ContentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

type ImageURL struct {
	URL string `json:"url"`
}

type ChatCompletionRequest struct {
	Model     string        `json:"model"`
	Messages  []ChatMessage `json:"messages"`
	MaxTokens int           `json:"max_tokens"`
}

type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int         `json:"index"`
	Message      TextMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type TextMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type VLM struct {
	ctx     context.Context
	conf    VLMConfig
	httpCli *http.Client
}

func NewVLM(ctx context.Context, conf VLMConfig) *VLM {
	httpCli := &http.Client{
		Timeout: conf.Timeout,
	}
	return &VLM{
		ctx:     ctx,
		conf:    conf,
		httpCli: httpCli,
	}
}

func (v *VLM) ChatCompletion(imageURL string, message *dao.Message) (string, error) {
	prompt := buildPromptWithDetectionBoxes(message)

	model := v.conf.Model
	if model == "" {
		model = "gpt-4-vision-preview"
	}

	req := ChatCompletionRequest{
		Model: model,
		Messages: []ChatMessage{
			{
				Role: "user",
				Content: []ContentPart{
					{
						Type: "text",
						Text: prompt,
					},
					{
						Type: "image_url",
						ImageURL: &ImageURL{
							URL: imageURL,
						},
					},
				},
			},
		},
		MaxTokens: 2000,
	}

	reqBody, _ := json.Marshal(req)

	apiURL := v.conf.BaseURL
	if !strings.HasSuffix(apiURL, "/") {
		apiURL += "/"
	}
	apiURL += "chat/completions"

	ctx, cancel := context.WithTimeout(v.ctx, v.conf.Timeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if v.conf.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+v.conf.APIKey)
	}

	resp, err := v.httpCli.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call VLM API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("VLM API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from VLM API")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func buildPromptWithDetectionBoxes(message *dao.Message) string {
	var prompt strings.Builder

	prompt.WriteString("请分析这张图片的主要内容，提供一个简洁的总结，方便后续根据图像内容进行检索。")

	if len(message.DetectBoxes) > 0 {
		prompt.WriteString("\n\n图片中检测到以下目标，请重点关注这些区域：\n")
		for i, box := range message.DetectBoxes {
			prompt.WriteString(fmt.Sprintf("%d. %s (置信度: %.2f, 位置: x1=%d, y1=%d, x2=%d, y2=%d)\n",
				i+1, box.Label, box.Confidence, box.X1, box.Y1, box.X2, box.Y2))
		}
		prompt.WriteString("\n请结合这些检测到的目标，描述图片的整体场景和主要内容。")
	}

	prompt.WriteString("\n\n请用中文回答，保持简洁明了。")

	return prompt.String()
}
