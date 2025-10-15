package agent

import (
	"context"
	"fmt"
	"io"
	"lumina/internal/utils"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	getCurrentTimeTool = NewTool(
		WithToolName("get_current_time"),
		WithToolDescription("Get current time"),
		WithToolParamsSchema[GetCurrentTimeRequest](),
		WithToolFunc(GetCurrentTime),
	)
	httpFetchTool = NewTool(
		WithToolName("httpFetch"),
		WithToolDescription("Fetches content from a URL via HTTP GET request"),
		WithToolParamsSchema[HttpToolParams](),
		WithToolFunc(handleHttpRequest),
	)
)

type GetCurrentTimeRequest struct {
	Format string `json:"format,omitempty" jsonschema:"description=Time format, default is %Y-%m-%dT%H:%M:%S%z"`
}

type GetCurrentTimeResponse struct {
	BaseToolResult
	Time string `json:"time" jsonschema:"description=Current time, RFC3339 format"`
}

// ConvertToGoLayout 将常见时间格式字符串（如 "YYYY-MM-DD HH:mm:ss"）
// 转换为 Go 的时间格式模板（如 "2006-01-02 15:04:05"）
func convertToGoLayout(format string) string {
	replacements := map[string]string{
		"YYYY": "2006",
		"YY":   "06",
		"MM":   "01",
		"DD":   "02",
		"HH":   "15", // 24小时制
		"hh":   "03", // 12小时制
		"mm":   "04",
		"ss":   "05",
		"SSS":  ".000", // 毫秒
		"AM":   "PM",
		"am":   "pm",
		"Z":    "-0700", // 时区
	}

	// 按长度从大到小替换，避免部分匹配被提前替掉
	keys := []string{"YYYY", "YY", "MM", "DD", "HH", "hh", "mm", "ss", "SSS", "AM", "am", "Z"}
	for _, k := range keys {
		format = strings.ReplaceAll(format, k, replacements[k])
	}
	return format
}

func GetCurrentTime(id string, req *GetCurrentTimeRequest) (*GetCurrentTimeResponse, error) {
	format := req.Format
	if format == "" {
		format = "2006-01-02 15:04:05"
	}
	return &GetCurrentTimeResponse{
		BaseToolResult: BaseToolResult{Id: id},
		Time:           time.Now().Format(convertToGoLayout(format)),
	}, nil
}

type HttpToolParams struct {
	URL string `json:"url" jsonschema:"description=URL to fetch"`
}

type HttpToolResult struct {
	BaseToolResult
	StatusCode int               `json:"status_code" jsonschema:"description=HTTP status code"`
	Body       string            `json:"body"        jsonschema:"description=Response body"`
	Headers    map[string]string `json:"headers"     jsonschema:"description=Response headers"`
}

func handleHttpRequest(callID string, params HttpToolParams) (*HttpToolResult, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, params.URL, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logrus.Errorf("Error closing response body: %v", closeErr)
		}
	}()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	headers := make(map[string]string)
	for name, values := range resp.Header {
		if len(values) > 0 {
			headers[name] = values[0]
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	mainContent := utils.ExtractMainContent(body)

	result := &HttpToolResult{
		BaseToolResult: BaseToolResult{Id: callID},
		StatusCode:     resp.StatusCode,
		Body:           mainContent,
		Headers:        headers,
	}

	return result, nil
}
