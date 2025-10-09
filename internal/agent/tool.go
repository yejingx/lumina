package agent

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/invopop/jsonschema"
)

var ErrCannotCreateSchema = errors.New("cannot create schema from output type")

var reflector = jsonschema.Reflector{
	AllowAdditionalProperties: false,
	DoNotReference:            true,
}

type ToolResult interface {
	GetId() string
}

type BaseToolResult struct {
	Id string `json:"id"`
}

func (b BaseToolResult) GetId() string {
	return b.Id
}

type Tool struct {
	Name             string                                           `json:"name"`
	Description      string                                           `json:"description"`
	ParametersSchema any                                              `json:"-"`
	Func             func(id string, args string) (ToolResult, error) `json:"-"`
}

// MarshalJSON implements custom JSON marshaling for Tool
func (t *Tool) MarshalJSON() ([]byte, error) {
	schema, err := t.GetParametersSchema()
	if err != nil {
		// Fallback to empty object if schema generation fails
		schema = map[string]any{}
	}
	if schema["properties"] != nil {
		schema = schema["properties"].(map[string]any)
	}
	
	return json.Marshal(struct {
		Name             string         `json:"name"`
		Description      string         `json:"description"`
		ParametersSchema map[string]any `json:"parameters_schema"`
	}{
		Name:             t.Name,
		Description:      t.Description,
		ParametersSchema: schema,
	})
}

type ToolOption func(*Tool)

func NewTool(options ...ToolOption) *Tool {
	t := &Tool{}
	for _, option := range options {
		option(t)
	}
	return t
}

func WithToolName(name string) ToolOption {
	return func(t *Tool) {
		t.Name = name
	}
}

func WithToolDescription(description string) ToolOption {
	return func(t *Tool) {
		t.Description = description
	}
}

func WithToolParamsSchema[T any]() ToolOption {
	return func(t *Tool) {
		t.ParametersSchema = new(T)
	}
}

func WithToolFunc[T any, P ToolResult](f func(id string, args T) (P, error)) ToolOption {
	return func(t *Tool) {
		t.Func = func(id string, args string) (ToolResult, error) {
			var params T
			if err := json.Unmarshal([]byte(args), &params); err != nil {
				return nil, err
			}
			return f(id, params)
		}
	}
}

func (t *Tool) GetParametersSchema() (map[string]any, error) {
	schema, err := json.Marshal(reflector.Reflect(t.ParametersSchema))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCannotCreateSchema, err)
	}
	var result map[string]any
	if err := json.Unmarshal(schema, &result); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCannotCreateSchema, err)
	}
	return result, nil
}

type ToolCall struct {
	Id       string `json:"id"`
	ToolName string `json:"tool_name"`
	Args     string `json:"args"`
}

func NewToolCall(id, toolName, args string) *ToolCall {
	return &ToolCall{
		Id:       id,
		ToolName: toolName,
		Args:     args,
	}
}
