package agent

import (
	"context"
	"fmt"
	"io"
	"net/http"
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

func GetCurrentTime(id string, req *GetCurrentTimeRequest) (*GetCurrentTimeResponse, error) {
	format := req.Format
	if format == "" {
		format = time.RFC3339
	}
	return &GetCurrentTimeResponse{
		Time: time.Now().Format(format),
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

	result := &HttpToolResult{
		BaseToolResult: BaseToolResult{Id: callID},
		StatusCode:     resp.StatusCode,
		Body:           string(body),
		Headers:        headers,
	}

	return result, nil
}
