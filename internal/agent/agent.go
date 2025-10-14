package agent

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/sirupsen/logrus"
)

type ThoughtPhase string

const (
	ThoughtPhaseThought     ThoughtPhase = "thought"
	ThoughtPhaseTool        ThoughtPhase = "tool"
	ThoughtPhaseObservation ThoughtPhase = "observation"
)

type AgentThought struct {
	Phase       ThoughtPhase `json:"phase,omitempty"`
	ID          string       `json:"id,omitempty"`
	Thought     string       `json:"thought,omitempty"`
	Observation string       `json:"observation,omitempty"`
	ToolCall    *ToolCall    `json:"toolCall,omitempty"`
}

// Value implements driver.Valuer interface for JSON serialization
func (t AgentThought) Value() (driver.Value, error) {
	return json.Marshal(t)
}

// Scan implements sql.Scanner interface for JSON deserialization
func (t *AgentThought) Scan(value any) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, t)
}

// AgentThoughtSlice is a custom type for handling []*AgentThought serialization
type AgentThoughtSlice []*AgentThought

// Value implements driver.Valuer interface for JSON serialization of slice
func (s AgentThoughtSlice) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan implements sql.Scanner interface for JSON deserialization of slice
func (s *AgentThoughtSlice) Scan(value any) error {
	if value == nil {
		*s = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, s)
}

func mergeMessages(messages []*LLMMessage) string {
	tmp := make([]string, 0, len(messages))
	for _, msg := range messages {
		tmp = append(tmp, fmt.Sprintf("%s: %s", msg.Role, msg.Content))
	}
	return strings.Join(tmp, "\n")
}

type Agent struct {
	name          string
	llm           *LLM
	tools         map[string]*Tool
	maxIterations int
	instruction   string
	logger        *logrus.Entry
}

func NewAgent(name string, llmConf LLMConfig, maxIterations int, instruction string) *Agent {
	a := &Agent{
		name:          name,
		llm:           NewLLM(llmConf),
		tools:         make(map[string]*Tool),
		maxIterations: maxIterations,
		instruction:   instruction,
		logger:        logrus.WithField("agent", name),
	}
	a.AddTool(getCurrentTimeTool)
	a.AddTool(httpFetchTool)
	return a
}

func (a *Agent) AddTool(tool *Tool) {
	a.tools[tool.Name] = tool
}

