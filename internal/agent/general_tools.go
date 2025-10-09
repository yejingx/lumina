package agent

import "time"

var (
	getCurrentTimeTool = NewTool(
		WithToolName("get_current_time"),
		WithToolDescription("Get current time"),
		WithToolParamsSchema[GetCurrentTimeRequest](),
		WithToolFunc(GetCurrentTime),
	)
)

type GetCurrentTimeRequest struct {
	Format string `json:"format,omitempty" jsonschema_description:"Time format, default is %Y-%m-%dT%H:%M:%S%z"`
}

type GetCurrentTimeResponse struct {
	BaseToolResult
	Time string `json:"time" jsonschema_description:"Current time, RFC3339 format"`
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
