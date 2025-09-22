package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"lumina/internal/dao"
	"lumina/internal/model"
	"net/http"
	"strings"
	"time"
)

type Dify struct {
	ctx     context.Context
	httpCli *http.Client
}

func NewDify(ctx context.Context) *Dify {
	httpCli := &http.Client{}
	return &Dify{
		ctx:     ctx,
		httpCli: httpCli,
	}
}

func (v *Dify) ChatCompletion(wf *model.Workflow, imageURL string, message *dao.AgentMessage, query string) (string, error) {
	// 构建检测结果的文本描述
	var detectionText strings.Builder
	if message != nil && len(message.DetectBoxes) > 0 {
		detectionText.WriteString("检测到的目标:\n")
		for i, box := range message.DetectBoxes {
			detectionText.WriteString(fmt.Sprintf("%d. 标签: %s, 置信度: %.2f, 位置: (%d,%d) - (%d,%d)\n",
				i+1, box.Label, box.Confidence, box.X1, box.Y1, box.X2, box.Y2))
		}
	} else {
		detectionText.WriteString("未检测到目标")
	}

	requestBody := map[string]any{
		"inputs": map[string]any{
			"detection": detectionText.String(),
			"image": map[string]string{
				"type":            "image",
				"transfer_method": "remote_url",
				"url":             imageURL,
			},
		},
		"query":         query,
		"response_mode": "blocking",
		"user":          "lumina-consumer",
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	ctx, cancel := context.WithTimeout(v.ctx, time.Duration(wf.Timeout)*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", wf.Endpoint+"/chat-messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if wf.Key != "" {
		req.Header.Set("Authorization", "Bearer "+wf.Key)
	}

	resp, err := v.httpCli.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response map[string]any
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if answer, ok := response["answer"].(string); ok {
		return answer, nil
	}

	return "", fmt.Errorf("no answer found in response")
}