func (a *Agent) Run(ctx context.Context, query string) (*LLMMessage, error) {
	messages := make([]*LLMMessage, 0)

	toolsDesc, _ := json.MarshalIndent(a.tools, "", "  ")
	tools := make([]*Tool, 0, len(a.tools))
	for _, tool := range a.tools {
		tools = append(tools, tool)
	}
	systemPrompt, err := RenderPrompt(map[string]any{
		"instruction": a.instruction,
		"tools":       string(toolsDesc),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to render prompt: %w", err)
	}
	messages = append(messages, &LLMMessage{
		Role:    RoleSystem,
		Content: systemPrompt,
	})

	a.logger.Debugf("system prompt:\n%s", systemPrompt)

	messages = append(messages, &LLMMessage{
		Role:    RoleUser,
		Content: query,
	})

	currentIteration := 0
	for {
		if currentIteration > a.maxIterations {
			messages = append(messages, &LLMMessage{
				Role:    RoleAssistant,
				Content: "I'm sorry, but I couldn't find a satisfactory answer within the allowed number of iterations. Here's what I know so far: " + mergeMessages(messages),
			})
			break
		}
		currentIteration += 1

		resp, err := a.llm.ChatCompletion(ctx, messages, tools)
		if err != nil {
			messages = append(messages, &LLMMessage{
				Role:    RoleAssistant,
				Content: "call llm failed: " + err.Error(),
			})
			break
		}

		a.logger.Debugf("llm response:\n%+v", resp)

		messages = append(messages, resp)
		if len(resp.ToolCalls) == 0 {
			break
		}
		toolCall := resp.ToolCalls[0]
		tool, ok := a.tools[toolCall.ToolName]
		if !ok {
			messages = append(messages, &LLMMessage{
				Role:    RoleTool,
				Content: fmt.Sprintf("Tool %s not found", toolCall.ToolName),
			})
			continue
		}
		toolRes, err := tool.Func(toolCall.Id, toolCall.Args)
		if err != nil {
			messages = append(messages, &LLMMessage{
				Role:    RoleAssistant,
				Content: fmt.Sprintf("Call tool %s failed: %s.", toolCall.ToolName, err.Error()),
			})
			continue
		}
		toolResStr, _ := json.MarshalIndent(toolRes, "", "  ")
		messages = append(messages, &LLMMessage{
			Role:       RoleTool,
			Content:    fmt.Sprintf("Tool %s returned: %s", toolCall.ToolName, string(toolResStr)),
			ToolCallId: toolCall.Id,
		})
	}

	return messages[len(messages)-1], nil
}

func (a *Agent) RunStream(ctx context.Context, query string, history []*LLMMessage, w io.Writer) ([]*AgentThought, error) {
	sseWriter := NewSSEMessageWriter(w)
	defer sseWriter.Close()

	messages := make([]*LLMMessage, 0)

	toolsDesc, _ := json.MarshalIndent(a.tools, "", "  ")
	tools := make([]*Tool, 0, len(a.tools))
	for _, tool := range a.tools {
		tools = append(tools, tool)
	}
	systemPrompt, err := RenderPrompt(map[string]any{
		"instruction": a.instruction,
		"tools":       string(toolsDesc),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to render prompt: %w", err)
	}
	messages = append(messages, &LLMMessage{
		Role:    RoleSystem,
		Content: systemPrompt,
	})

	a.logger.Debugf("system prompt:\n%s", systemPrompt)

	if len(history) > 0 {
		messages = append(messages, history...)
	}

	messages = append(messages, &LLMMessage{
		Role:    RoleUser,
		Content: query,
	})

	agentThoughts := make([]*AgentThought, 0)

	for i := 0; i < a.maxIterations; i++ {
		response, err := a.llm.ChatCompletionStream(ctx, messages, tools, sseWriter)
		if err != nil {
			return nil, fmt.Errorf("streaming chat completion failed: %w", err)
		}

		messages = append(messages, response)

		if len(response.ToolCalls) == 0 {
			agentThoughts = append(agentThoughts, &AgentThought{
				Phase:   ThoughtPhaseThought,
				ID:      response.ID,
				Thought: response.Content,
			})
			break
		}

		toolCall := response.ToolCalls[0]

		thought := &AgentThought{
			Phase:    ThoughtPhaseTool,
			ID:       response.ID,
			Thought:  response.Content,
			ToolCall: &toolCall,
		}

		err = sseWriter.Write(thought)
		if err != nil {
			return nil, fmt.Errorf("failed to write tool call: %w", err)
		}

		tool, exists := a.tools[toolCall.ToolName]
		if !exists {
			content := fmt.Sprintf("Error: tool %s not found", toolCall.ToolName)
			messages = append(messages, &LLMMessage{
				Role:       RoleTool,
				Content:    content,
				ToolCallId: toolCall.Id,
			})
			thought.Observation = content
			err = sseWriter.Write(thought)
			if err != nil {
				return nil, fmt.Errorf("failed to write tool call: %w", err)
			}
			agentThoughts = append(agentThoughts, thought)
			continue
		}

		a.logger.Debugf("executing tool %s with args: %s", toolCall.ToolName, toolCall.Args)
		result, err := tool.Func(toolCall.Id, toolCall.Args)
		thought.Phase = ThoughtPhaseObservation
		if err != nil {
			a.logger.Errorf("tool %s execution failed: %v", toolCall.ToolName, err)
			content := fmt.Sprintf("Call tool %s failed: %s.", toolCall.ToolName, err.Error())
			messages = append(messages, &LLMMessage{
				Role:       RoleTool,
				Content:    content,
				ToolCallId: toolCall.Id,
			})
			thought.Observation = content
		} else {
			resultJSON, _ := json.MarshalIndent(result, "", "  ")
			content := string(resultJSON)
			messages = append(messages, &LLMMessage{
				Role:       RoleTool,
				Content:    content,
				ToolCallId: toolCall.Id,
			})
			thought.Observation = content
		}
		err = sseWriter.Write(thought)
		if err != nil {
			return nil, fmt.Errorf("failed to write tool call: %w", err)
		}
		agentThoughts = append(agentThoughts, thought)
	}

	return agentThoughts, nil
}
