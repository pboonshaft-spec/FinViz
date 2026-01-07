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

		// Monte Carlo Simulation Tools
		{
			Name:        "run_monte_carlo",
			Description: "Run a Monte Carlo simulation with specified parameters. Returns projections, success rate, milestones, and insights. Results are automatically saved to simulation history.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"time_horizon_years": map[string]interface{}{
						"type":        "integer",
						"description": "Number of years to project (1-80). Required.",
					},
					"monthly_contribution": map[string]interface{}{
						"type":        "number",
						"description": "Monthly investment contribution in dollars. Defaults to 0.",
					},
					"retirement_age": map[string]interface{}{
						"type":        "integer",
						"description": "Age at which retirement begins. Defaults to 65.",
					},
					"current_age": map[string]interface{}{
						"type":        "integer",
						"description": "Current age of the user. Required.",
					},
					"retirement_spending": map[string]interface{}{
						"type":        "number",
						"description": "Monthly spending in retirement in dollars. Defaults to 0.",
					},
					"expected_return": map[string]interface{}{
						"type":        "number",
						"description": "Expected annual return as decimal (0.07 = 7%). Defaults to 0.07.",
					},
					"volatility": map[string]interface{}{
						"type":        "number",
						"description": "Standard deviation as decimal (0.15 = 15%). Defaults to 0.15.",
					},
					"inflation_rate": map[string]interface{}{
						"type":        "number",
						"description": "Expected inflation rate as decimal (0.03 = 3%). Defaults to 0.03.",
					},
					"social_security_amount": map[string]interface{}{
						"type":        "number",
						"description": "Monthly Social Security benefit in dollars. Defaults to 0.",
					},
					"social_security_age": map[string]interface{}{
						"type":        "integer",
						"description": "Age Social Security begins (62-70). Defaults to 67.",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Optional name for this simulation scenario (e.g., 'Conservative estimate', 'Early retirement').",
					},
					"notes": map[string]interface{}{
						"type":        "string",
						"description": "Optional notes about this simulation.",
					},
				},
				"required": []string{"time_horizon_years", "current_age"},
			},
		},
		{
			Name:        "get_simulation_history",
			Description: "Retrieve past Monte Carlo simulations for the user. Returns a list of saved projections with their parameters and key results.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of simulations to return. Defaults to 10.",
					},
					"favorites_only": map[string]interface{}{
						"type":        "boolean",
						"description": "Only return favorite simulations. Defaults to false.",
					},
				},
				"required": []string{},
			},
		},
		{
			Name:        "get_simulation_details",
			Description: "Get detailed results from a specific saved simulation by ID. Returns full projections, milestones, and insights.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"simulation_id": map[string]interface{}{
						"type":        "integer",
						"description": "The ID of the simulation to retrieve.",
					},
				},
				"required": []string{"simulation_id"},
			},
		},
		{
			Name:        "compare_simulations",
			Description: "Compare two or more saved simulations side by side, showing differences in parameters and outcomes.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"simulation_ids": map[string]interface{}{
						"type":        "array",
						"description": "Array of simulation IDs to compare (2-5 simulations).",
						"items":       map[string]interface{}{"type": "integer"},
					},
				},
				"required": []string{"simulation_ids"},
			},
		},
		{
			Name:        "run_what_if_analysis",
			Description: "Run a 'what if' scenario analysis to answer questions like 'What if I retire 2 years earlier?' or 'What if I saved $500 more per month?'. Compares baseline parameters against modified parameters and returns a detailed impact analysis with natural language summary. Use this tool when the user asks hypothetical retirement or financial planning questions.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"question": map[string]interface{}{
						"type":        "string",
						"description": "The what-if question being answered, used for labeling the analysis (e.g., 'What if I retire at 62?').",
					},
					"baseline_simulation_id": map[string]interface{}{
						"type":        "integer",
						"description": "Optional ID of a saved simulation to use as baseline. If not provided, uses current assets/debts with default parameters.",
					},
					"monthly_contribution_change": map[string]interface{}{
						"type":        "number",
						"description": "Change to monthly contribution in dollars (positive or negative). E.g., 500 means save $500 more, -200 means save $200 less.",
					},
					"retirement_age_change": map[string]interface{}{
						"type":        "integer",
						"description": "Change to retirement age in years (positive or negative). E.g., -2 means retire 2 years earlier, 3 means retire 3 years later.",
					},
					"retirement_spending_change": map[string]interface{}{
						"type":        "number",
						"description": "Change to monthly retirement spending in dollars (positive or negative).",
					},
					"social_security_amount": map[string]interface{}{
						"type":        "number",
						"description": "Set Social Security monthly benefit to this amount (absolute, not a change).",
					},
					"social_security_age": map[string]interface{}{
						"type":        "integer",
						"description": "Set Social Security start age to this value (absolute, 62-70).",
					},
					"expected_return_change": map[string]interface{}{
						"type":        "number",
						"description": "Change to expected return as decimal (e.g., 0.01 means +1% return, -0.02 means -2% return).",
					},
					"volatility_change": map[string]interface{}{
						"type":        "number",
						"description": "Change to volatility as decimal (e.g., 0.05 means +5% volatility).",
					},
					"time_horizon_years": map[string]interface{}{
						"type":        "integer",
						"description": "Override time horizon for the analysis (if different from baseline).",
					},
					"current_age": map[string]interface{}{
						"type":        "integer",
						"description": "Current age (required if no baseline_simulation_id provided).",
					},
				},
				"required": []string{"question"},
			},
		},

		// Advanced Analysis Tools
		{
			Name:        "optimize_social_security",
			Description: "Analyze Social Security claiming strategies to find the optimal age to claim benefits. Compares claiming at ages 62-70 with lifetime benefit calculations, breakeven analysis, and personalized recommendations based on life expectancy.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"birth_date": map[string]interface{}{
						"type":        "string",
						"description": "Date of birth in YYYY-MM-DD format. Required.",
					},
					"estimated_pia": map[string]interface{}{
						"type":        "number",
						"description": "Primary Insurance Amount (monthly benefit at full retirement age). Get this from SSA.gov statement. Provide either this OR current_annual_earnings.",
					},
					"current_annual_earnings": map[string]interface{}{
						"type":        "number",
						"description": "Current annual earnings for PIA estimation. Used if estimated_pia not provided.",
					},
					"life_expectancy_years": map[string]interface{}{
						"type":        "integer",
						"description": "Expected age at death for lifetime benefit calculation. Defaults to 85.",
					},
					"spouse_birth_date": map[string]interface{}{
						"type":        "string",
						"description": "Spouse's birth date for spousal benefit analysis (optional, future feature).",
					},
				},
				"required": []string{"birth_date"},
			},
		},
		{
			Name:        "analyze_spending_patterns",
			Description: "Deep analysis of spending behavior and patterns from transaction history. Identifies recurring subscriptions, lifestyle inflation, savings rate trends, and provides actionable insights. Categorizes spending as essential vs discretionary.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"months": map[string]interface{}{
						"type":        "integer",
						"description": "Number of months to analyze. Defaults to 6, max 24.",
					},
					"compare_to_prior": map[string]interface{}{
						"type":        "boolean",
						"description": "Compare to prior period of same length. Defaults to false.",
					},
				},
				"required": []string{},
			},
		},
		{
			Name:        "check_portfolio_drift",
			Description: "Analyze portfolio allocation drift from target and recommend rebalancing trades. Compares current asset allocation to target, identifies over/under-weighted categories, and suggests specific buy/sell actions to rebalance.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"target_allocation": map[string]interface{}{
						"type":        "object",
						"description": "Target allocation by category (e.g., {\"Stocks\": 60, \"Bonds\": 30, \"Cash\": 10}). Must sum to 100. If not provided, uses age-based defaults.",
					},
					"drift_threshold": map[string]interface{}{
						"type":        "number",
						"description": "Percentage drift that triggers rebalancing recommendation. Defaults to 5.",
					},
					"age": map[string]interface{}{
						"type":        "integer",
						"description": "User's age for default allocation calculation (100-age rule for stocks). Only used if target_allocation not provided.",
					},
				},
				"required": []string{},
			},
		},
		{
			Name:        "project_tax_liability",
			Description: "Estimate current year federal tax liability based on income data. Calculates marginal and effective rates, shows bracket breakdown, and provides tax optimization suggestions for 401(k), IRA, HSA, and Roth conversions.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"filing_status": map[string]interface{}{
						"type":        "string",
						"description": "Tax filing status.",
						"enum":        []string{"single", "married_filing_jointly", "married_filing_separate", "head_of_household"},
					},
					"annual_income": map[string]interface{}{
						"type":        "number",
						"description": "Known annual income. If not provided, estimates from YTD transaction data.",
					},
					"itemized_deductions": map[string]interface{}{
						"type":        "number",
						"description": "Total itemized deductions if greater than standard deduction.",
					},
					"ytd_withholdings": map[string]interface{}{
						"type":        "number",
						"description": "Year-to-date tax withholdings for refund/owe estimation.",
					},
				},
				"required": []string{},
			},
		},
		{
			Name:        "analyze_tax_document",
			Description: "Analyze an uploaded tax document (1040, W-2, 1099) from the document vault. Extracts key financial data including income, deductions, credits, and tax liability. Generates optimization opportunities based on the tax data and user's current financial situation. Requires the document to be uploaded first via the Documents tab.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"document_id": map[string]interface{}{
						"type":        "integer",
						"description": "The ID of the tax document to analyze. Must be a PDF uploaded to the document vault, ideally in the tax_returns category.",
					},
				},
				"required": []string{"document_id"},
			},
		},

		// Report Generation Tool
		{
			Name:        "generate_report",
			Description: "Generate a professional PDF financial plan report for the user. The report includes net worth summary, asset/debt details, Monte Carlo projections, milestones, insights, and recommendations. Returns a download URL.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"include_simulation": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether to include Monte Carlo simulation in the report. Defaults to true.",
					},
					"time_horizon_years": map[string]interface{}{
						"type":        "integer",
						"description": "Number of years to project for the simulation. Defaults to 30.",
					},
					"monthly_contribution": map[string]interface{}{
						"type":        "number",
						"description": "Monthly investment contribution for the simulation. Defaults to 0.",
					},
					"retirement_age": map[string]interface{}{
						"type":        "integer",
						"description": "Age at which retirement begins. Defaults to 65.",
					},
					"current_age": map[string]interface{}{
						"type":        "integer",
						"description": "Current age of the user. Defaults to 35.",
					},
					"retirement_spending": map[string]interface{}{
						"type":        "number",
						"description": "Monthly spending in retirement. Defaults to 0.",
					},
				},
				"required": []string{},
			},
		},
	}
}

