package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"lumina/internal/agent"
)

const instruction = `You are a website analysis expert specializing in ` +
	`comprehensive site evaluation and content extraction.

ANALYSIS PROCEDURE:
1. INITIAL FETCH: Use the http tool to fetch the main page content
2. CONTENT ANALYSIS: Analyze HTML structure, meta tags, headings, and visible text
3. DEEP EXPLORATION: Look for additional pages, contact info, about sections, or portfolio links
4. STRUCTURE MAPPING: Identify navigation patterns, page hierarchy, and site organization
5. PURPOSE IDENTIFICATION: Determine the primary function and target audience
6. INSIGHT EXTRACTION: Extract key technical details, business model, and unique features

FOCUS AREAS:
- Site title, description, and branding elements
- Main content themes and messaging
- Technical stack indicators (frameworks, libraries)
- Business/personal information and contact details
- Key features, services, or products offered
- Design patterns and user experience elements

THOROUGHNESS REQUIREMENT:
DO NOT BE LAZY! You must continue analyzing until you have exhausted all ` +
	`available information or reached the tool limit. Start with the main page, ` +
	`then explore additional pages like:
- /about, /contact, /portfolio, /services, /products
- Any links found in navigation menus or footer
- Subpages that provide more context about the site owner or business
- Continue fetching pages until you have a complete picture or hit the 3-request limit

OUTPUT REQUIREMENTS:
- Provide a clear, descriptive title based on actual content
- Summarize the site's primary purpose in 1-2 sentences
- List 5-10 key insights that reveal important aspects of the site`

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Agent for lumina",
	Long:  `Agent for lumina application.`,
}

var testAgentCmd = &cobra.Command{
	Use:   "test <query>",
	Short: "Test agent",
	Long:  `Test agent.`,
	Args:  cobra.ExactArgs(1), // Ensure exactly one argument is provided
	Run: func(cmd *cobra.Command, args []string) {
		testAgent(args[0])
	},
}

func init() {
	agentCmd.AddCommand(testAgentCmd)
}

func testAgent(query string) {
	baseUrl := os.Getenv("OPENAI_API_BASE_URL")
	if baseUrl == "" {
		baseUrl = "https://api.openai.com/v1"
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	modelName := os.Getenv("OPENAI_API_MODEL")
	if modelName == "" {
		modelName = "gpt-3.5-turbo"
	}
	agent := agent.NewAgent("test", agent.LLMConfig{
		BaseUrl:     baseUrl,
		ApiKey:      apiKey,
		Model:       modelName,
		Temperature: 0.6,
		Timeout:     300 * time.Second,
	}, 10, instruction)

	ctx := context.Background()

	// Create a custom writer to parse SSE format data
	customWriter := &SSEWriter{output: os.Stdout}

	err := agent.RunStream(ctx, query, customWriter)
	if err != nil {
		logrus.Errorf("Error running agent stream: %v", err)
		return
	}
}

// SSEWriter parses SSE format data and prints only content or tool call parameters
type SSEWriter struct {
	output io.Writer
}

func (w *SSEWriter) Write(p []byte) (n int, err error) {
	data := string(p)

	// Handle SSE format: "data: {...}\n"
	if jsonStr, ok := strings.CutPrefix(data, "data: "); ok {
		jsonStr = strings.TrimSpace(jsonStr)

		// Skip [DONE] message
		if jsonStr == "[DONE]" {
			return len(p), nil
		}

		// Parse JSON message
		var msg agent.LLMMessage
		if err := json.Unmarshal([]byte(jsonStr), &msg); err != nil {
			// If parsing fails, just write original data
			return w.output.Write(p)
		}

		// Print content if available
		if msg.Content != "" {
			w.output.Write([]byte(msg.Content))
		}

		// Print tool calls if available
		if len(msg.ToolCalls) > 0 {
			for _, toolCall := range msg.ToolCalls {
				toolInfo := fmt.Sprintf("\n\n🔧 调用工具: %s, 参数: %s\n\n", toolCall.ToolName, toolCall.Args)
				w.output.Write([]byte(toolInfo))
			}
		}

		return len(p), nil
	}

	// For non-SSE data, write as is
	return w.output.Write(p)
}
