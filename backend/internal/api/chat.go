package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/finviz/backend/internal/claude"
)

var claudeClient *claude.Client
var aureliaPrompt string

func init() {
	claudeClient = claude.NewClient()

	// Load the Aurelia system prompt
	promptPath := os.Getenv("AURELIA_PROMPT_PATH")
	if promptPath == "" {
		promptPath = "/app/config/aurelia_prompt.txt"
	}

	promptBytes, err := os.ReadFile(promptPath)
	if err != nil {
		// Fallback to embedded prompt if file not found
		aureliaPrompt = getDefaultAureliaPrompt()
	} else {
		aureliaPrompt = string(promptBytes)
	}

	claudeClient.SetSystemPrompt(aureliaPrompt)
	claudeClient.SetTools(claude.GetAureliaTools())
}

// ChatRequest represents the incoming chat request
type ChatRequest struct {
	Messages       []ChatMessage `json:"messages"`
	ConversationID string        `json:"conversationId,omitempty"`
}

// ChatMessage represents a message in the conversation
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse represents the response from the chat endpoint
type ChatResponse struct {
	Response   string                   `json:"response"`
	ToolsUsed  []string                 `json:"toolsUsed,omitempty"`
	Artifacts  []map[string]interface{} `json:"artifacts,omitempty"`
	TokenUsage map[string]int           `json:"tokenUsage,omitempty"`
}

// handleChat handles chat requests with the Aurelia agent
func handleChat(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	if !claudeClient.IsConfigured() {
		respondError(w, http.StatusServiceUnavailable, "Chat service is not configured. Please set ANTHROPIC_API_KEY.")
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.Messages) == 0 {
		respondError(w, http.StatusBadRequest, "At least one message is required")
		return
	}

	// Convert chat messages to Claude format
	messages := convertToClaude(req.Messages)

	// Create tool executor for this user
	toolExecutor := claude.NewToolExecutor(user.ID)

	// Track tools used and artifacts
	var toolsUsed []string
	var artifacts []map[string]interface{}

	// Agentic loop: continue until we get a final response (not tool_use)
	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		response, err := claudeClient.SendMessage(messages)
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("Chat error: %v", err))
			return
		}

		// Check if we have tool calls to execute
		if claudeClient.HasToolUse(response) {
			toolCalls := claudeClient.ExtractToolCalls(response)

			// Execute each tool
			results := make(map[string]string)
			for _, tc := range toolCalls {
				toolsUsed = append(toolsUsed, tc.Name)

				result, err := toolExecutor.ExecuteTool(tc.Name, tc.Input)
				if err != nil {
					results[tc.ID] = fmt.Sprintf("Error: %v", err)
				} else {
					results[tc.ID] = result

					// Check if this is an artifact (chart, table, metric_card)
					if isArtifactTool(tc.Name) {
						var artifact map[string]interface{}
						if json.Unmarshal([]byte(result), &artifact) == nil {
							artifacts = append(artifacts, artifact)
						}
					}
				}
			}

			// Add the assistant's response (with tool use) to messages
			messages = append(messages, claudeClient.CreateAssistantMessage(response))

			// Add tool results to messages
			messages = append(messages, claudeClient.CreateToolResultMessage(toolCalls, results))

			// Continue the loop to get the next response
			continue
		}

		// No tool use - we have a final response
		textResponse := claudeClient.ExtractTextResponse(response)

		resp := ChatResponse{
			Response:  textResponse,
			ToolsUsed: uniqueStrings(toolsUsed),
			Artifacts: artifacts,
			TokenUsage: map[string]int{
				"input":  response.Usage.InputTokens,
				"output": response.Usage.OutputTokens,
			},
		}

		respondJSON(w, http.StatusOK, resp)
		return
	}

	respondError(w, http.StatusInternalServerError, "Max iterations reached without a final response")
}

// handleChatStatus returns whether chat is available
func handleChatStatus(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]bool{
		"configured": claudeClient.IsConfigured(),
	})
}

// convertToClaude converts chat messages to Claude API format
func convertToClaude(messages []ChatMessage) []claude.Message {
	var result []claude.Message

	for _, msg := range messages {
		result = append(result, claude.Message{
			Role: msg.Role,
			Content: []claude.ContentBlock{
				{
					Type: "text",
					Text: msg.Content,
				},
			},
		})
	}

	return result
}

// isArtifactTool checks if a tool produces a renderable artifact
func isArtifactTool(name string) bool {
	artifactTools := []string{"create_chart", "create_table", "create_metric_card"}
	for _, t := range artifactTools {
		if name == t {
			return true
		}
	}
	return false
}

// uniqueStrings returns unique strings from a slice
func uniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// getDefaultAureliaPrompt returns a fallback prompt if file not found
func getDefaultAureliaPrompt() string {
	return strings.TrimSpace(`
You are "Aurelia," a seasoned financial wealth advisor with over 30 years of experience in wealth management and strategic investing.

Your core attributes:
- Data-driven and evidence-based in all recommendations
- Direct and clear without being dismissive of human concerns
- Focused on long-term wealth accumulation and preservation
- Committed to explaining complex concepts in accessible language
- Respectful of user autonomy while providing honest assessments

You have access to the user's financial data through tools. Use these tools to provide personalized advice:
- get_user_assets: View their investment accounts, real estate, and other assets
- get_user_debts: View their loans, mortgages, and other liabilities
- get_user_transactions: View their recent spending and income
- get_net_worth_summary: Get a complete financial picture
- get_monthly_cash_flow: Analyze their income vs expenses

You can also research topics:
- web_search: Search for current financial information
- search_tax_code: Look up IRS publications and tax rules
- get_current_rates: Get current tax brackets, contribution limits, etc.

When visualizing data, use:
- create_chart: Create pie, bar, line, or area charts
- create_table: Create formatted data tables
- create_metric_card: Highlight key financial metrics

Always provide a disclaimer that you are an AI assistant and not a licensed financial advisor. Recommend consulting professionals for significant financial decisions.
`)
}