// GetAdvisorTools returns additional tools available only to advisors
func GetAdvisorTools() []Tool {
	return []Tool{
		{
			Name:        "list_clients",
			Description: "List all clients associated with this advisor. Returns client names, net worth summary, and last activity.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},
		{
			Name:        "switch_client_context",
			Description: "Switch to working with a specific client. All subsequent data tools will operate on the selected client's data until switched back.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"client_id": map[string]interface{}{
						"type":        "integer",
						"description": "The client ID to switch to. Use 0 to switch back to advisor's own view.",
					},
				},
				"required": []string{"client_id"},
			},
		},
		{
			Name:        "get_client_summary",
			Description: "Get a comprehensive summary for a specific client including net worth, recent activity, and simulation history.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"client_id": map[string]interface{}{
						"type":        "integer",
						"description": "The client ID to summarize.",
					},
				},
				"required": []string{"client_id"},
			},
		},

		// Meeting Prep Tool
		{
			Name:        "generate_meeting_prep",
			Description: "Generate a comprehensive client briefing document for an upcoming meeting. Includes net worth summary, spending trends, simulation review, agenda items, and talking points. Automatically incorporates advisor notes including pinned notes and action items.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"client_id": map[string]interface{}{
						"type":        "integer",
						"description": "The client ID to prepare for. If omitted, uses the current client context.",
					},
					"focus_areas": map[string]interface{}{
						"type":        "array",
						"description": "Optional array of focus areas to include in the agenda. Options: 'retirement', 'tax', 'estate', 'budget', 'insurance', 'goals'.",
						"items":       map[string]interface{}{"type": "string"},
					},
				},
				"required": []string{},
			},
		},

		// Client Notes Tools
		{
			Name:        "get_client_notes",
			Description: "Retrieve all notes for a client. Notes are organized by category (general, meeting, goal, concern, action_item, personal) and can be pinned for priority.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"client_id": map[string]interface{}{
						"type":        "integer",
						"description": "The client ID. If omitted, uses the current client context.",
					},
					"category": map[string]interface{}{
						"type":        "string",
						"description": "Optional filter by category.",
						"enum":        []string{"general", "meeting", "goal", "concern", "action_item", "personal"},
					},
				},
				"required": []string{},
			},
		},
		{
			Name:        "add_client_note",
			Description: "Add a new note to a client's file. Notes are visible only to the advisor and can be used for meeting prep, tracking action items, and remembering important client details.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"client_id": map[string]interface{}{
						"type":        "integer",
						"description": "The client ID. If omitted, uses the current client context.",
					},
					"note": map[string]interface{}{
						"type":        "string",
						"description": "The note content. Required.",
					},
					"category": map[string]interface{}{
						"type":        "string",
						"description": "Category for the note. Defaults to 'general'.",
						"enum":        []string{"general", "meeting", "goal", "concern", "action_item", "personal"},
					},
					"is_pinned": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether to pin this note for priority display. Defaults to false.",
					},
				},
				"required": []string{"note"},
			},
		},
		{
			Name:        "update_client_note",
			Description: "Update an existing client note. Can modify the content, category, or pinned status.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"note_id": map[string]interface{}{
						"type":        "integer",
						"description": "The ID of the note to update. Required.",
					},
					"note": map[string]interface{}{
						"type":        "string",
						"description": "New note content.",
					},
					"category": map[string]interface{}{
						"type":        "string",
						"description": "New category for the note.",
						"enum":        []string{"general", "meeting", "goal", "concern", "action_item", "personal"},
					},
					"is_pinned": map[string]interface{}{
						"type":        "boolean",
						"description": "New pinned status.",
					},
				},
				"required": []string{"note_id"},
			},
		},
		{
			Name:        "delete_client_note",
			Description: "Delete a client note by ID.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"note_id": map[string]interface{}{
						"type":        "integer",
						"description": "The ID of the note to delete. Required.",
					},
				},
				"required": []string{"note_id"},
			},
		},
	}
}
