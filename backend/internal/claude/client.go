package claude

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	anthropicAPIURL = "https://api.anthropic.com/v1/messages"
	defaultModel    = "claude-sonnet-4-20250514"
	maxTokens       = 4096
)

// Client handles communication with the Claude API
type Client struct {
	apiKey       string
	httpClient   *http.Client
	systemPrompt string
	tools        []Tool
}

// Message represents a conversation message
type Message struct {
	Role    string        `json:"role"`
	Content []ContentBlock `json:"content,omitempty"`
}

// ContentBlock represents content in a message
type ContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   string          `json:"content,omitempty"`
}

// Tool represents a tool available to Claude
type Tool struct {
	// For custom function tools
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"input_schema,omitempty"`

	// For built-in tools like web_search
	Type           string   `json:"type,omitempty"`
	MaxUses        int      `json:"max_uses,omitempty"`
	AllowedDomains []string `json:"allowed_domains,omitempty"`
	BlockedDomains []string `json:"blocked_domains,omitempty"`
}

// IsBuiltInTool returns true if this is a built-in Claude tool (not a custom function)
func (t Tool) IsBuiltInTool() bool {
	return t.Type != ""
}

// Request represents an API request to Claude
type Request struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
	Tools     []Tool    `json:"tools,omitempty"`
}

// Response represents an API response from Claude
type Response struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason"`
	StopSequence string         `json:"stop_sequence,omitempty"`
	Usage        Usage          `json:"usage"`
}

// Usage represents token usage
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ToolCall represents a tool call extracted from a response
type ToolCall struct {
	ID    string
	Name  string
	Input map[string]interface{}
}

// NewClient creates a new Claude API client
func NewClient() *Client {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("Warning: ANTHROPIC_API_KEY not set")
	}

	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// SetSystemPrompt sets the system prompt for conversations
func (c *Client) SetSystemPrompt(prompt string) {
	c.systemPrompt = prompt
}

// SetTools sets the available tools
func (c *Client) SetTools(tools []Tool) {
	c.tools = tools
}

// IsConfigured returns whether the client has an API key
func (c *Client) IsConfigured() bool {
	return c.apiKey != ""
}

// SendMessage sends a message to Claude and returns the response
func (c *Client) SendMessage(messages []Message) (*Response, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("Claude API key not configured")
	}

	req := Request{
		Model:     defaultModel,
		MaxTokens: maxTokens,
		System:    c.systemPrompt,
		Messages:  messages,
		Tools:     c.tools,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", anthropicAPIURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var response Response
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// ExtractToolCalls extracts tool calls from a response
func (c *Client) ExtractToolCalls(response *Response) []ToolCall {
	var toolCalls []ToolCall

	for _, block := range response.Content {
		if block.Type == "tool_use" {
			var input map[string]interface{}
			if err := json.Unmarshal(block.Input, &input); err != nil {
				input = make(map[string]interface{})
			}

			toolCalls = append(toolCalls, ToolCall{
				ID:    block.ID,
				Name:  block.Name,
				Input: input,
			})
		}
	}

	return toolCalls
}

// ExtractTextResponse extracts the text response from a response
func (c *Client) ExtractTextResponse(response *Response) string {
	for _, block := range response.Content {
		if block.Type == "text" {
			return block.Text
		}
	}
	return ""
}

// HasToolUse checks if the response contains tool use
func (c *Client) HasToolUse(response *Response) bool {
	return response.StopReason == "tool_use"
}

// CreateToolResultMessage creates a message with tool results
func (c *Client) CreateToolResultMessage(toolCalls []ToolCall, results map[string]string) Message {
	var content []ContentBlock

	for _, tc := range toolCalls {
		result, ok := results[tc.ID]
		if !ok {
			result = "Tool execution failed"
		}

		content = append(content, ContentBlock{
			Type:      "tool_result",
			ToolUseID: tc.ID,
			Content:   result,
		})
	}

	return Message{
		Role:    "user",
		Content: content,
	}
}

// CreateAssistantMessage creates an assistant message from a response
func (c *Client) CreateAssistantMessage(response *Response) Message {
	return Message{
		Role:    "assistant",
		Content: response.Content,
	}
}
