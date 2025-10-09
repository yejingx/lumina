package main

import (
	"context"
	"io"
	"net/http"
	"os"
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

type HttpToolParams struct {
	URL string `json:"url" jsonschema_description:"URL to fetch"`
}

type HttpToolResult struct {
	agent.BaseToolResult
	StatusCode int               `json:"status_code" jsonschema_description:"HTTP status code"`
	Body       string            `json:"body"        jsonschema_description:"Response body"`
	Headers    map[string]string `json:"headers"     jsonschema_description:"Response headers"`
}

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

func createErrorResult(callID string, statusCode int, errorMsg string) HttpToolResult {
	return HttpToolResult{
		BaseToolResult: agent.BaseToolResult{Id: callID},
		StatusCode:     statusCode,
		Body:           errorMsg,
		Headers:        map[string]string{},
	}
}

func extractHeaders(header http.Header) map[string]string {
	headers := make(map[string]string)
	for name, values := range header {
		if len(values) > 0 {
			headers[name] = values[0]
		}
	}

	return headers
}

func handleHttpRequest(callID string, params HttpToolParams) (HttpToolResult, error) {
	logrus.Debugf("HTTP CALL: GET %s", params.URL)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, params.URL, nil)
	if err != nil {
		return createErrorResult(callID, 0, "Error creating request: "+err.Error()), nil
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return createErrorResult(callID, 0, "Error: "+err.Error()), nil
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logrus.Errorf("Error closing response body: %v", closeErr)
		}
	}()

	headers := extractHeaders(resp.Header)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return HttpToolResult{
			BaseToolResult: agent.BaseToolResult{Id: callID},
			StatusCode:     resp.StatusCode,
			Body:           "Error reading response body: " + err.Error(),
			Headers:        headers,
		}, nil
	}

	result := HttpToolResult{
		BaseToolResult: agent.BaseToolResult{Id: callID},
		StatusCode:     resp.StatusCode,
		Body:           string(body),
		Headers:        headers,
	}

	logrus.Debugf("HTTP RESULT: %d - %d bytes", resp.StatusCode, len(body))

	return result, nil
}

func registerTestRools(a *agent.Agent) {
	a.AddTool(agent.NewTool(
		agent.WithToolName("http"),
		agent.WithToolDescription("Fetches content from a URL via HTTP GET request"),
		agent.WithToolParamsSchema[HttpToolParams](),
		agent.WithToolFunc(handleHttpRequest),
	))
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
	registerTestRools(agent)

	ctx := context.Background()
	result, err := agent.Run(ctx, query)
	if err != nil {
		logrus.Errorf("Error running agent: %v", err)
		return
	}
	logrus.Infof("Agent result: %v", result.Content)
}
