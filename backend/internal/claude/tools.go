package claude

// GetAureliaTools returns all tools available to the Aurelia agent
func GetAureliaTools() []Tool {
	return []Tool{
		// App Data Tools
		{
			Name:        "get_user_assets",
			Description: "Fetch all of the user's assets including investment accounts, real estate, cash, and other holdings. Returns asset names, types, current values, and expected returns.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},
		{
			Name:        "get_user_debts",
			Description: "Fetch all of the user's debts including mortgages, loans, credit cards, and other liabilities. Returns debt names, current balances, interest rates, and minimum payments.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},
		{
			Name:        "get_user_transactions",
			Description: "Fetch the user's transaction history for a given date range. Returns transaction details including date, amount, category, and merchant.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"start_date": map[string]interface{}{
						"type":        "string",
						"description": "Start date in YYYY-MM-DD format. Defaults to 30 days ago.",
					},
					"end_date": map[string]interface{}{
						"type":        "string",
						"description": "End date in YYYY-MM-DD format. Defaults to today.",
					},
					"category": map[string]interface{}{
						"type":        "string",
						"description": "Optional category filter (e.g., 'FOOD_AND_DRINK', 'TRAVEL', 'SHOPPING').",
					},
				},
				"required": []string{},
			},
		},
		{
			Name:        "get_net_worth_summary",
			Description: "Get a summary of the user's net worth including total assets, total debts, and net worth calculation. Also includes breakdown by asset type.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},
		{
			Name:        "get_monthly_cash_flow",
			Description: "Get the user's income and expense summary for recent months. Returns total income, total expenses, net cash flow, and breakdown by category.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"months": map[string]interface{}{
						"type":        "integer",
						"description": "Number of months to analyze. Defaults to 3.",
					},
				},
				"required": []string{},
			},
		},

		// Built-in Web Search Tool (Claude handles this automatically)
		{
			Type:    "web_search_20250305",
			Name:    "web_search",
			MaxUses: 5, // Limit searches per request
		},
		{
			Name:        "get_current_rates",
			Description: "Get current financial rates and data including federal funds rate, tax brackets, contribution limits, and other commonly referenced financial figures.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"rate_type": map[string]interface{}{
						"type":        "string",
						"description": "Type of rate to retrieve. Options: 'federal_funds', 'tax_brackets', '401k_limits', 'ira_limits', 'hsa_limits', 'social_security'.",
						"enum":        []string{"federal_funds", "tax_brackets", "401k_limits", "ira_limits", "hsa_limits", "social_security"},
					},
				},
				"required": []string{"rate_type"},
			},
		},

		// UI/Visualization Tools (A2UI)
		{
			Name:        "create_chart",
			Description: "Create a chart visualization to display to the user. Use this to show data visually such as asset allocation pie charts, net worth trends, spending breakdowns, or projection charts.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"chart_type": map[string]interface{}{
						"type":        "string",
						"description": "Type of chart to create.",
						"enum":        []string{"pie", "bar", "line", "area", "donut"},
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Title for the chart.",
					},
					"data": map[string]interface{}{
						"type":        "array",
						"description": "Array of data points. Each point should have 'label' and 'value' properties.",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"label": map[string]interface{}{"type": "string"},
								"value": map[string]interface{}{"type": "number"},
							},
						},
					},
					"colors": map[string]interface{}{
						"type":        "array",
						"description": "Optional array of hex color codes for the data series.",
						"items":       map[string]interface{}{"type": "string"},
					},
				},
				"required": []string{"chart_type", "title", "data"},
			},
		},
		{
			Name:        "create_table",
			Description: "Create a formatted table to display structured data to the user. Use this for comparison tables, amortization schedules, or detailed breakdowns.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Title for the table.",
					},
					"headers": map[string]interface{}{
						"type":        "array",
						"description": "Array of column header strings.",
						"items":       map[string]interface{}{"type": "string"},
					},
					"rows": map[string]interface{}{
						"type":        "array",
						"description": "Array of rows, where each row is an array of cell values.",
						"items": map[string]interface{}{
							"type":  "array",
							"items": map[string]interface{}{"type": "string"},
						},
					},
				},
				"required": []string{"title", "headers", "rows"},
			},
		},
		{
			Name:        "create_metric_card",
			Description: "Create a metric card to highlight a key financial figure. Use this for net worth, savings rate, or other important single values.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"label": map[string]interface{}{
						"type":        "string",
						"description": "Label for the metric (e.g., 'Net Worth', 'Monthly Savings').",
					},
					"value": map[string]interface{}{
						"type":        "string",
						"description": "Formatted value to display (e.g., '$125,000', '15%').",
					},
					"change": map[string]interface{}{
						"type":        "string",
						"description": "Optional change indicator (e.g., '+5% this month').",
					},
					"trend": map[string]interface{}{
						"type":        "string",
						"description": "Trend direction for styling.",
						"enum":        []string{"up", "down", "neutral"},
					},
				},
				"required": []string{"label", "value"},
			},
		},
	}
}
