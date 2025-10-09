package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

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
		for _, toolCall := range resp.ToolCalls {
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
			messages = append(messages, &LLMMessage{
				Role:       RoleTool,
				Content:    fmt.Sprintf("Tool %s returned: %s", toolCall.ToolName, toolRes),
				ToolCallId: toolCall.Id,
			})
		}
	}

	return messages[len(messages)-1], nil
}
