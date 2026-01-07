package claude

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/finviz/backend/internal/db"
	"github.com/finviz/backend/internal/models"
	"github.com/finviz/backend/internal/reports"
	"github.com/finviz/backend/internal/simulation"
	"github.com/finviz/backend/internal/storage"
	"github.com/finviz/backend/internal/taxparser"
)

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	return math.Abs(x)
}

// ToolExecutor handles execution of tools for a specific user
type ToolExecutor struct {
	UserID        int
	IsAdvisor     bool
	ClientContext int // Non-zero when advisor is working with a client
}

// NewToolExecutor creates a new tool executor for a user
func NewToolExecutor(userID int) *ToolExecutor {
	return &ToolExecutor{UserID: userID}
}

// NewAdvisorToolExecutor creates a new tool executor for an advisor
func NewAdvisorToolExecutor(advisorID int, clientID int) *ToolExecutor {
	return &ToolExecutor{
		UserID:        advisorID,
		IsAdvisor:     true,
		ClientContext: clientID,
	}
}

// GetEffectiveUserID returns the user ID to use for data operations
func (e *ToolExecutor) GetEffectiveUserID() int {
	if e.ClientContext != 0 {
		return e.ClientContext
	}
	return e.UserID
}

// ExecuteTool executes a tool and returns the result as a string
func (e *ToolExecutor) ExecuteTool(name string, input map[string]interface{}) (string, error) {
	switch name {
	case "get_user_assets":
		return e.getUserAssets()
	case "get_user_debts":
		return e.getUserDebts()
	case "get_user_transactions":
		return e.getUserTransactions(input)
	case "get_net_worth_summary":
		return e.getNetWorthSummary()
	case "get_monthly_cash_flow":
		return e.getMonthlyCashFlow(input)
	case "get_current_rates":
		return e.getCurrentRates(input)
	case "create_chart":
		return e.createChart(input)
	case "create_table":
		return e.createTable(input)
	case "create_metric_card":
		return e.createMetricCard(input)
	// Monte Carlo tools
	case "run_monte_carlo":
		return e.runMonteCarlo(input)
	case "get_simulation_history":
		return e.getSimulationHistory(input)
	case "get_simulation_details":
		return e.getSimulationDetails(input)
	case "compare_simulations":
		return e.compareSimulations(input)
	case "run_what_if_analysis":
		return e.runWhatIfAnalysis(input)
	case "generate_report":
		return e.generateReport(input)
	// Advanced Analysis Tools
	case "optimize_social_security":
		return e.optimizeSocialSecurity(input)
	case "analyze_spending_patterns":
		return e.analyzeSpendingPatterns(input)
	case "check_portfolio_drift":
		return e.checkPortfolioDrift(input)
	case "project_tax_liability":
		return e.projectTaxLiability(input)
	case "analyze_tax_document":
		return e.analyzeTaxDocument(input)
	case "generate_meeting_prep":
		return e.generateMeetingPrep(input)
	// Advisor-only tools
	case "list_clients":
		return e.listClients()
	case "switch_client_context":
		return e.switchClientContext(input)
	case "get_client_summary":
		return e.getClientSummary(input)
	// Client notes tools (advisor-only)
	case "get_client_notes":
		return e.getClientNotes(input)
	case "add_client_note":
		return e.addClientNote(input)
	case "update_client_note":
		return e.updateClientNote(input)
	case "delete_client_note":
		return e.deleteClientNote(input)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

// getUserAssets fetches all assets for the user
func (e *ToolExecutor) getUserAssets() (string, error) {
	userID := e.GetEffectiveUserID()
	rows, err := db.DB.Query(`
		SELECT a.id, a.name, at.name as type_name, a.current_value,
			   COALESCE(a.custom_return, at.default_return) as expected_return,
			   COALESCE(a.custom_volatility, at.default_volatility) as volatility
		FROM assets a
		LEFT JOIN asset_types at ON a.type_id = at.id
		WHERE a.user_id = ?
		ORDER BY a.current_value DESC
	`, userID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type Asset struct {
		ID             int      `json:"id"`
		Name           string   `json:"name"`
		Type           string   `json:"type"`
		CurrentValue   float64  `json:"current_value"`
		ExpectedReturn *float64 `json:"expected_return,omitempty"`
		Volatility     *float64 `json:"volatility,omitempty"`
	}

	var assets []Asset
	var totalValue float64

	for rows.Next() {
		var a Asset
		if err := rows.Scan(&a.ID, &a.Name, &a.Type, &a.CurrentValue, &a.ExpectedReturn, &a.Volatility); err != nil {
			continue
		}
		assets = append(assets, a)
		totalValue += a.CurrentValue
	}

	result := map[string]interface{}{
		"assets":      assets,
		"total_value": totalValue,
		"count":       len(assets),
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonBytes), nil
}

// getUserDebts fetches all debts for the user
func (e *ToolExecutor) getUserDebts() (string, error) {
	userID := e.GetEffectiveUserID()
	rows, err := db.DB.Query(`
		SELECT id, name, current_balance, interest_rate, minimum_payment
		FROM debts
		WHERE user_id = ?
		ORDER BY current_balance DESC
	`, userID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type Debt struct {
		ID             int      `json:"id"`
		Name           string   `json:"name"`
		CurrentBalance float64  `json:"current_balance"`
		InterestRate   *float64 `json:"interest_rate,omitempty"`
		MinimumPayment *float64 `json:"minimum_payment,omitempty"`
	}

	var debts []Debt
	var totalDebt float64

	for rows.Next() {
		var d Debt
		if err := rows.Scan(&d.ID, &d.Name, &d.CurrentBalance, &d.InterestRate, &d.MinimumPayment); err != nil {
			continue
		}
		debts = append(debts, d)
		totalDebt += d.CurrentBalance
	}

	result := map[string]interface{}{
		"debts":      debts,
		"total_debt": totalDebt,
		"count":      len(debts),
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonBytes), nil
}

// getUserTransactions fetches transactions for the user
func (e *ToolExecutor) getUserTransactions(input map[string]interface{}) (string, error) {
	// Default date range: last 30 days
	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")

	if sd, ok := input["start_date"].(string); ok && sd != "" {
		startDate = sd
	}
	if ed, ok := input["end_date"].(string); ok && ed != "" {
		endDate = ed
	}

	query := `
		SELECT id, name, amount, date, category, merchant_name
		FROM transactions
		WHERE user_id = ? AND date >= ? AND date <= ?
	`
	args := []interface{}{e.GetEffectiveUserID(), startDate, endDate}

	if category, ok := input["category"].(string); ok && category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}

	query += " ORDER BY date DESC LIMIT 100"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type Transaction struct {
		ID           int     `json:"id"`
		Name         string  `json:"name"`
		Amount       float64 `json:"amount"`
		Date         string  `json:"date"`
		Category     *string `json:"category,omitempty"`
		MerchantName *string `json:"merchant_name,omitempty"`
	}

	var transactions []Transaction
	var totalIncome, totalExpenses float64

	// Income categories that should be counted as income regardless of amount sign
	incomeCategories := map[string]bool{
		"INCOME": true, "INCOME_WAGES": true, "INCOME_DIVIDENDS": true,
		"INCOME_INTEREST": true, "TRANSFER_IN": true,
	}

	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.ID, &t.Name, &t.Amount, &t.Date, &t.Category, &t.MerchantName); err != nil {
			continue
		}
		transactions = append(transactions, t)

		// Check if category indicates income
		cat := ""
		if t.Category != nil {
			cat = *t.Category
		}
		isIncomeCategory := incomeCategories[cat]

		if t.Amount < 0 || isIncomeCategory {
			totalIncome += abs(t.Amount)
		} else {
			totalExpenses += t.Amount
		}
	}

	result := map[string]interface{}{
		"transactions":   transactions,
		"count":          len(transactions),
		"date_range":     map[string]string{"start": startDate, "end": endDate},
		"total_income":   totalIncome,
		"total_expenses": totalExpenses,
		"net_cash_flow":  totalIncome - totalExpenses,
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonBytes), nil
}

// getNetWorthSummary calculates net worth summary
func (e *ToolExecutor) getNetWorthSummary() (string, error) {
	userID := e.GetEffectiveUserID()
	// Get total assets by type
	assetRows, err := db.DB.Query(`
		SELECT COALESCE(at.name, 'Other') as type_name, SUM(a.current_value) as total
		FROM assets a
		LEFT JOIN asset_types at ON a.type_id = at.id
		WHERE a.user_id = ?
		GROUP BY at.name
		ORDER BY total DESC
	`, userID)
	if err != nil {
		return "", err
	}
	defer assetRows.Close()

	assetsByType := make(map[string]float64)
	var totalAssets float64

	for assetRows.Next() {
		var typeName string
		var total float64
		if err := assetRows.Scan(&typeName, &total); err != nil {
			continue
		}
		assetsByType[typeName] = total
		totalAssets += total
	}

	// Get total debts
	var totalDebts float64
	err = db.DB.QueryRow(`SELECT COALESCE(SUM(current_balance), 0) FROM debts WHERE user_id = ?`, userID).Scan(&totalDebts)
	if err != nil {
		totalDebts = 0
	}

	result := map[string]interface{}{
		"total_assets":   totalAssets,
		"total_debts":    totalDebts,
		"net_worth":      totalAssets - totalDebts,
		"assets_by_type": assetsByType,
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonBytes), nil
}

// getMonthlyCashFlow calculates monthly cash flow
func (e *ToolExecutor) getMonthlyCashFlow(input map[string]interface{}) (string, error) {
	userID := e.GetEffectiveUserID()
	months := 3
	if m, ok := input["months"].(float64); ok {
		months = int(m)
	}

	startDate := time.Now().AddDate(0, -months, 0).Format("2006-01-02")

	rows, err := db.DB.Query(`
		SELECT
			DATE_FORMAT(date, '%Y-%m') as month,
			COALESCE(SUM(CASE
				WHEN amount < 0 THEN ABS(amount)
				WHEN category IN ('INCOME', 'INCOME_WAGES', 'INCOME_DIVIDENDS', 'INCOME_INTEREST', 'TRANSFER_IN') THEN ABS(amount)
				WHEN subcategory LIKE 'INCOME%' OR subcategory LIKE 'TRANSFER_IN%' THEN ABS(amount)
				ELSE 0
			END), 0) as income,
			COALESCE(SUM(CASE
				WHEN amount > 0
				AND category NOT IN ('INCOME', 'INCOME_WAGES', 'INCOME_DIVIDENDS', 'INCOME_INTEREST', 'TRANSFER_IN')
				AND (subcategory IS NULL OR (subcategory NOT LIKE 'INCOME%' AND subcategory NOT LIKE 'TRANSFER_IN%'))
				THEN amount
				ELSE 0
			END), 0) as expenses
		FROM transactions
		WHERE user_id = ? AND date >= ?
		GROUP BY DATE_FORMAT(date, '%Y-%m')
		ORDER BY month DESC
	`, userID, startDate)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type MonthlyData struct {
		Month       string  `json:"month"`
		Income      float64 `json:"income"`
		Expenses    float64 `json:"expenses"`
		NetCashFlow float64 `json:"net_cash_flow"`
	}

	var monthlyData []MonthlyData
	var totalIncome, totalExpenses float64

	for rows.Next() {
		var m MonthlyData
		if err := rows.Scan(&m.Month, &m.Income, &m.Expenses); err != nil {
			continue
		}
		m.NetCashFlow = m.Income - m.Expenses
		monthlyData = append(monthlyData, m)
		totalIncome += m.Income
		totalExpenses += m.Expenses
	}

	// Get spending by category (excluding income categories)
	catRows, err := db.DB.Query(`
		SELECT COALESCE(category, 'Uncategorized') as category, SUM(amount) as total
		FROM transactions
		WHERE user_id = ? AND date >= ? AND amount > 0
		AND category NOT IN ('INCOME', 'INCOME_WAGES', 'INCOME_DIVIDENDS', 'INCOME_INTEREST', 'TRANSFER_IN')
		AND (subcategory IS NULL OR (subcategory NOT LIKE 'INCOME%' AND subcategory NOT LIKE 'TRANSFER_IN%'))
		GROUP BY category
		ORDER BY total DESC
		LIMIT 10
	`, userID, startDate)
	if err == nil {
		defer catRows.Close()
	}

	categoryBreakdown := make(map[string]float64)
	if catRows != nil {
		for catRows.Next() {
			var cat string
			var total float64
			if catRows.Scan(&cat, &total) == nil {
				categoryBreakdown[cat] = total
			}
		}
	}

	avgMonthlyIncome := float64(0)
	avgMonthlyExpenses := float64(0)
	if len(monthlyData) > 0 {
		avgMonthlyIncome = totalIncome / float64(len(monthlyData))
		avgMonthlyExpenses = totalExpenses / float64(len(monthlyData))
	}

	savingsRate := float64(0)
	if totalIncome > 0 {
		savingsRate = (totalIncome - totalExpenses) / totalIncome * 100
	}

	result := map[string]interface{}{
		"monthly_data":         monthlyData,
		"total_income":         totalIncome,
		"total_expenses":       totalExpenses,
		"net_cash_flow":        totalIncome - totalExpenses,
		"avg_monthly_income":   avgMonthlyIncome,
		"avg_monthly_expenses": avgMonthlyExpenses,
		"savings_rate_percent": savingsRate,
		"category_breakdown":   categoryBreakdown,
		"period_months":        months,
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonBytes), nil
}

// runMonteCarlo runs a Monte Carlo simulation and saves it
func (e *ToolExecutor) runMonteCarlo(input map[string]interface{}) (string, error) {
	userID := e.GetEffectiveUserID()

	// Get user's current assets and debts
	assets, err := e.fetchAssets(userID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch assets: %w", err)
	}

	debts, err := e.fetchDebts(userID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch debts: %w", err)
	}

	// Build simulation parameters
	params := models.DefaultSimulationParams()

	// Required parameters
	if th, ok := input["time_horizon_years"].(float64); ok {
		params.TimeHorizonYears = int(th)
	} else {
		return "", fmt.Errorf("time_horizon_years is required")
	}

	if ca, ok := input["current_age"].(float64); ok {
		params.CurrentAge = int(ca)
	} else {
		return "", fmt.Errorf("current_age is required")
	}

	// Optional parameters with defaults
	if mc, ok := input["monthly_contribution"].(float64); ok {
		params.MonthlyContribution = mc
	}
	if ra, ok := input["retirement_age"].(float64); ok {
		params.RetirementAge = int(ra)
	}
	if rs, ok := input["retirement_spending"].(float64); ok {
		params.RetirementSpending = rs
	}
	if er, ok := input["expected_return"].(float64); ok {
		params.ExpectedReturn = er
	}
	if v, ok := input["volatility"].(float64); ok {
		params.Volatility = v
	}
	if ir, ok := input["inflation_rate"].(float64); ok {
		params.InflationRate = ir
	}
	if ss, ok := input["social_security_amount"].(float64); ok {
		params.SocialSecurityAmount = ss
	}
	if ssa, ok := input["social_security_age"].(float64); ok {
		params.SocialSecurityAge = int(ssa)
	}

	// Run the simulation
	result := simulation.RunMonteCarloWithParams(assets, debts, &params)

	// Always save simulation results
	paramsJSON, _ := json.Marshal(params)
	resultsJSON, _ := json.Marshal(result)

	var startingNW float64
	for _, a := range assets {
		startingNW += a.CurrentValue
	}
	for _, d := range debts {
		startingNW -= d.CurrentBalance
	}

	var name, notes *string
	if n, ok := input["name"].(string); ok && n != "" {
		name = &n
	}
	if n, ok := input["notes"].(string); ok && n != "" {
		notes = &n
	}

	runByUserID := e.UserID

	finalP50 := 0.0
	if len(result.Projections) > 0 {
		finalP50 = result.Projections[len(result.Projections)-1].P50
	}

	_, err = db.DB.Exec(`
		INSERT INTO simulation_history
		(user_id, run_by_user_id, name, notes, params, results, starting_net_worth, final_p50, success_rate, time_horizon_years)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, userID, runByUserID, name, notes, paramsJSON, resultsJSON,
		startingNW, finalP50, result.Summary.SuccessRate, params.TimeHorizonYears)

	if err != nil {
		// Log but don't fail - simulation still ran
		fmt.Printf("Warning: failed to save simulation: %v\n", err)
	}

	// Return simulation results
	output := map[string]interface{}{
		"success_rate":       result.Summary.SuccessRate,
		"starting_net_worth": startingNW,
		"projections":        result.Projections,
		"milestones":         result.Milestones,
		"insights":           result.Insights,
		"parameters": map[string]interface{}{
			"time_horizon_years":   params.TimeHorizonYears,
			"monthly_contribution": params.MonthlyContribution,
			"retirement_age":       params.RetirementAge,
			"current_age":          params.CurrentAge,
			"retirement_spending":  params.RetirementSpending,
			"expected_return":      params.ExpectedReturn,
			"volatility":           params.Volatility,
			"inflation_rate":       params.InflationRate,
		},
		"saved": true,
	}

	jsonBytes, _ := json.MarshalIndent(output, "", "  ")
	return string(jsonBytes), nil
}

// fetchAssets retrieves assets for Monte Carlo simulation
func (e *ToolExecutor) fetchAssets(userID int) ([]models.Asset, error) {
	rows, err := db.DB.Query(`
		SELECT a.id, a.user_id, a.name, a.type_id, a.current_value,
		       a.custom_return, a.custom_volatility, a.created_at, a.updated_at,
		       at.id, at.name, at.default_return, at.default_volatility
		FROM assets a
		LEFT JOIN asset_types at ON a.type_id = at.id
		WHERE a.user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []models.Asset
	for rows.Next() {
		var a models.Asset
		var at models.AssetType
		var customReturn, customVol *float64
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.TypeID, &a.CurrentValue,
			&customReturn, &customVol, &a.CreatedAt, &a.UpdatedAt,
			&at.ID, &at.Name, &at.DefaultReturn, &at.DefaultVolatility); err != nil {
			continue
		}
		a.CustomReturn = customReturn
		a.CustomVolatility = customVol
		a.AssetType = &at
		assets = append(assets, a)
	}
	return assets, nil
}

// fetchDebts retrieves debts for Monte Carlo simulation
func (e *ToolExecutor) fetchDebts(userID int) ([]models.Debt, error) {
	rows, err := db.DB.Query(`
		SELECT id, user_id, name, current_balance, interest_rate, minimum_payment, created_at, updated_at
		FROM debts
		WHERE user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var debts []models.Debt
	for rows.Next() {
		var d models.Debt
		var rate, minPay *float64
		if err := rows.Scan(&d.ID, &d.UserID, &d.Name, &d.CurrentBalance, &rate, &minPay, &d.CreatedAt, &d.UpdatedAt); err != nil {
			continue
		}
		d.InterestRate = rate
		d.MinimumPayment = minPay
		debts = append(debts, d)
	}
	return debts, nil
}

// getSimulationHistory retrieves past simulations
func (e *ToolExecutor) getSimulationHistory(input map[string]interface{}) (string, error) {
	userID := e.GetEffectiveUserID()

	limit := 10
	if l, ok := input["limit"].(float64); ok {
		limit = int(l)
	}

	query := `
		SELECT id, name, notes, starting_net_worth, final_p50, success_rate,
		       time_horizon_years, is_favorite, created_at
		FROM simulation_history
		WHERE user_id = ?
	`
	args := []interface{}{userID}

	if favOnly, ok := input["favorites_only"].(bool); ok && favOnly {
		query += " AND is_favorite = TRUE"
	}

	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type SimSummary struct {
		ID               int      `json:"id"`
		Name             *string  `json:"name,omitempty"`
		Notes            *string  `json:"notes,omitempty"`
		StartingNetWorth float64  `json:"starting_net_worth"`
		FinalP50         float64  `json:"final_p50"`
		SuccessRate      float64  `json:"success_rate"`
		TimeHorizonYears int      `json:"time_horizon_years"`
		IsFavorite       bool     `json:"is_favorite"`
		CreatedAt        string   `json:"created_at"`
	}

	var simulations []SimSummary
	for rows.Next() {
		var s SimSummary
		var createdAt time.Time
		if err := rows.Scan(&s.ID, &s.Name, &s.Notes, &s.StartingNetWorth,
			&s.FinalP50, &s.SuccessRate, &s.TimeHorizonYears, &s.IsFavorite, &createdAt); err != nil {
			continue
		}
		s.CreatedAt = createdAt.Format(time.RFC3339)
		simulations = append(simulations, s)
	}

	result := map[string]interface{}{
		"simulations": simulations,
		"count":       len(simulations),
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonBytes), nil
}

// getSimulationDetails retrieves full details of a simulation
func (e *ToolExecutor) getSimulationDetails(input map[string]interface{}) (string, error) {
	userID := e.GetEffectiveUserID()

	simID, ok := input["simulation_id"].(float64)
	if !ok {
		return "", fmt.Errorf("simulation_id is required")
	}

	var sim struct {
		ID               int
		Name             *string
		Notes            *string
		Params           []byte
		Results          []byte
		StartingNetWorth float64
		FinalP50         float64
		SuccessRate      float64
		TimeHorizonYears int
		IsFavorite       bool
		CreatedAt        time.Time
	}

	err := db.DB.QueryRow(`
		SELECT id, name, notes, params, results, starting_net_worth, final_p50,
		       success_rate, time_horizon_years, is_favorite, created_at
		FROM simulation_history
		WHERE id = ? AND user_id = ?
	`, int(simID), userID).Scan(&sim.ID, &sim.Name, &sim.Notes, &sim.Params, &sim.Results,
		&sim.StartingNetWorth, &sim.FinalP50, &sim.SuccessRate, &sim.TimeHorizonYears,
		&sim.IsFavorite, &sim.CreatedAt)

	if err != nil {
		return "", fmt.Errorf("simulation not found")
	}

	var params map[string]interface{}
	var results map[string]interface{}
	json.Unmarshal(sim.Params, &params)
	json.Unmarshal(sim.Results, &results)

	result := map[string]interface{}{
		"id":                 sim.ID,
		"name":               sim.Name,
		"notes":              sim.Notes,
		"params":             params,
		"results":            results,
		"starting_net_worth": sim.StartingNetWorth,
		"final_p50":          sim.FinalP50,
		"success_rate":       sim.SuccessRate,
		"time_horizon_years": sim.TimeHorizonYears,
		"is_favorite":        sim.IsFavorite,
		"created_at":         sim.CreatedAt.Format(time.RFC3339),
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonBytes), nil
}

// compareSimulations compares multiple simulations
func (e *ToolExecutor) compareSimulations(input map[string]interface{}) (string, error) {
	userID := e.GetEffectiveUserID()

	idsRaw, ok := input["simulation_ids"].([]interface{})
	if !ok || len(idsRaw) < 2 {
		return "", fmt.Errorf("at least 2 simulation_ids are required")
	}

	var ids []int
	for _, id := range idsRaw {
		if idFloat, ok := id.(float64); ok {
			ids = append(ids, int(idFloat))
		}
	}

	if len(ids) < 2 || len(ids) > 5 {
		return "", fmt.Errorf("must provide 2-5 simulation IDs")
	}

	type SimComparison struct {
		ID               int                    `json:"id"`
		Name             *string                `json:"name,omitempty"`
		Params           map[string]interface{} `json:"params"`
		SuccessRate      float64                `json:"success_rate"`
		StartingNetWorth float64                `json:"starting_net_worth"`
		FinalP50         float64                `json:"final_p50"`
		TimeHorizonYears int                    `json:"time_horizon_years"`
	}

	var simulations []SimComparison
	for _, id := range ids {
		var sim struct {
			ID               int
			Name             *string
			Params           []byte
			SuccessRate      float64
			StartingNetWorth float64
			FinalP50         float64
			TimeHorizonYears int
		}

		err := db.DB.QueryRow(`
			SELECT id, name, params, success_rate, starting_net_worth, final_p50, time_horizon_years
			FROM simulation_history
			WHERE id = ? AND user_id = ?
		`, id, userID).Scan(&sim.ID, &sim.Name, &sim.Params, &sim.SuccessRate,
			&sim.StartingNetWorth, &sim.FinalP50, &sim.TimeHorizonYears)

		if err != nil {
			continue
		}

		var params map[string]interface{}
		json.Unmarshal(sim.Params, &params)

		simulations = append(simulations, SimComparison{
			ID:               sim.ID,
			Name:             sim.Name,
			Params:           params,
			SuccessRate:      sim.SuccessRate,
			StartingNetWorth: sim.StartingNetWorth,
			FinalP50:         sim.FinalP50,
			TimeHorizonYears: sim.TimeHorizonYears,
		})
	}

	result := map[string]interface{}{
		"simulations": simulations,
		"count":       len(simulations),
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonBytes), nil
}

// runWhatIfAnalysis runs a what-if scenario comparison
func (e *ToolExecutor) runWhatIfAnalysis(input map[string]interface{}) (string, error) {
	userID := e.GetEffectiveUserID()

	question, _ := input["question"].(string)
	if question == "" {
		question = "What-If Analysis"
	}

	// Get user's current assets and debts
	assets, err := e.fetchAssets(userID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch assets: %w", err)
	}

	debts, err := e.fetchDebts(userID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch debts: %w", err)
	}

	// Calculate starting net worth
	var startingNW float64
	for _, a := range assets {
		startingNW += a.CurrentValue
	}
	for _, d := range debts {
		startingNW -= d.CurrentBalance
	}

	// Build baseline parameters - either from saved simulation or defaults
	baselineParams := models.DefaultSimulationParams()

	if baselineID, ok := input["baseline_simulation_id"].(float64); ok && baselineID > 0 {
		// Load baseline from saved simulation
		var paramsJSON []byte
		err := db.DB.QueryRow(`
			SELECT params FROM simulation_history WHERE id = ? AND user_id = ?
		`, int(baselineID), userID).Scan(&paramsJSON)
		if err == nil {
			json.Unmarshal(paramsJSON, &baselineParams)
		}
	} else {
		// Use provided current_age or default
		if ca, ok := input["current_age"].(float64); ok {
			baselineParams.CurrentAge = int(ca)
		}
	}

	// Override time horizon if provided
	if th, ok := input["time_horizon_years"].(float64); ok {
		baselineParams.TimeHorizonYears = int(th)
	}

	baselineParams.ApplyDefaults()

	// Create modified parameters (copy baseline)
	modifiedParams := baselineParams

	// Track what changed for the summary
	var changes []string

	// Apply changes
	if mc, ok := input["monthly_contribution_change"].(float64); ok && mc != 0 {
		modifiedParams.MonthlyContribution += mc
		if mc > 0 {
			changes = append(changes, fmt.Sprintf("Saving $%.0f more per month", mc))
		} else {
			changes = append(changes, fmt.Sprintf("Saving $%.0f less per month", -mc))
		}
	}

	if ra, ok := input["retirement_age_change"].(float64); ok && ra != 0 {
		modifiedParams.RetirementAge += int(ra)
		if ra < 0 {
			changes = append(changes, fmt.Sprintf("Retiring %d years earlier (age %d)", int(-ra), modifiedParams.RetirementAge))
		} else {
			changes = append(changes, fmt.Sprintf("Retiring %d years later (age %d)", int(ra), modifiedParams.RetirementAge))
		}
	}

	if rs, ok := input["retirement_spending_change"].(float64); ok && rs != 0 {
		modifiedParams.RetirementSpending += rs
		if rs > 0 {
			changes = append(changes, fmt.Sprintf("Spending $%.0f more per month in retirement", rs))
		} else {
			changes = append(changes, fmt.Sprintf("Spending $%.0f less per month in retirement", -rs))
		}
	}

	if ss, ok := input["social_security_amount"].(float64); ok {
		modifiedParams.SocialSecurityAmount = ss
		changes = append(changes, fmt.Sprintf("Social Security: $%.0f/month", ss))
	}

	if ssa, ok := input["social_security_age"].(float64); ok {
		modifiedParams.SocialSecurityAge = int(ssa)
		changes = append(changes, fmt.Sprintf("Taking Social Security at age %d", int(ssa)))
	}

	if er, ok := input["expected_return_change"].(float64); ok && er != 0 {
		modifiedParams.ExpectedReturn += er
		if er > 0 {
			changes = append(changes, fmt.Sprintf("Higher expected return (+%.1f%%)", er*100))
		} else {
			changes = append(changes, fmt.Sprintf("Lower expected return (%.1f%%)", er*100))
		}
	}

	if v, ok := input["volatility_change"].(float64); ok && v != 0 {
		modifiedParams.Volatility += v
		if v > 0 {
			changes = append(changes, fmt.Sprintf("Higher volatility (+%.1f%%)", v*100))
		} else {
			changes = append(changes, fmt.Sprintf("Lower volatility (%.1f%%)", v*100))
		}
	}

	// Run both simulations
	baselineResult := simulation.RunMonteCarloWithParams(assets, debts, &baselineParams)
	modifiedResult := simulation.RunMonteCarloWithParams(assets, debts, &modifiedParams)

	// Calculate differences
	successRateDiff := modifiedResult.Summary.SuccessRate - baselineResult.Summary.SuccessRate
	var finalP50Baseline, finalP50Modified float64
	if len(baselineResult.Projections) > 0 {
		finalP50Baseline = baselineResult.Projections[len(baselineResult.Projections)-1].P50
	}
	if len(modifiedResult.Projections) > 0 {
		finalP50Modified = modifiedResult.Projections[len(modifiedResult.Projections)-1].P50
	}
	finalWealthDiff := finalP50Modified - finalP50Baseline

	// Calculate total contributions difference
	baselineContribs := baselineParams.MonthlyContribution * 12 * float64(baselineParams.RetirementAge-baselineParams.CurrentAge)
	modifiedContribs := modifiedParams.MonthlyContribution * 12 * float64(modifiedParams.RetirementAge-modifiedParams.CurrentAge)
	contributionsDiff := modifiedContribs - baselineContribs

	// Generate natural language summary
	var summaryParts []string

	// Success rate impact
	if successRateDiff >= 5 {
		summaryParts = append(summaryParts, fmt.Sprintf("This change significantly improves your retirement success rate by %.1f percentage points (from %.1f%% to %.1f%%).", successRateDiff, baselineResult.Summary.SuccessRate, modifiedResult.Summary.SuccessRate))
	} else if successRateDiff >= 1 {
		summaryParts = append(summaryParts, fmt.Sprintf("This modestly improves your success rate by %.1f percentage points (%.1f%% to %.1f%%).", successRateDiff, baselineResult.Summary.SuccessRate, modifiedResult.Summary.SuccessRate))
	} else if successRateDiff <= -5 {
		summaryParts = append(summaryParts, fmt.Sprintf("Warning: This significantly reduces your success rate by %.1f percentage points (from %.1f%% to %.1f%%).", -successRateDiff, baselineResult.Summary.SuccessRate, modifiedResult.Summary.SuccessRate))
	} else if successRateDiff <= -1 {
		summaryParts = append(summaryParts, fmt.Sprintf("This slightly reduces your success rate by %.1f percentage points (%.1f%% to %.1f%%).", -successRateDiff, baselineResult.Summary.SuccessRate, modifiedResult.Summary.SuccessRate))
	} else {
		summaryParts = append(summaryParts, fmt.Sprintf("This has minimal impact on your success rate (remains around %.1f%%).", modifiedResult.Summary.SuccessRate))
	}

	// Wealth impact
	if finalWealthDiff > 100000 {
		summaryParts = append(summaryParts, fmt.Sprintf("Your median projected wealth at retirement increases by $%s.", formatLargeNumber(finalWealthDiff)))
	} else if finalWealthDiff < -100000 {
		summaryParts = append(summaryParts, fmt.Sprintf("Your median projected wealth at retirement decreases by $%s.", formatLargeNumber(-finalWealthDiff)))
	}

	// Contribution impact
	if contributionsDiff > 10000 {
		summaryParts = append(summaryParts, fmt.Sprintf("This requires $%s more in total contributions.", formatLargeNumber(contributionsDiff)))
	} else if contributionsDiff < -10000 {
		summaryParts = append(summaryParts, fmt.Sprintf("This saves you $%s in total contributions.", formatLargeNumber(-contributionsDiff)))
	}

	// Build recommendation
	var recommendation string
	if successRateDiff >= 5 && modifiedResult.Summary.SuccessRate >= 80 {
		recommendation = "This is a strong improvement. Consider implementing this change."
	} else if successRateDiff >= 0 && modifiedResult.Summary.SuccessRate >= 80 {
		recommendation = "Your plan remains strong with this change."
	} else if modifiedResult.Summary.SuccessRate >= 70 {
		recommendation = "Your plan is still viable but has room for improvement."
	} else if modifiedResult.Summary.SuccessRate >= 50 {
		recommendation = "This puts your plan at moderate risk. Consider adjustments."
	} else {
		recommendation = "This scenario has significant risk. Additional savings or a later retirement may be needed."
	}

	result := map[string]interface{}{
		"question": question,
		"changes":  changes,
		"baseline": map[string]interface{}{
			"success_rate":         baselineResult.Summary.SuccessRate,
			"final_p50":            finalP50Baseline,
			"monthly_contribution": baselineParams.MonthlyContribution,
			"retirement_age":       baselineParams.RetirementAge,
			"retirement_spending":  baselineParams.RetirementSpending,
			"total_contributions":  baselineContribs,
		},
		"modified": map[string]interface{}{
			"success_rate":         modifiedResult.Summary.SuccessRate,
			"final_p50":            finalP50Modified,
			"monthly_contribution": modifiedParams.MonthlyContribution,
			"retirement_age":       modifiedParams.RetirementAge,
			"retirement_spending":  modifiedParams.RetirementSpending,
			"total_contributions":  modifiedContribs,
		},
		"impact": map[string]interface{}{
			"success_rate_change":   successRateDiff,
			"final_wealth_change":   finalWealthDiff,
			"contributions_change":  contributionsDiff,
			"summary":               summaryParts,
			"recommendation":        recommendation,
		},
		"starting_net_worth": startingNW,
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonBytes), nil
}

// formatLargeNumber formats a large number with K/M suffix
func formatLargeNumber(val float64) string {
	if val >= 1000000 {
		return fmt.Sprintf("%.1fM", val/1000000)
	}
	if val >= 1000 {
		return fmt.Sprintf("%.0fK", val/1000)
	}
	return fmt.Sprintf("%.0f", val)
}

// generateReport generates a PDF financial plan report
func (e *ToolExecutor) generateReport(input map[string]interface{}) (string, error) {
	userID := e.GetEffectiveUserID()

	// Get user info
	var userName, userEmail string
	db.DB.QueryRow(`SELECT name, email FROM users WHERE id = ?`, userID).Scan(&userName, &userEmail)

	// Determine advisor name if in client context
	var advisorName string
	if e.ClientContext != 0 {
		db.DB.QueryRow(`SELECT name FROM users WHERE id = ?`, e.UserID).Scan(&advisorName)
	}

	// Get assets with types
	assets, err := e.fetchAssets(userID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch assets: %w", err)
	}

	// Get debts
	debts, err := e.fetchDebts(userID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch debts: %w", err)
	}

	// Calculate totals
	var totalAssets, totalDebts float64
	for _, a := range assets {
		totalAssets += a.CurrentValue
	}
	for _, d := range debts {
		totalDebts += d.CurrentBalance
	}
	netWorth := totalAssets - totalDebts

	// Prepare report data
	reportData := reports.ReportData{
		ClientName:  userName,
		AdvisorName: advisorName,
		GeneratedAt: time.Now(),
		Assets:      assets,
		Debts:       debts,
		TotalAssets: totalAssets,
		TotalDebts:  totalDebts,
		NetWorth:    netWorth,
	}

	// Check if simulation should be included
	includeSim := true
	if incl, ok := input["include_simulation"].(bool); ok {
		includeSim = incl
	}

	if includeSim {
		params := models.DefaultSimulationParams()

		if th, ok := input["time_horizon_years"].(float64); ok {
			params.TimeHorizonYears = int(th)
		}
		if mc, ok := input["monthly_contribution"].(float64); ok {
			params.MonthlyContribution = mc
		}
		if ra, ok := input["retirement_age"].(float64); ok {
			params.RetirementAge = int(ra)
		}
		if ca, ok := input["current_age"].(float64); ok {
			params.CurrentAge = int(ca)
		}
		if rs, ok := input["retirement_spending"].(float64); ok {
			params.RetirementSpending = rs
		}

		params.ApplyDefaults()
		simResult := simulation.RunMonteCarloWithParams(assets, debts, &params)
		reportData.Simulation = &simResult
		reportData.Params = &params
	}

	// Generate PDF
	pdfBytes, err := reports.GenerateFinancialPlanReport(reportData)
	if err != nil {
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	// Generate filename
	filename := fmt.Sprintf("financial_plan_%s_%s.pdf",
		sanitizeFilename(userName),
		time.Now().Format("2006-01-02"))

	// Check if we should save to document vault (default: true)
	saveToVault := true
	if stv, ok := input["save_to_vault"].(bool); ok {
		saveToVault = stv
	}

	var docID int64
	var downloadURL string

	if saveToVault && storage.DefaultStorage != nil {
		// Save to document vault
		storagePath, err := storage.DefaultStorage.Save(pdfBytes, filename, true)
		if err == nil {
			// Save document record
			result, err := db.DB.Exec(`
				INSERT INTO documents (user_id, uploaded_by, name, original_name, mime_type, size, category, storage_path, encrypted)
				VALUES (?, ?, ?, ?, 'application/pdf', ?, 'reports', ?, TRUE)
			`, userID, e.UserID, filename, filename, len(pdfBytes), storagePath)

			if err == nil {
				docID, _ = result.LastInsertId()
				downloadURL = fmt.Sprintf("/api/documents/%d/download", docID)

				// If advisor generated for client, auto-share with client
				if e.ClientContext != 0 && e.ClientContext != e.UserID {
					db.DB.Exec(`
						INSERT INTO document_shares (document_id, shared_with_id, shared_by_id, permission)
						VALUES (?, ?, ?, 'download')
					`, docID, userID, e.UserID)
				}
			}
		}
	}

	// Build response
	result := map[string]interface{}{
		"status":   "success",
		"filename": filename,
		"size":     len(pdfBytes),
		"message":  fmt.Sprintf("Generated financial plan report for %s (%d pages, %d KB)", userName, estimatePages(len(pdfBytes)), len(pdfBytes)/1024),
	}

	// Include document vault info if saved
	if docID > 0 {
		result["document_id"] = docID
		result["download_url"] = downloadURL
		result["saved_to_vault"] = true
	}

	// Also include base64 for immediate display/download
	pdfBase64 := base64.StdEncoding.EncodeToString(pdfBytes)
	result["pdf_data"] = pdfBase64

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonBytes), nil
}

// sanitizeFilename removes/replaces characters that are unsafe for filenames
func sanitizeFilename(name string) string {
	result := make([]byte, 0, len(name))
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			result = append(result, byte(c))
		} else if c == ' ' {
			result = append(result, '_')
		}
	}
	if len(result) == 0 {
		return "report"
	}
	return string(result)
}

// estimatePages estimates the number of pages based on file size
func estimatePages(size int) int {
	// Rough estimate: ~50KB per page for a PDF with text and graphics
	pages := size / 50000
	if pages < 1 {
		pages = 1
	}
	return pages
}

// listClients lists all clients for an advisor
func (e *ToolExecutor) listClients() (string, error) {
	if !e.IsAdvisor {
		return "", fmt.Errorf("this tool is only available to advisors")
	}

	rows, err := db.DB.Query(`
		SELECT u.id, u.email, u.name, ac.status, ac.access_level, ac.created_at,
		       COALESCE(SUM(a.current_value), 0) as total_assets,
		       COALESCE(SUM(d.current_balance), 0) as total_debts
		FROM advisor_clients ac
		JOIN users u ON ac.client_id = u.id
		LEFT JOIN assets a ON u.id = a.user_id
		LEFT JOIN debts d ON u.id = d.user_id
		WHERE ac.advisor_id = ? AND ac.status = 'active'
		GROUP BY u.id, u.email, u.name, ac.status, ac.access_level, ac.created_at
		ORDER BY u.name
	`, e.UserID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type ClientSummary struct {
		ID          int     `json:"id"`
		Email       string  `json:"email"`
		Name        string  `json:"name"`
		Status      string  `json:"status"`
		AccessLevel string  `json:"access_level"`
		TotalAssets float64 `json:"total_assets"`
		TotalDebts  float64 `json:"total_debts"`
		NetWorth    float64 `json:"net_worth"`
		AddedAt     string  `json:"added_at"`
	}

	var clients []ClientSummary
	for rows.Next() {
		var c ClientSummary
		var addedAt time.Time
		if err := rows.Scan(&c.ID, &c.Email, &c.Name, &c.Status, &c.AccessLevel,
			&addedAt, &c.TotalAssets, &c.TotalDebts); err != nil {
			continue
		}
		c.NetWorth = c.TotalAssets - c.TotalDebts
		c.AddedAt = addedAt.Format(time.RFC3339)
		clients = append(clients, c)
	}

	result := map[string]interface{}{
		"clients": clients,
		"count":   len(clients),
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonBytes), nil
}

// switchClientContext switches to a specific client
func (e *ToolExecutor) switchClientContext(input map[string]interface{}) (string, error) {
	if !e.IsAdvisor {
		return "", fmt.Errorf("this tool is only available to advisors")
	}

	clientID, ok := input["client_id"].(float64)
	if !ok {
		return "", fmt.Errorf("client_id is required")
	}

	id := int(clientID)

	// Switch back to advisor's own view
	if id == 0 {
		e.ClientContext = 0
		return `{"status": "switched", "context": "advisor", "message": "Switched back to your own view"}`, nil
	}

	// Verify advisor has access to this client
	var accessLevel string
	err := db.DB.QueryRow(`
		SELECT access_level FROM advisor_clients
		WHERE advisor_id = ? AND client_id = ? AND status = 'active'
	`, e.UserID, id).Scan(&accessLevel)

	if err != nil {
		return "", fmt.Errorf("client not found or access denied")
	}

	// Get client details
	var clientName, clientEmail string
	db.DB.QueryRow(`SELECT name, email FROM users WHERE id = ?`, id).Scan(&clientName, &clientEmail)

	e.ClientContext = id

	result := map[string]interface{}{
		"status":       "switched",
		"context":      "client",
		"client_id":    id,
		"client_name":  clientName,
		"client_email": clientEmail,
		"access_level": accessLevel,
		"message":      fmt.Sprintf("Now working with client: %s", clientName),
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonBytes), nil
}

// optimizeSocialSecurity analyzes Social Security claiming strategies
func (e *ToolExecutor) optimizeSocialSecurity(input map[string]interface{}) (string, error) {
	// Parse birth date
	birthDateStr, ok := input["birth_date"].(string)
	if !ok || birthDateStr == "" {
		return "", fmt.Errorf("birth_date is required (format: YYYY-MM-DD)")
	}

	birthDate, err := time.Parse("2006-01-02", birthDateStr)
	if err != nil {
		return "", fmt.Errorf("invalid birth_date format, use YYYY-MM-DD")
	}

	// Calculate current age and full retirement age (FRA)
	now := time.Now()
	currentAge := now.Year() - birthDate.Year()
	if now.YearDay() < birthDate.YearDay() {
		currentAge--
	}

	// Determine FRA based on birth year (simplified - 67 for those born 1960+)
	birthYear := birthDate.Year()
	var fraAge int
	var fraMonths int
	switch {
	case birthYear <= 1937:
		fraAge, fraMonths = 65, 0
	case birthYear == 1938:
		fraAge, fraMonths = 65, 2
	case birthYear == 1939:
		fraAge, fraMonths = 65, 4
	case birthYear == 1940:
		fraAge, fraMonths = 65, 6
	case birthYear == 1941:
		fraAge, fraMonths = 65, 8
	case birthYear == 1942:
		fraAge, fraMonths = 65, 10
	case birthYear >= 1943 && birthYear <= 1954:
		fraAge, fraMonths = 66, 0
	case birthYear == 1955:
		fraAge, fraMonths = 66, 2
	case birthYear == 1956:
		fraAge, fraMonths = 66, 4
	case birthYear == 1957:
		fraAge, fraMonths = 66, 6
	case birthYear == 1958:
		fraAge, fraMonths = 66, 8
	case birthYear == 1959:
		fraAge, fraMonths = 66, 10
	default: // 1960+
		fraAge, fraMonths = 67, 0
	}

	// Get PIA (Primary Insurance Amount at FRA)
	var pia float64
	if estimatedPia, ok := input["estimated_pia"].(float64); ok && estimatedPia > 0 {
		pia = estimatedPia
	} else if annualEarnings, ok := input["current_annual_earnings"].(float64); ok && annualEarnings > 0 {
		// Rough PIA estimate: ~42% of average indexed monthly earnings for typical earner
		// This is a simplification; real calculation uses AIME and bend points
		aime := annualEarnings / 12
		// 2024 bend points approximation
		if aime <= 1174 {
			pia = aime * 0.90
		} else if aime <= 7078 {
			pia = 1174*0.90 + (aime-1174)*0.32
		} else {
			pia = 1174*0.90 + (7078-1174)*0.32 + (aime-7078)*0.15
		}
	} else {
		return "", fmt.Errorf("either estimated_pia or current_annual_earnings is required")
	}

	// Life expectancy (default to 85 if not provided)
	lifeExpectancy := 85
	if le, ok := input["life_expectancy_years"].(float64); ok && le > 0 {
		lifeExpectancy = int(le)
	}

	// Calculate benefit at each claiming age (62-70)
	type ClaimingStrategy struct {
		ClaimAge        int     `json:"claim_age"`
		MonthlyBenefit  float64 `json:"monthly_benefit"`
		AnnualBenefit   float64 `json:"annual_benefit"`
		LifetimeBenefit float64 `json:"lifetime_benefit"`
		BreakevenVs62   int     `json:"breakeven_vs_62,omitempty"` // Age when this strategy beats age 62
		Adjustment      string  `json:"adjustment"`
	}

	var strategies []ClaimingStrategy
	var benefit62 float64

	for claimAge := 62; claimAge <= 70; claimAge++ {
		// Calculate months before/after FRA
		fraAgeDecimal := float64(fraAge) + float64(fraMonths)/12.0
		monthsFromFRA := (float64(claimAge) - fraAgeDecimal) * 12

		var adjustmentFactor float64
		var adjustmentDesc string

		if monthsFromFRA < 0 {
			// Early claiming reduction
			monthsEarly := -monthsFromFRA
			if monthsEarly <= 36 {
				// 5/9 of 1% per month for first 36 months
				adjustmentFactor = 1 - (monthsEarly * 5 / 9 / 100)
			} else {
				// 5/12 of 1% for months beyond 36
				adjustmentFactor = 1 - (36*5/9/100 + (monthsEarly-36)*5/12/100)
			}
			adjustmentDesc = fmt.Sprintf("%.1f%% reduction (%.0f months early)", (1-adjustmentFactor)*100, monthsEarly)
		} else if monthsFromFRA > 0 {
			// Delayed retirement credits (8% per year / 2/3% per month)
			monthsLate := monthsFromFRA
			adjustmentFactor = 1 + (monthsLate * 2 / 3 / 100)
			adjustmentDesc = fmt.Sprintf("+%.1f%% credits (%.0f months delayed)", (adjustmentFactor-1)*100, monthsLate)
		} else {
			adjustmentFactor = 1.0
			adjustmentDesc = "Full retirement age (100%)"
		}

		monthlyBenefit := pia * adjustmentFactor
		annualBenefit := monthlyBenefit * 12
		yearsReceiving := lifeExpectancy - claimAge
		if yearsReceiving < 0 {
			yearsReceiving = 0
		}
		lifetimeBenefit := annualBenefit * float64(yearsReceiving)

		if claimAge == 62 {
			benefit62 = monthlyBenefit
		}

		strategy := ClaimingStrategy{
			ClaimAge:        claimAge,
			MonthlyBenefit:  math.Round(monthlyBenefit*100) / 100,
			AnnualBenefit:   math.Round(annualBenefit*100) / 100,
			LifetimeBenefit: math.Round(lifetimeBenefit*100) / 100,
			Adjustment:      adjustmentDesc,
		}

		// Calculate breakeven age vs claiming at 62
		if claimAge > 62 && benefit62 > 0 {
			// Find age where cumulative benefits equal
			// At age X: (X-62)*benefit62 = (X-claimAge)*monthlyBenefit*12
			// Solving: X = (claimAge*monthlyBenefit*12 - 62*benefit62*12) / (monthlyBenefit*12 - benefit62*12)
			annual62 := benefit62 * 12
			if annualBenefit > annual62 {
				breakeven := (float64(claimAge)*annualBenefit - 62*annual62) / (annualBenefit - annual62)
				strategy.BreakevenVs62 = int(math.Ceil(breakeven))
			}
		}

		strategies = append(strategies, strategy)
	}

	// Find optimal strategy based on lifetime benefits
	var optimalAge int
	var maxLifetime float64
	for _, s := range strategies {
		if s.LifetimeBenefit > maxLifetime {
			maxLifetime = s.LifetimeBenefit
			optimalAge = s.ClaimAge
		}
	}

	// Generate insights
	var insights []string
	insights = append(insights, fmt.Sprintf("Based on life expectancy of %d, claiming at age %d maximizes lifetime benefits", lifeExpectancy, optimalAge))

	if lifeExpectancy >= 85 {
		insights = append(insights, "With longer life expectancy, delaying benefits typically pays off")
	} else if lifeExpectancy <= 78 {
		insights = append(insights, "With shorter life expectancy, earlier claiming may be more beneficial")
	}

	// Check if health or cash needs mentioned
	if currentAge >= 62 {
		insights = append(insights, "You're currently eligible to claim - consider health status and income needs")
	}

	result := map[string]interface{}{
		"birth_date":       birthDateStr,
		"current_age":      currentAge,
		"full_retirement_age": fmt.Sprintf("%d years, %d months", fraAge, fraMonths),
		"estimated_pia":    math.Round(pia*100) / 100,
		"life_expectancy":  lifeExpectancy,
		"claiming_strategies": strategies,
		"optimal_strategy": map[string]interface{}{
			"claim_age":        optimalAge,
			"lifetime_benefit": maxLifetime,
			"reasoning":        fmt.Sprintf("Maximizes total benefits assuming life expectancy of %d", lifeExpectancy),
		},
		"insights": insights,
		"disclaimer": "This analysis is for educational purposes. Social Security rules are complex - consult SSA.gov or a financial advisor for personalized guidance.",
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonBytes), nil
}

// projectTaxLiability estimates current year tax liability
func (e *ToolExecutor) projectTaxLiability(input map[string]interface{}) (string, error) {
	userID := e.GetEffectiveUserID()
	now := time.Now()
	currentYear := now.Year()

	// Get filing status (default to single)
	filingStatus := "single"
	if fs, ok := input["filing_status"].(string); ok {
		filingStatus = fs
	}

	// 2024 Tax Brackets (will be updated annually)
	type TaxBracket struct {
		Rate   float64
		Single float64
		MFJ    float64 // Married Filing Jointly
	}

	brackets := []TaxBracket{
		{0.10, 11600, 23200},
		{0.12, 47150, 94300},
		{0.22, 100525, 201050},
		{0.24, 191950, 383900},
		{0.32, 243725, 487450},
		{0.35, 609350, 731200},
		{0.37, math.MaxFloat64, math.MaxFloat64},
	}

	// Standard deductions
	standardDeductions := map[string]float64{
		"single":                  14600,
		"married_filing_jointly":  29200,
		"married_filing_separate": 14600,
		"head_of_household":       21900,
	}

	standardDeduction := standardDeductions["single"]
	if sd, ok := standardDeductions[filingStatus]; ok {
		standardDeduction = sd
	}

	// Get annual income - either from input or estimate from transactions
	var annualIncome float64
	var incomeSource string

	if ai, ok := input["annual_income"].(float64); ok && ai > 0 {
		annualIncome = ai
		incomeSource = "user_provided"
	} else {
		// Estimate from YTD transactions
		ytdStart := fmt.Sprintf("%d-01-01", currentYear)
		ytdEnd := now.Format("2006-01-02")

		var ytdIncome float64
		err := db.DB.QueryRow(`
			SELECT COALESCE(SUM(ABS(amount)), 0)
			FROM transactions
			WHERE user_id = ? AND date >= ? AND date <= ?
			AND (amount < 0 OR category IN ('INCOME', 'INCOME_WAGES', 'INCOME_DIVIDENDS', 'INCOME_INTEREST'))
		`, userID, ytdStart, ytdEnd).Scan(&ytdIncome)

		if err != nil || ytdIncome == 0 {
			return `{"error": "No income data found", "suggestion": "Provide annual_income parameter or connect bank accounts"}`, nil
		}

		// Extrapolate to full year
		dayOfYear := now.YearDay()
		annualIncome = ytdIncome * 365.0 / float64(dayOfYear)
		incomeSource = fmt.Sprintf("estimated_from_ytd (%d days of data)", dayOfYear)
	}

	// Get itemized deductions if provided, otherwise use standard
	useStandardDeduction := true
	totalDeductions := standardDeduction

	if itemized, ok := input["itemized_deductions"].(float64); ok && itemized > standardDeduction {
		totalDeductions = itemized
		useStandardDeduction = false
	}

	// Calculate taxable income
	taxableIncome := annualIncome - totalDeductions
	if taxableIncome < 0 {
		taxableIncome = 0
	}

	// Calculate federal tax
	var federalTax float64
	var marginalRate float64
	remainingIncome := taxableIncome
	previousBracketMax := 0.0

	type BracketBreakdown struct {
		Rate      float64 `json:"rate"`
		RangeLow  float64 `json:"range_low"`
		RangeHigh float64 `json:"range_high"`
		TaxableInRange float64 `json:"taxable_in_range"`
		TaxInRange     float64 `json:"tax_in_range"`
	}

	var bracketBreakdown []BracketBreakdown

	for _, bracket := range brackets {
		bracketMax := bracket.Single
		if filingStatus == "married_filing_jointly" {
			bracketMax = bracket.MFJ
		}

		bracketRange := bracketMax - previousBracketMax
		taxableInBracket := math.Min(remainingIncome, bracketRange)

		if taxableInBracket > 0 {
			taxInBracket := taxableInBracket * bracket.Rate
			federalTax += taxInBracket
			marginalRate = bracket.Rate

			bracketBreakdown = append(bracketBreakdown, BracketBreakdown{
				Rate:           bracket.Rate * 100,
				RangeLow:       previousBracketMax,
				RangeHigh:      bracketMax,
				TaxableInRange: math.Round(taxableInBracket*100) / 100,
				TaxInRange:     math.Round(taxInBracket*100) / 100,
			})

			remainingIncome -= taxableInBracket
		}

		if remainingIncome <= 0 {
			break
		}

		previousBracketMax = bracketMax
	}

	effectiveRate := 0.0
	if annualIncome > 0 {
		effectiveRate = federalTax / annualIncome * 100
	}

	// Get withholdings if provided
	var estimatedRefund float64
	var withholdingStatus string
	if withholdings, ok := input["ytd_withholdings"].(float64); ok && withholdings > 0 {
		// Extrapolate withholdings to full year
		dayOfYear := now.YearDay()
		annualWithholdings := withholdings * 365.0 / float64(dayOfYear)
		estimatedRefund = annualWithholdings - federalTax
		withholdingStatus = "projected"
	} else {
		withholdingStatus = "not_provided"
	}

	// Generate optimization suggestions
	var suggestions []string

	// 401k suggestion
	limit401k := 23000.0 // 2024 limit (catchup for 50+ is +$7500)
	if marginalRate >= 0.22 {
		taxSavings := limit401k * marginalRate
		suggestions = append(suggestions, fmt.Sprintf("Maximize 401(k): Contributing $%.0f saves ~$%.0f in taxes at your %.0f%% marginal rate", limit401k, taxSavings, marginalRate*100))
	}

	// Traditional IRA suggestion
	if annualIncome < 87000 { // Rough phase-out check
		iraLimit := 7000.0
		iraSavings := iraLimit * marginalRate
		suggestions = append(suggestions, fmt.Sprintf("Traditional IRA contribution of $%.0f could save ~$%.0f", iraLimit, iraSavings))
	}

	// HSA suggestion
	hsaLimit := 4150.0 // 2024 single
	if filingStatus == "married_filing_jointly" {
		hsaLimit = 8300.0
	}
	hsaSavings := hsaLimit * marginalRate
	suggestions = append(suggestions, fmt.Sprintf("HSA contribution of $%.0f saves $%.0f (triple tax advantage)", hsaLimit, hsaSavings))

	// Roth conversion suggestion if in low bracket
	if marginalRate <= 0.22 && taxableIncome > 0 {
		remainingIn22 := 0.0
		if filingStatus == "married_filing_jointly" {
			remainingIn22 = 201050 - taxableIncome
		} else {
			remainingIn22 = 100525 - taxableIncome
		}
		if remainingIn22 > 0 {
			suggestions = append(suggestions, fmt.Sprintf("Consider Roth conversion of up to $%.0f while in %.0f%% bracket", remainingIn22, marginalRate*100))
		}
	}

	// Charitable contribution suggestion if itemizing
	if !useStandardDeduction {
		suggestions = append(suggestions, "Consider bunching charitable contributions or using donor-advised fund")
	}

	result := map[string]interface{}{
		"tax_year": currentYear,
		"filing_status": filingStatus,
		"income": map[string]interface{}{
			"annual_income": math.Round(annualIncome*100) / 100,
			"source":        incomeSource,
		},
		"deductions": map[string]interface{}{
			"type":           map[bool]string{true: "standard", false: "itemized"}[useStandardDeduction],
			"amount":         totalDeductions,
			"standard_deduction": standardDeduction,
		},
		"tax_calculation": map[string]interface{}{
			"taxable_income":   math.Round(taxableIncome*100) / 100,
			"federal_tax":      math.Round(federalTax*100) / 100,
			"marginal_rate":    marginalRate * 100,
			"effective_rate":   math.Round(effectiveRate*100) / 100,
			"bracket_breakdown": bracketBreakdown,
		},
		"withholding_analysis": map[string]interface{}{
			"status":           withholdingStatus,
			"estimated_refund": math.Round(estimatedRefund*100) / 100,
		},
		"optimization_suggestions": suggestions,
		"disclaimer": "This is an estimate for planning purposes. Consult a tax professional for actual tax preparation.",
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonBytes), nil
}

// checkPortfolioDrift analyzes asset allocation drift and recommends rebalancing
func (e *ToolExecutor) checkPortfolioDrift(input map[string]interface{}) (string, error) {
	userID := e.GetEffectiveUserID()

	// Get drift threshold (default 5%)
	driftThreshold := 5.0
	if dt, ok := input["drift_threshold"].(float64); ok && dt > 0 {
		driftThreshold = dt
	}

	// Get target allocation from input
	targetAllocation := make(map[string]float64)
	if ta, ok := input["target_allocation"].(map[string]interface{}); ok {
		for k, v := range ta {
			if val, ok := v.(float64); ok {
				targetAllocation[k] = val
			}
		}
	}

	// If no target provided, use age-based defaults
	if len(targetAllocation) == 0 {
		// Try to get user's age from input or use default allocation
		age := 35
		if a, ok := input["age"].(float64); ok {
			age = int(a)
		}

		// Classic "100 minus age" rule for stocks
		stockAllocation := 100 - age
		if stockAllocation < 20 {
			stockAllocation = 20
		}
		if stockAllocation > 90 {
			stockAllocation = 90
		}

		targetAllocation = map[string]float64{
			"Stocks":      float64(stockAllocation),
			"Bonds":       float64(100-stockAllocation) * 0.7,
			"Cash":        float64(100-stockAllocation) * 0.3,
		}
	}

	// Verify target sums to 100
	totalTarget := 0.0
	for _, v := range targetAllocation {
		totalTarget += v
	}
	if totalTarget < 99 || totalTarget > 101 {
		return "", fmt.Errorf("target_allocation must sum to 100%% (got %.1f%%)", totalTarget)
	}

	// Get current assets by type
	rows, err := db.DB.Query(`
		SELECT COALESCE(at.name, 'Other') as type_name, SUM(a.current_value) as total
		FROM assets a
		LEFT JOIN asset_types at ON a.type_id = at.id
		WHERE a.user_id = ?
		GROUP BY at.name
	`, userID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch assets: %w", err)
	}
	defer rows.Close()

	currentAllocation := make(map[string]float64)
	var totalValue float64

	for rows.Next() {
		var typeName string
		var total float64
		if err := rows.Scan(&typeName, &total); err != nil {
			continue
		}
		currentAllocation[typeName] = total
		totalValue += total
	}

	if totalValue == 0 {
		return `{"error": "No assets found", "suggestion": "Add assets to your portfolio first"}`, nil
	}

	// Map asset types to allocation categories
	typeToCategory := map[string]string{
		"Stocks":             "Stocks",
		"Stock":              "Stocks",
		"ETF":                "Stocks",
		"Mutual Fund":        "Stocks",
		"Index Fund":         "Stocks",
		"Equity":             "Stocks",
		"Bonds":              "Bonds",
		"Bond":               "Bonds",
		"Fixed Income":       "Bonds",
		"Treasury":           "Bonds",
		"Cash":               "Cash",
		"Savings":            "Cash",
		"Money Market":       "Cash",
		"Checking":           "Cash",
		"Real Estate":        "Real Estate",
		"REIT":               "Real Estate",
		"Property":           "Real Estate",
		"Cryptocurrency":     "Alternative",
		"Crypto":             "Alternative",
		"Commodities":        "Alternative",
		"Gold":               "Alternative",
		"Alternative":        "Alternative",
		"401(k)":             "Retirement", // These will be broken down further if possible
		"IRA":                "Retirement",
		"Roth IRA":           "Retirement",
	}

	// Aggregate by category
	categoryTotals := make(map[string]float64)
	for typeName, value := range currentAllocation {
		category := typeToCategory[typeName]
		if category == "" {
			category = "Other"
		}
		categoryTotals[category] += value
	}

	// Calculate current percentages and drift
	type AllocationDrift struct {
		Category       string  `json:"category"`
		CurrentValue   float64 `json:"current_value"`
		CurrentPercent float64 `json:"current_percent"`
		TargetPercent  float64 `json:"target_percent"`
		DriftPercent   float64 `json:"drift_percent"`
		DriftValue     float64 `json:"drift_value"`
		Action         string  `json:"action"`
	}

	var drifts []AllocationDrift
	var totalDrift float64
	needsRebalancing := false

	// Check each target category
	for category, targetPct := range targetAllocation {
		currentValue := categoryTotals[category]
		currentPct := currentValue / totalValue * 100
		drift := currentPct - targetPct
		driftValue := (drift / 100) * totalValue

		action := "Hold"
		if drift > driftThreshold {
			action = fmt.Sprintf("Sell $%.0f", abs(driftValue))
			needsRebalancing = true
		} else if drift < -driftThreshold {
			action = fmt.Sprintf("Buy $%.0f", abs(driftValue))
			needsRebalancing = true
		}

		drifts = append(drifts, AllocationDrift{
			Category:       category,
			CurrentValue:   math.Round(currentValue*100) / 100,
			CurrentPercent: math.Round(currentPct*100) / 100,
			TargetPercent:  targetPct,
			DriftPercent:   math.Round(drift*100) / 100,
			DriftValue:     math.Round(driftValue*100) / 100,
			Action:         action,
		})

		totalDrift += abs(drift)
	}

	// Check for categories in portfolio but not in target
	for category, value := range categoryTotals {
		if _, exists := targetAllocation[category]; !exists && value > 0 {
			pct := value / totalValue * 100
			drifts = append(drifts, AllocationDrift{
				Category:       category,
				CurrentValue:   math.Round(value*100) / 100,
				CurrentPercent: math.Round(pct*100) / 100,
				TargetPercent:  0,
				DriftPercent:   math.Round(pct*100) / 100,
				DriftValue:     math.Round(value*100) / 100,
				Action:         fmt.Sprintf("Consider reallocation ($%.0f)", value),
			})
		}
	}

	// Sort by absolute drift
	for i := 0; i < len(drifts); i++ {
		for j := i + 1; j < len(drifts); j++ {
			if abs(drifts[j].DriftPercent) > abs(drifts[i].DriftPercent) {
				drifts[i], drifts[j] = drifts[j], drifts[i]
			}
		}
	}

	// Generate rebalancing trades
	type RebalanceTrade struct {
		Action   string  `json:"action"`
		Category string  `json:"category"`
		Amount   float64 `json:"amount"`
		Reason   string  `json:"reason"`
	}

	var trades []RebalanceTrade
	if needsRebalancing {
		for _, d := range drifts {
			if d.DriftPercent > driftThreshold {
				trades = append(trades, RebalanceTrade{
					Action:   "SELL",
					Category: d.Category,
					Amount:   abs(d.DriftValue),
					Reason:   fmt.Sprintf("Over target by %.1f%%", d.DriftPercent),
				})
			} else if d.DriftPercent < -driftThreshold {
				trades = append(trades, RebalanceTrade{
					Action:   "BUY",
					Category: d.Category,
					Amount:   abs(d.DriftValue),
					Reason:   fmt.Sprintf("Under target by %.1f%%", -d.DriftPercent),
				})
			}
		}
	}

	// Generate insights
	var insights []string

	if needsRebalancing {
		insights = append(insights, fmt.Sprintf("Portfolio drift exceeds %.0f%% threshold - rebalancing recommended", driftThreshold))
	} else {
		insights = append(insights, "Portfolio is within acceptable drift range - no immediate action needed")
	}

	// Check for concentration risk
	for _, d := range drifts {
		if d.CurrentPercent > 50 {
			insights = append(insights, fmt.Sprintf("High concentration in %s (%.1f%%) - consider diversification", d.Category, d.CurrentPercent))
			break
		}
	}

	// Tax-aware note
	if needsRebalancing {
		insights = append(insights, "Consider tax implications before selling - prioritize rebalancing in tax-advantaged accounts")
	}

	result := map[string]interface{}{
		"portfolio_value":      math.Round(totalValue*100) / 100,
		"drift_threshold":      driftThreshold,
		"needs_rebalancing":    needsRebalancing,
		"total_drift":          math.Round(totalDrift*100) / 100,
		"allocation_analysis":  drifts,
		"target_allocation":    targetAllocation,
		"recommended_trades":   trades,
		"insights":             insights,
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonBytes), nil
}

// analyzeSpendingPatterns provides behavioral insights from transaction history
func (e *ToolExecutor) analyzeSpendingPatterns(input map[string]interface{}) (string, error) {
	userID := e.GetEffectiveUserID()

	// Get number of months to analyze (default 6)
	months := 6
	if m, ok := input["months"].(float64); ok && m > 0 {
		months = int(m)
		if months > 24 {
			months = 24
		}
	}

	compareToPrior := false
	if ctp, ok := input["compare_to_prior"].(bool); ok {
		compareToPrior = ctp
	}

	// Calculate date ranges
	now := time.Now()
	endDate := now.Format("2006-01-02")
	startDate := now.AddDate(0, -months, 0).Format("2006-01-02")

	// Get transactions for current period
	rows, err := db.DB.Query(`
		SELECT id, name, amount, date, category, subcategory, merchant_name
		FROM transactions
		WHERE user_id = ? AND date >= ? AND date <= ?
		ORDER BY date DESC
	`, userID, startDate, endDate)
	if err != nil {
		return "", fmt.Errorf("failed to fetch transactions: %w", err)
	}
	defer rows.Close()

	type Transaction struct {
		ID           int
		Name         string
		Amount       float64
		Date         string
		Category     *string
		Subcategory  *string
		MerchantName *string
	}

	var transactions []Transaction
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.ID, &t.Name, &t.Amount, &t.Date, &t.Category, &t.Subcategory, &t.MerchantName); err != nil {
			continue
		}
		transactions = append(transactions, t)
	}

	if len(transactions) == 0 {
		return `{"error": "No transactions found for the specified period", "suggestion": "Connect a bank account via Plaid or import transactions via CSV"}`, nil
	}

	// Categorize transactions
	incomeCategories := map[string]bool{
		"INCOME": true, "INCOME_WAGES": true, "INCOME_DIVIDENDS": true,
		"INCOME_INTEREST": true, "TRANSFER_IN": true,
	}

	essentialCategories := map[string]bool{
		"RENT_AND_UTILITIES": true, "FOOD_AND_DRINK": true, "MEDICAL": true,
		"TRANSPORTATION": true, "LOAN_PAYMENTS": true, "INSURANCE": true,
		"HOME_IMPROVEMENT": true, "GENERAL_SERVICES": true,
	}

	// Analyze spending by category
	categoryTotals := make(map[string]float64)
	merchantTotals := make(map[string]float64)
	monthlyData := make(map[string]map[string]float64) // month -> {income, expenses}
	var totalIncome, totalExpenses float64
	var essentialSpending, discretionarySpending float64

	// Track recurring transactions (same merchant, similar amount)
	merchantCounts := make(map[string]int)
	merchantAmounts := make(map[string][]float64)

	for _, t := range transactions {
		cat := "Uncategorized"
		if t.Category != nil {
			cat = *t.Category
		}

		merchant := "Unknown"
		if t.MerchantName != nil && *t.MerchantName != "" {
			merchant = *t.MerchantName
		} else if t.Name != "" {
			merchant = t.Name
		}

		// Get month key
		monthKey := t.Date[:7] // YYYY-MM

		if monthlyData[monthKey] == nil {
			monthlyData[monthKey] = make(map[string]float64)
		}

		isIncomeCategory := incomeCategories[cat]
		if t.Amount < 0 || isIncomeCategory {
			totalIncome += abs(t.Amount)
			monthlyData[monthKey]["income"] += abs(t.Amount)
		} else {
			totalExpenses += t.Amount
			monthlyData[monthKey]["expenses"] += t.Amount
			categoryTotals[cat] += t.Amount
			merchantTotals[merchant] += t.Amount
			merchantCounts[merchant]++
			merchantAmounts[merchant] = append(merchantAmounts[merchant], t.Amount)

			if essentialCategories[cat] {
				essentialSpending += t.Amount
			} else {
				discretionarySpending += t.Amount
			}
		}
	}

	// Find recurring subscriptions (same merchant appearing monthly with similar amounts)
	type Subscription struct {
		Merchant      string  `json:"merchant"`
		AvgAmount     float64 `json:"avg_amount"`
		Frequency     int     `json:"occurrences"`
		AnnualCost    float64 `json:"annual_cost"`
		IsLikelySubscription bool `json:"is_likely_subscription"`
	}

	var subscriptions []Subscription
	for merchant, count := range merchantCounts {
		if count >= 2 && len(merchantAmounts[merchant]) >= 2 {
			amounts := merchantAmounts[merchant]
			sum := 0.0
			for _, a := range amounts {
				sum += a
			}
			avg := sum / float64(len(amounts))

			// Check if amounts are consistent (subscription-like)
			isConsistent := true
			for _, a := range amounts {
				if abs(a-avg)/avg > 0.2 { // More than 20% variance
					isConsistent = false
					break
				}
			}

			if isConsistent && avg > 5 { // At least $5 per occurrence
				annualCost := avg * 12 / float64(months) * float64(count)
				subscriptions = append(subscriptions, Subscription{
					Merchant:             merchant,
					AvgAmount:            math.Round(avg*100) / 100,
					Frequency:            count,
					AnnualCost:           math.Round(annualCost*100) / 100,
					IsLikelySubscription: count >= months/2, // At least every other month
				})
			}
		}
	}

	// Sort subscriptions by annual cost
	for i := 0; i < len(subscriptions); i++ {
		for j := i + 1; j < len(subscriptions); j++ {
			if subscriptions[j].AnnualCost > subscriptions[i].AnnualCost {
				subscriptions[i], subscriptions[j] = subscriptions[j], subscriptions[i]
			}
		}
	}
	if len(subscriptions) > 10 {
		subscriptions = subscriptions[:10]
	}

	// Calculate monthly trends
	type MonthlyTrend struct {
		Month    string  `json:"month"`
		Income   float64 `json:"income"`
		Expenses float64 `json:"expenses"`
		Savings  float64 `json:"savings"`
		Rate     float64 `json:"savings_rate"`
	}

	var monthlyTrends []MonthlyTrend
	var monthKeys []string
	for k := range monthlyData {
		monthKeys = append(monthKeys, k)
	}
	// Sort months
	for i := 0; i < len(monthKeys); i++ {
		for j := i + 1; j < len(monthKeys); j++ {
			if monthKeys[j] < monthKeys[i] {
				monthKeys[i], monthKeys[j] = monthKeys[j], monthKeys[i]
			}
		}
	}

	for _, month := range monthKeys {
		data := monthlyData[month]
		income := data["income"]
		expenses := data["expenses"]
		savings := income - expenses
		rate := 0.0
		if income > 0 {
			rate = savings / income * 100
		}
		monthlyTrends = append(monthlyTrends, MonthlyTrend{
			Month:    month,
			Income:   math.Round(income*100) / 100,
			Expenses: math.Round(expenses*100) / 100,
			Savings:  math.Round(savings*100) / 100,
			Rate:     math.Round(rate*100) / 100,
		})
	}

	// Top spending categories
	type CategorySpend struct {
		Category   string  `json:"category"`
		Amount     float64 `json:"amount"`
		Percentage float64 `json:"percentage"`
	}

	var topCategories []CategorySpend
	for cat, amount := range categoryTotals {
		pct := 0.0
		if totalExpenses > 0 {
			pct = amount / totalExpenses * 100
		}
		topCategories = append(topCategories, CategorySpend{
			Category:   cat,
			Amount:     math.Round(amount*100) / 100,
			Percentage: math.Round(pct*100) / 100,
		})
	}
	// Sort by amount descending
	for i := 0; i < len(topCategories); i++ {
		for j := i + 1; j < len(topCategories); j++ {
			if topCategories[j].Amount > topCategories[i].Amount {
				topCategories[i], topCategories[j] = topCategories[j], topCategories[i]
			}
		}
	}
	if len(topCategories) > 10 {
		topCategories = topCategories[:10]
	}

	// Generate insights
	var insights []string

	// Savings rate insight
	avgSavingsRate := 0.0
	if totalIncome > 0 {
		avgSavingsRate = (totalIncome - totalExpenses) / totalIncome * 100
	}
	if avgSavingsRate >= 20 {
		insights = append(insights, fmt.Sprintf("Strong savings rate of %.1f%% - you're saving more than the recommended 20%%", avgSavingsRate))
	} else if avgSavingsRate >= 10 {
		insights = append(insights, fmt.Sprintf("Moderate savings rate of %.1f%% - aim for 20%% if possible", avgSavingsRate))
	} else if avgSavingsRate >= 0 {
		insights = append(insights, fmt.Sprintf("Low savings rate of %.1f%% - consider reducing discretionary spending", avgSavingsRate))
	} else {
		insights = append(insights, fmt.Sprintf("Negative savings rate of %.1f%% - expenses exceed income", avgSavingsRate))
	}

	// Essential vs discretionary
	if totalExpenses > 0 {
		essentialPct := essentialSpending / totalExpenses * 100
		if essentialPct > 80 {
			insights = append(insights, fmt.Sprintf("%.1f%% of spending is on essentials - limited room for cuts", essentialPct))
		} else if essentialPct < 50 {
			insights = append(insights, fmt.Sprintf("%.1f%% of spending is discretionary - opportunity to increase savings", 100-essentialPct))
		}
	}

	// Subscription insight
	totalSubCost := 0.0
	for _, sub := range subscriptions {
		if sub.IsLikelySubscription {
			totalSubCost += sub.AnnualCost
		}
	}
	if totalSubCost > 500 {
		insights = append(insights, fmt.Sprintf("Identified ~$%.0f/year in recurring subscriptions - review for unused services", totalSubCost))
	}

	// Trend insight (compare first half to second half of period)
	if len(monthlyTrends) >= 4 {
		halfPoint := len(monthlyTrends) / 2
		firstHalfExpenses := 0.0
		secondHalfExpenses := 0.0
		for i, mt := range monthlyTrends {
			if i < halfPoint {
				firstHalfExpenses += mt.Expenses
			} else {
				secondHalfExpenses += mt.Expenses
			}
		}
		firstHalfAvg := firstHalfExpenses / float64(halfPoint)
		secondHalfAvg := secondHalfExpenses / float64(len(monthlyTrends)-halfPoint)
		changePercent := (secondHalfAvg - firstHalfAvg) / firstHalfAvg * 100

		if changePercent > 10 {
			insights = append(insights, fmt.Sprintf("Spending increased %.1f%% in recent months - possible lifestyle inflation", changePercent))
		} else if changePercent < -10 {
			insights = append(insights, fmt.Sprintf("Spending decreased %.1f%% in recent months - good progress", -changePercent))
		}
	}

	result := map[string]interface{}{
		"period": map[string]interface{}{
			"months":     months,
			"start_date": startDate,
			"end_date":   endDate,
		},
		"summary": map[string]interface{}{
			"total_income":           math.Round(totalIncome*100) / 100,
			"total_expenses":         math.Round(totalExpenses*100) / 100,
			"net_savings":            math.Round((totalIncome-totalExpenses)*100) / 100,
			"avg_monthly_income":     math.Round(totalIncome/float64(months)*100) / 100,
			"avg_monthly_expenses":   math.Round(totalExpenses/float64(months)*100) / 100,
			"savings_rate_percent":   math.Round(avgSavingsRate*100) / 100,
			"essential_spending":     math.Round(essentialSpending*100) / 100,
			"discretionary_spending": math.Round(discretionarySpending*100) / 100,
		},
		"top_categories":     topCategories,
		"monthly_trends":     monthlyTrends,
		"recurring_expenses": subscriptions,
		"insights":           insights,
		"transaction_count":  len(transactions),
	}

	// Compare to prior period if requested
	if compareToPrior {
		priorEndDate := startDate
		priorStartDate := now.AddDate(0, -months*2, 0).Format("2006-01-02")

		priorRows, err := db.DB.Query(`
			SELECT SUM(CASE WHEN amount < 0 OR category IN ('INCOME', 'INCOME_WAGES', 'INCOME_DIVIDENDS', 'INCOME_INTEREST', 'TRANSFER_IN') THEN ABS(amount) ELSE 0 END) as income,
			       SUM(CASE WHEN amount > 0 AND category NOT IN ('INCOME', 'INCOME_WAGES', 'INCOME_DIVIDENDS', 'INCOME_INTEREST', 'TRANSFER_IN') THEN amount ELSE 0 END) as expenses
			FROM transactions
			WHERE user_id = ? AND date >= ? AND date < ?
		`, userID, priorStartDate, priorEndDate)
		if err == nil {
			defer priorRows.Close()
			if priorRows.Next() {
				var priorIncome, priorExpenses float64
				priorRows.Scan(&priorIncome, &priorExpenses)

				incomeChange := 0.0
				expenseChange := 0.0
				if priorIncome > 0 {
					incomeChange = (totalIncome - priorIncome) / priorIncome * 100
				}
				if priorExpenses > 0 {
					expenseChange = (totalExpenses - priorExpenses) / priorExpenses * 100
				}

				result["prior_period_comparison"] = map[string]interface{}{
					"prior_period":      fmt.Sprintf("%s to %s", priorStartDate, priorEndDate),
					"prior_income":      math.Round(priorIncome*100) / 100,
					"prior_expenses":    math.Round(priorExpenses*100) / 100,
					"income_change_pct": math.Round(incomeChange*100) / 100,
					"expense_change_pct": math.Round(expenseChange*100) / 100,
				}
			}
		}
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonBytes), nil
}

// getClientSummary gets a comprehensive summary for a client
func (e *ToolExecutor) getClientSummary(input map[string]interface{}) (string, error) {
	if !e.IsAdvisor {
		return "", fmt.Errorf("this tool is only available to advisors")
	}

	clientID, ok := input["client_id"].(float64)
	if !ok {
		return "", fmt.Errorf("client_id is required")
	}

	id := int(clientID)

	// Verify advisor has access
	var accessLevel string
	err := db.DB.QueryRow(`
		SELECT access_level FROM advisor_clients
		WHERE advisor_id = ? AND client_id = ? AND status = 'active'
	`, e.UserID, id).Scan(&accessLevel)

	if err != nil {
		return "", fmt.Errorf("client not found or access denied")
	}

	// Get client info
	var clientName, clientEmail string
	var createdAt time.Time
	db.DB.QueryRow(`SELECT name, email, created_at FROM users WHERE id = ?`, id).Scan(&clientName, &clientEmail, &createdAt)

	// Get asset summary
	var totalAssets float64
	var assetCount int
	db.DB.QueryRow(`SELECT COALESCE(SUM(current_value), 0), COUNT(*) FROM assets WHERE user_id = ?`, id).Scan(&totalAssets, &assetCount)

	// Get debt summary
	var totalDebts float64
	var debtCount int
	db.DB.QueryRow(`SELECT COALESCE(SUM(current_balance), 0), COUNT(*) FROM debts WHERE user_id = ?`, id).Scan(&totalDebts, &debtCount)

	// Get recent simulations
	simRows, _ := db.DB.Query(`
		SELECT id, name, success_rate, final_p50, created_at
		FROM simulation_history
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT 5
	`, id)
	defer simRows.Close()

	type SimBrief struct {
		ID          int     `json:"id"`
		Name        *string `json:"name,omitempty"`
		SuccessRate float64 `json:"success_rate"`
		FinalP50    float64 `json:"final_p50"`
		CreatedAt   string  `json:"created_at"`
	}

	var recentSims []SimBrief
	for simRows.Next() {
		var s SimBrief
		var createdAt time.Time
		if simRows.Scan(&s.ID, &s.Name, &s.SuccessRate, &s.FinalP50, &createdAt) == nil {
			s.CreatedAt = createdAt.Format(time.RFC3339)
			recentSims = append(recentSims, s)
		}
	}

	result := map[string]interface{}{
		"client": map[string]interface{}{
			"id":         id,
			"name":       clientName,
			"email":      clientEmail,
			"created_at": createdAt.Format(time.RFC3339),
		},
		"access_level": accessLevel,
		"net_worth": map[string]interface{}{
			"total_assets": totalAssets,
			"total_debts":  totalDebts,
			"net_worth":    totalAssets - totalDebts,
			"asset_count":  assetCount,
			"debt_count":   debtCount,
		},
		"recent_simulations":   recentSims,
		"simulation_count":     len(recentSims),
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonBytes), nil
}

// analyzeTaxDocument analyzes an uploaded tax document from the document vault
func (e *ToolExecutor) analyzeTaxDocument(input map[string]interface{}) (string, error) {
	userID := e.GetEffectiveUserID()

	// Get document ID from input
	docID, ok := input["document_id"].(float64)
	if !ok {
		return "", fmt.Errorf("document_id is required")
	}

	// Query document and verify access
	var doc struct {
		ID          int
		UserID      int
		UploadedBy  int
		StoragePath string
		Encrypted   bool
		MimeType    string
		Category    string
		Name        string
	}

	err := db.DB.QueryRow(`
		SELECT id, user_id, uploaded_by, storage_path, encrypted, mime_type, category, name
		FROM documents
		WHERE id = ? AND deleted_at IS NULL
	`, int(docID)).Scan(&doc.ID, &doc.UserID, &doc.UploadedBy,
		&doc.StoragePath, &doc.Encrypted, &doc.MimeType, &doc.Category, &doc.Name)

	if err != nil {
		return "", fmt.Errorf("document not found or access denied")
	}

	// Check access: user owns document, uploaded it, or has advisor access
	hasAccess := doc.UserID == userID || doc.UploadedBy == userID

	if !hasAccess {
		// Check if shared with user
		var shareCount int
		db.DB.QueryRow(`
			SELECT COUNT(*) FROM document_shares
			WHERE document_id = ? AND shared_with_id = ?
			AND (expires_at IS NULL OR expires_at > NOW())
		`, doc.ID, userID).Scan(&shareCount)
		hasAccess = shareCount > 0
	}

	if !hasAccess && e.IsAdvisor {
		// Check advisor-client relationship
		var accessLevel string
		db.DB.QueryRow(`
			SELECT access_level FROM advisor_clients
			WHERE advisor_id = ? AND client_id = ? AND status = 'active'
		`, e.UserID, doc.UserID).Scan(&accessLevel)
		hasAccess = accessLevel != ""
	}

	if !hasAccess {
		return "", fmt.Errorf("access denied to this document")
	}

	// Verify it's a PDF
	if doc.MimeType != "application/pdf" {
		return "", fmt.Errorf("only PDF documents are supported for tax analysis (got %s)", doc.MimeType)
	}

	// Load document from storage
	pdfBytes, err := storage.DefaultStorage.Load(doc.StoragePath, doc.Encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to load document: %w", err)
	}

	// Parse the tax document
	taxData, err := taxparser.ParsePDFContent(pdfBytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse document: %w", err)
	}

	// Generate comprehensive analysis
	analysis := e.generateTaxAnalysis(taxData, doc.Name)

	jsonBytes, _ := json.MarshalIndent(analysis, "", "  ")
	return string(jsonBytes), nil
}

// generateTaxAnalysis creates comprehensive analysis from extracted tax data
func (e *ToolExecutor) generateTaxAnalysis(data *taxparser.ExtractedTaxData, docName string) map[string]interface{} {
	userID := e.GetEffectiveUserID()

	result := map[string]interface{}{
		"document_name": docName,
		"document_type": data.DocumentType,
		"tax_year":      data.TaxYear,
		"confidence":    math.Round(data.Confidence*100) / 100,
	}

	if len(data.ParseErrors) > 0 {
		result["parse_warnings"] = data.ParseErrors
	}

	// Add extracted data based on document type
	switch data.DocumentType {
	case taxparser.DocType1040:
		result["filing_status"] = data.FilingStatus

		incomeSummary := map[string]interface{}{}
		if data.TotalIncome != nil {
			incomeSummary["total_income"] = *data.TotalIncome
		}
		if data.AGI != nil {
			incomeSummary["agi"] = *data.AGI
		}
		if data.TaxableIncome != nil {
			incomeSummary["taxable_income"] = *data.TaxableIncome
		}
		if len(incomeSummary) > 0 {
			result["income_summary"] = incomeSummary
		}

		deductions := map[string]interface{}{}
		if data.StandardDeduction != nil {
			deductions["type"] = "standard"
			deductions["amount"] = *data.StandardDeduction
		} else if data.ItemizedDeductions != nil {
			deductions["type"] = "itemized"
			deductions["amount"] = *data.ItemizedDeductions
		}
		if len(deductions) > 0 {
			result["deductions"] = deductions
		}

		taxSummary := map[string]interface{}{}
		if data.TotalTax != nil {
			taxSummary["total_tax"] = *data.TotalTax
		}
		if data.TotalPayments != nil {
			taxSummary["total_payments"] = *data.TotalPayments
		}
		if data.RefundAmount != nil {
			taxSummary["refund"] = *data.RefundAmount
		}
		if data.AmountOwed != nil {
			taxSummary["amount_owed"] = *data.AmountOwed
		}

		// Calculate effective tax rate
		if data.AGI != nil && *data.AGI > 0 && data.TotalTax != nil {
			effectiveRate := (*data.TotalTax / *data.AGI) * 100
			taxSummary["effective_rate"] = math.Round(effectiveRate*100) / 100
		}
		if len(taxSummary) > 0 {
			result["tax_summary"] = taxSummary
		}

	case taxparser.DocTypeW2:
		result["employer"] = data.Employer

		wagesAndWithholding := map[string]interface{}{}
		if data.WagesTips != nil {
			wagesAndWithholding["wages_tips"] = *data.WagesTips
		}
		if data.FederalWithheld != nil {
			wagesAndWithholding["federal_withheld"] = *data.FederalWithheld
		}
		if data.SocialSecurityWages != nil {
			wagesAndWithholding["ss_wages"] = *data.SocialSecurityWages
		}
		if data.SocialSecurityTax != nil {
			wagesAndWithholding["ss_tax"] = *data.SocialSecurityTax
		}
		if data.MedicareWages != nil {
			wagesAndWithholding["medicare_wages"] = *data.MedicareWages
		}
		if data.MedicareTax != nil {
			wagesAndWithholding["medicare_tax"] = *data.MedicareTax
		}
		if len(wagesAndWithholding) > 0 {
			result["wages_and_withholding"] = wagesAndWithholding
		}

	case taxparser.DocType1099:
		result["payer"] = data.Payer
		result["income_type"] = data.IncomeType
		if data.GrossIncome != nil {
			result["gross_income"] = *data.GrossIncome
		}
	}

	// Generate optimization suggestions based on extracted data
	result["optimization_opportunities"] = e.generateTaxOptimizations(data, userID)

	return result
}

// generateTaxOptimizations creates actionable tax optimization suggestions
func (e *ToolExecutor) generateTaxOptimizations(data *taxparser.ExtractedTaxData, userID int) []map[string]interface{} {
	var suggestions []map[string]interface{}

	if data.DocumentType == taxparser.DocType1040 {
		// Get user's current retirement account balances
		var retirement401k, rothIRA, traditionalIRA float64
		db.DB.QueryRow(`
			SELECT
				COALESCE(SUM(CASE WHEN at.name LIKE '%401%' THEN a.current_value ELSE 0 END), 0),
				COALESCE(SUM(CASE WHEN at.name LIKE '%Roth%' THEN a.current_value ELSE 0 END), 0),
				COALESCE(SUM(CASE WHEN at.name LIKE '%Traditional IRA%' OR at.name LIKE '%IRA%' THEN a.current_value ELSE 0 END), 0)
			FROM assets a
			LEFT JOIN asset_types at ON a.type_id = at.id
			WHERE a.user_id = ?
		`, userID).Scan(&retirement401k, &rothIRA, &traditionalIRA)

		// 401(k) contribution space
		if data.AGI != nil && *data.AGI > 50000 {
			limit401k := 23000.0 // 2024 limit
			taxSavings := 0.0
			if data.TaxableIncome != nil && data.TotalTax != nil && *data.TaxableIncome > 0 {
				effectiveRate := *data.TotalTax / *data.TaxableIncome
				taxSavings = limit401k * effectiveRate
			}

			desc := fmt.Sprintf("Contributing the full $%.0f to a 401(k) could reduce taxable income.", limit401k)
			if retirement401k > 0 {
				desc += fmt.Sprintf(" Current 401(k) balance: $%.0f.", retirement401k)
			}

			suggestion := map[string]interface{}{
				"category":    "retirement",
				"title":       "401(k) Contribution Space",
				"description": desc,
			}
			if taxSavings > 0 {
				suggestion["potential_savings"] = math.Round(taxSavings*100) / 100
			}
			suggestions = append(suggestions, suggestion)
		}

		// Roth conversion opportunity
		if data.TaxableIncome != nil {
			// Check if in 22% bracket or lower (good for Roth)
			bracket22Limit := 100525.0 // Single 2024
			if data.FilingStatus == "married_filing_jointly" {
				bracket22Limit = 201050.0
			}

			if *data.TaxableIncome < bracket22Limit {
				rothSpace := bracket22Limit - *data.TaxableIncome
				suggestions = append(suggestions, map[string]interface{}{
					"category":    "roth_conversion",
					"title":       "Roth Conversion Opportunity",
					"description": "You're in a lower tax bracket. Consider converting traditional IRA/401(k) to Roth to lock in current tax rates.",
					"roth_space":  math.Round(rothSpace*100) / 100,
				})
			}
		}

		// HSA suggestion
		if data.StandardDeduction != nil && *data.StandardDeduction > 0 {
			hsaLimit := 4150.0 // 2024 single
			if data.FilingStatus == "married_filing_jointly" {
				hsaLimit = 8300.0
			}
			suggestions = append(suggestions, map[string]interface{}{
				"category":        "hsa",
				"title":           "HSA Contribution",
				"description":     "If you have a high-deductible health plan, HSA contributions provide triple tax advantage: deduction now, tax-free growth, and tax-free withdrawals for medical expenses.",
				"max_contribution": hsaLimit,
			})
		}

		// Charitable giving / itemization opportunity
		if data.StandardDeduction != nil && data.ItemizedDeductions == nil {
			suggestions = append(suggestions, map[string]interface{}{
				"category":    "deductions",
				"title":       "Bunching Strategy",
				"description": "Consider bunching charitable donations or medical expenses into a single year to exceed the standard deduction and itemize.",
			})
		}
	}

	// W-2 specific suggestions
	if data.DocumentType == taxparser.DocTypeW2 {
		if data.WagesTips != nil && data.FederalWithheld != nil {
			// Check if withholding seems appropriate
			estimatedTax := *data.WagesTips * 0.22 // Rough estimate at 22%
			withholdingRatio := *data.FederalWithheld / estimatedTax

			if withholdingRatio < 0.8 {
				suggestions = append(suggestions, map[string]interface{}{
					"category":    "withholding",
					"title":       "Review Withholding",
					"description": "Your federal withholding may be lower than expected. Consider reviewing your W-4 to avoid a large tax bill.",
				})
			} else if withholdingRatio > 1.3 {
				suggestions = append(suggestions, map[string]interface{}{
					"category":    "withholding",
					"title":       "Adjust Withholding",
					"description": "You may be over-withholding. Consider adjusting your W-4 to increase take-home pay.",
				})
			}
		}
	}

	// 1099 specific suggestions
	if data.DocumentType == taxparser.DocType1099 {
		if data.GrossIncome != nil && *data.GrossIncome > 1000 {
			suggestions = append(suggestions, map[string]interface{}{
				"category":    "estimated_taxes",
				"title":       "Quarterly Estimated Taxes",
				"description": fmt.Sprintf("With $%.0f in 1099 income, you may need to make quarterly estimated tax payments to avoid penalties.", *data.GrossIncome),
			})

			if data.IncomeType == "NEC" || data.IncomeType == "MISC" {
				suggestions = append(suggestions, map[string]interface{}{
					"category":    "self_employment",
					"title":       "Self-Employment Tax",
					"description": "Self-employment income is subject to an additional 15.3% SE tax. Consider setting aside extra for taxes.",
				})
			}
		}
	}

	return suggestions
}

// generateMeetingPrep creates a comprehensive client briefing document for advisors
func (e *ToolExecutor) generateMeetingPrep(input map[string]interface{}) (string, error) {
	// This tool is advisor-only
	if !e.IsAdvisor {
		return "", fmt.Errorf("this tool is only available for advisors")
	}

	// Get client ID from input or use current context
	var clientID int
	if cid, ok := input["client_id"].(float64); ok {
		clientID = int(cid)
	} else if e.ClientContext != 0 {
		clientID = e.ClientContext
	} else {
		return "", fmt.Errorf("client_id is required or switch to a client context first")
	}

	// Verify advisor has access to this client
	var accessLevel string
	err := db.DB.QueryRow(`
		SELECT access_level FROM advisor_clients
		WHERE advisor_id = ? AND client_id = ? AND status = 'active'
	`, e.UserID, clientID).Scan(&accessLevel)
	if err != nil {
		return "", fmt.Errorf("you don't have access to this client")
	}

	// Get focus areas (optional)
	var focusAreas []string
	if fa, ok := input["focus_areas"].([]interface{}); ok {
		for _, area := range fa {
			if s, ok := area.(string); ok {
				focusAreas = append(focusAreas, s)
			}
		}
	}

	// Gather client information
	var clientName, clientEmail string
	var clientCreatedAt time.Time
	err = db.DB.QueryRow(`
		SELECT name, email, created_at FROM users WHERE id = ?
	`, clientID).Scan(&clientName, &clientEmail, &clientCreatedAt)
	if err != nil {
		return "", fmt.Errorf("client not found")
	}

	// === NET WORTH ANALYSIS ===
	// Current totals
	var totalAssets, totalDebts float64
	db.DB.QueryRow(`SELECT COALESCE(SUM(current_value), 0) FROM assets WHERE user_id = ?`, clientID).Scan(&totalAssets)
	db.DB.QueryRow(`SELECT COALESCE(SUM(current_balance), 0) FROM debts WHERE user_id = ?`, clientID).Scan(&totalDebts)
	netWorth := totalAssets - totalDebts

	// Get asset breakdown by type
	assetRows, _ := db.DB.Query(`
		SELECT at.name, SUM(a.current_value) as total
		FROM assets a
		LEFT JOIN asset_types at ON a.type_id = at.id
		WHERE a.user_id = ?
		GROUP BY at.name
		ORDER BY total DESC
	`, clientID)
	defer assetRows.Close()

	assetAllocation := make(map[string]float64)
	for assetRows.Next() {
		var typeName string
		var total float64
		if assetRows.Scan(&typeName, &total) == nil {
			assetAllocation[typeName] = total
		}
	}

	// === RECENT TRANSACTIONS ANALYSIS ===
	// Get last 3 months of spending by category
	threeMonthsAgo := time.Now().AddDate(0, -3, 0).Format("2006-01-02")
	oneMonthAgo := time.Now().AddDate(0, -1, 0).Format("2006-01-02")

	// Recent month spending
	var recentMonthSpending float64
	db.DB.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = ? AND amount > 0 AND date >= ?
	`, clientID, oneMonthAgo).Scan(&recentMonthSpending)

	// 3-month average
	var threeMonthSpending float64
	db.DB.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = ? AND amount > 0 AND date >= ?
	`, clientID, threeMonthsAgo).Scan(&threeMonthSpending)
	avgMonthlySpending := threeMonthSpending / 3

	// Top spending categories (last 3 months)
	catRows, _ := db.DB.Query(`
		SELECT COALESCE(category, 'Uncategorized'), SUM(amount) as total
		FROM transactions
		WHERE user_id = ? AND amount > 0 AND date >= ?
		GROUP BY category
		ORDER BY total DESC
		LIMIT 5
	`, clientID, threeMonthsAgo)
	defer catRows.Close()

	type CategorySpend struct {
		Category string  `json:"category"`
		Amount   float64 `json:"amount"`
	}
	var topCategories []CategorySpend
	for catRows.Next() {
		var cs CategorySpend
		if catRows.Scan(&cs.Category, &cs.Amount) == nil {
			topCategories = append(topCategories, cs)
		}
	}

	// === SIMULATION HISTORY ===
	simRows, _ := db.DB.Query(`
		SELECT id, name, success_rate, final_p50, starting_net_worth, time_horizon_years, created_at
		FROM simulation_history
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT 5
	`, clientID)
	defer simRows.Close()

	var recentSims []MeetingPrepSimSummary
	for simRows.Next() {
		var sim MeetingPrepSimSummary
		if simRows.Scan(&sim.ID, &sim.Name, &sim.SuccessRate, &sim.FinalP50, &sim.StartingNW, &sim.TimeHorizon, &sim.CreatedAt) == nil {
			recentSims = append(recentSims, sim)
		}
	}

	// === GENERATE AGENDA ITEMS ===
	agendaItems := []map[string]interface{}{}

	// Always include net worth review
	agendaItems = append(agendaItems, map[string]interface{}{
		"topic":       "Net Worth Review",
		"description": fmt.Sprintf("Current net worth: $%.0f (Assets: $%.0f, Debts: $%.0f)", netWorth, totalAssets, totalDebts),
		"priority":    "high",
	})

	// Spending analysis if transactions exist
	if recentMonthSpending > 0 {
		spendingChange := ((recentMonthSpending - avgMonthlySpending) / avgMonthlySpending) * 100
		spendingDesc := fmt.Sprintf("Last month: $%.0f (3-month avg: $%.0f)", recentMonthSpending, avgMonthlySpending)
		if abs(spendingChange) > 10 {
			if spendingChange > 0 {
				spendingDesc += fmt.Sprintf(". Spending UP %.0f%% vs average.", spendingChange)
			} else {
				spendingDesc += fmt.Sprintf(". Spending DOWN %.0f%% vs average.", -spendingChange)
			}
		}
		agendaItems = append(agendaItems, map[string]interface{}{
			"topic":       "Spending Review",
			"description": spendingDesc,
			"priority":    "medium",
		})
	}

	// Retirement/simulation review if exists
	if len(recentSims) > 0 {
		latestSim := recentSims[0]
		simDesc := fmt.Sprintf("Latest simulation: %.0f%% success rate, projected $%.0f at end of %d years",
			latestSim.SuccessRate, latestSim.FinalP50, latestSim.TimeHorizon)
		priority := "medium"
		if latestSim.SuccessRate < 70 {
			priority = "high"
			simDesc += ". SUCCESS RATE BELOW 70% - discuss strategies to improve."
		}
		agendaItems = append(agendaItems, map[string]interface{}{
			"topic":       "Retirement Projection Review",
			"description": simDesc,
			"priority":    priority,
		})
	}

	// Asset allocation check
	if len(assetAllocation) > 0 {
		agendaItems = append(agendaItems, map[string]interface{}{
			"topic":       "Portfolio Allocation Review",
			"description": "Review current allocation and consider rebalancing if needed.",
			"priority":    "medium",
		})
	}

	// Add focus area items
	for _, area := range focusAreas {
		switch area {
		case "tax":
			agendaItems = append(agendaItems, map[string]interface{}{
				"topic":       "Tax Planning",
				"description": "Review tax optimization opportunities: Roth conversions, HSA, 401(k) contributions, tax-loss harvesting.",
				"priority":    "medium",
			})
		case "estate":
			agendaItems = append(agendaItems, map[string]interface{}{
				"topic":       "Estate Planning",
				"description": "Review beneficiaries, wills, trusts, and power of attorney documents.",
				"priority":    "medium",
			})
		case "insurance":
			agendaItems = append(agendaItems, map[string]interface{}{
				"topic":       "Insurance Review",
				"description": "Evaluate life, disability, long-term care, and umbrella coverage.",
				"priority":    "medium",
			})
		case "goals":
			agendaItems = append(agendaItems, map[string]interface{}{
				"topic":       "Goals Review",
				"description": "Discuss progress toward financial goals and any changes in priorities.",
				"priority":    "high",
			})
		}
	}

	// Always add action items topic
	agendaItems = append(agendaItems, map[string]interface{}{
		"topic":       "Action Items & Next Steps",
		"description": "Review outstanding tasks and set new action items.",
		"priority":    "high",
	})

	// === CLIENT NOTES ===
	noteRows, _ := db.DB.Query(`
		SELECT id, note, category, is_pinned, created_at
		FROM client_notes
		WHERE advisor_id = ? AND client_id = ?
		ORDER BY is_pinned DESC, created_at DESC
		LIMIT 10
	`, e.UserID, clientID)
	defer noteRows.Close()

	type ClientNote struct {
		ID        int       `json:"id"`
		Note      string    `json:"note"`
		Category  string    `json:"category"`
		IsPinned  bool      `json:"is_pinned"`
		CreatedAt time.Time `json:"created_at"`
	}
	var clientNotes []ClientNote
	var pinnedNotes []ClientNote
	var actionItemNotes []ClientNote

	for noteRows.Next() {
		var n ClientNote
		if noteRows.Scan(&n.ID, &n.Note, &n.Category, &n.IsPinned, &n.CreatedAt) == nil {
			clientNotes = append(clientNotes, n)
			if n.IsPinned {
				pinnedNotes = append(pinnedNotes, n)
			}
			if n.Category == "action_item" {
				actionItemNotes = append(actionItemNotes, n)
			}
		}
	}

	// Add pinned notes to talking points
	for _, note := range pinnedNotes {
		if len(note.Note) > 100 {
			// Truncate long notes in talking points
			agendaItems = append(agendaItems, map[string]interface{}{
				"topic":       fmt.Sprintf("[%s] Pinned Note", note.Category),
				"description": note.Note[:100] + "...",
				"priority":    "high",
			})
		}
	}

	// Add action items to agenda if they exist
	if len(actionItemNotes) > 0 {
		actionDesc := fmt.Sprintf("%d outstanding action items to review:", len(actionItemNotes))
		for i, item := range actionItemNotes {
			if i < 3 { // Show first 3
				actionDesc += fmt.Sprintf("\n  - %s", item.Note)
				if len(item.Note) > 50 {
					actionDesc = actionDesc[:len(actionDesc)-len(item.Note)+47] + "..."
				}
			}
		}
		agendaItems = append(agendaItems, map[string]interface{}{
			"topic":       "Outstanding Action Items",
			"description": actionDesc,
			"priority":    "high",
		})
	}

	// === BUILD BRIEFING DOCUMENT ===
	briefing := map[string]interface{}{
		"generated_at": time.Now().Format(time.RFC3339),
		"client": map[string]interface{}{
			"id":            clientID,
			"name":          clientName,
			"email":         clientEmail,
			"client_since":  clientCreatedAt.Format("January 2006"),
		},
		"net_worth_summary": map[string]interface{}{
			"total_assets":     totalAssets,
			"total_debts":      totalDebts,
			"net_worth":        netWorth,
			"asset_allocation": assetAllocation,
		},
		"spending_summary": map[string]interface{}{
			"last_month_total":   recentMonthSpending,
			"three_month_avg":    math.Round(avgMonthlySpending*100) / 100,
			"top_categories":     topCategories,
		},
		"simulation_summary": map[string]interface{}{
			"recent_simulations": recentSims,
			"count":              len(recentSims),
		},
		"meeting_agenda": agendaItems,
		"talking_points": generateTalkingPoints(netWorth, recentMonthSpending, avgMonthlySpending, convertToSimBriefs(recentSims)),
		"advisor_notes": map[string]interface{}{
			"notes": clientNotes,
			"count": len(clientNotes),
			"pinned_count": len(pinnedNotes),
			"action_items_count": len(actionItemNotes),
		},
	}

	jsonBytes, _ := json.MarshalIndent(briefing, "", "  ")
	return string(jsonBytes), nil
}

// simulationBrief is a simplified simulation data for talking points
type simulationBrief struct {
	SuccessRate float64
}

// MeetingPrepSimSummary is used internally in generateMeetingPrep
type MeetingPrepSimSummary struct {
	ID          int       `json:"id"`
	Name        *string   `json:"name,omitempty"`
	SuccessRate float64   `json:"success_rate"`
	FinalP50    float64   `json:"final_p50"`
	StartingNW  float64   `json:"starting_net_worth"`
	TimeHorizon int       `json:"time_horizon_years"`
	CreatedAt   time.Time `json:"created_at"`
}

// convertToSimBriefs converts MeetingPrepSimSummary slice to simulationBrief slice
func convertToSimBriefs(sims []MeetingPrepSimSummary) []simulationBrief {
	briefs := make([]simulationBrief, len(sims))
	for i, s := range sims {
		briefs[i] = simulationBrief{SuccessRate: s.SuccessRate}
	}
	return briefs
}

// generateTalkingPoints creates key discussion points based on client data
func generateTalkingPoints(netWorth, recentSpending, avgSpending float64, sims []simulationBrief) []string {
	points := []string{}

	// Net worth milestone check
	milestones := []float64{100000, 250000, 500000, 750000, 1000000, 2000000, 5000000}
	for i := len(milestones) - 1; i >= 0; i-- {
		if netWorth >= milestones[i] {
			points = append(points, fmt.Sprintf("Congratulations on reaching the $%.0fK net worth milestone!", milestones[i]/1000))
			break
		}
	}

	// Spending trend
	if avgSpending > 0 {
		change := ((recentSpending - avgSpending) / avgSpending) * 100
		if change > 20 {
			points = append(points, fmt.Sprintf("Spending increased %.0f%% last month vs 3-month average. Any unusual expenses?", change))
		} else if change < -20 {
			points = append(points, fmt.Sprintf("Great job reducing spending by %.0f%% last month!", -change))
		}
	}

	// Simulation insights
	if len(sims) > 0 {
		latest := sims[0]
		if latest.SuccessRate >= 90 {
			points = append(points, "Excellent retirement outlook - consider whether you can increase lifestyle spending or retire earlier.")
		} else if latest.SuccessRate >= 70 {
			points = append(points, "Solid retirement trajectory. Small improvements in savings could boost success rate further.")
		} else if latest.SuccessRate >= 50 {
			points = append(points, "Retirement plan needs attention. Discuss increasing contributions or adjusting timeline.")
		} else {
			points = append(points, "URGENT: Retirement success rate below 50%. Need to discuss significant changes to the plan.")
		}
	}

	// Add general topics if no specific points
	if len(points) == 0 {
		points = append(points, "Review any changes in income, expenses, or life circumstances since last meeting.")
	}

	return points
}

// =====================
// CLIENT NOTES TOOLS
// =====================

// getClientNotes retrieves notes for a client (advisor only)
func (e *ToolExecutor) getClientNotes(input map[string]interface{}) (string, error) {
	if !e.IsAdvisor {
		return "", fmt.Errorf("this tool is only available to advisors")
	}

	// Get client ID from input or use current context
	var clientID int
	if cid, ok := input["client_id"].(float64); ok {
		clientID = int(cid)
	} else if e.ClientContext != 0 {
		clientID = e.ClientContext
	} else {
		return "", fmt.Errorf("client_id is required or switch to a client context first")
	}

	// Verify advisor has access to this client
	var accessLevel string
	err := db.DB.QueryRow(`
		SELECT access_level FROM advisor_clients
		WHERE advisor_id = ? AND client_id = ? AND status = 'active'
	`, e.UserID, clientID).Scan(&accessLevel)
	if err != nil {
		return "", fmt.Errorf("you don't have access to this client")
	}

	// Optional category filter
	category := ""
	if cat, ok := input["category"].(string); ok {
		category = cat
	}

	// Build query
	var query string
	var args []interface{}
	if category != "" {
		query = `SELECT id, advisor_id, client_id, note, category, is_pinned, created_at, updated_at
			FROM client_notes
			WHERE advisor_id = ? AND client_id = ? AND category = ?
			ORDER BY is_pinned DESC, created_at DESC`
		args = []interface{}{e.UserID, clientID, category}
	} else {
		query = `SELECT id, advisor_id, client_id, note, category, is_pinned, created_at, updated_at
			FROM client_notes
			WHERE advisor_id = ? AND client_id = ?
			ORDER BY is_pinned DESC, created_at DESC`
		args = []interface{}{e.UserID, clientID}
	}

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return "", fmt.Errorf("failed to fetch notes: %w", err)
	}
	defer rows.Close()

	type Note struct {
		ID        int       `json:"id"`
		Note      string    `json:"note"`
		Category  string    `json:"category"`
		IsPinned  bool      `json:"is_pinned"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	var notes []Note
	for rows.Next() {
		var n Note
		var advisorID, cID int
		if err := rows.Scan(&n.ID, &advisorID, &cID, &n.Note, &n.Category, &n.IsPinned, &n.CreatedAt, &n.UpdatedAt); err != nil {
			continue
		}
		notes = append(notes, n)
	}

	// Get client name for context
	var clientName string
	db.DB.QueryRow(`SELECT name FROM users WHERE id = ?`, clientID).Scan(&clientName)

	result := map[string]interface{}{
		"client_id":   clientID,
		"client_name": clientName,
		"notes":       notes,
		"count":       len(notes),
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonBytes), nil
}

// addClientNote creates a new note for a client (advisor only)
func (e *ToolExecutor) addClientNote(input map[string]interface{}) (string, error) {
	if !e.IsAdvisor {
		return "", fmt.Errorf("this tool is only available to advisors")
	}

	// Get client ID from input or use current context
	var clientID int
	if cid, ok := input["client_id"].(float64); ok {
		clientID = int(cid)
	} else if e.ClientContext != 0 {
		clientID = e.ClientContext
	} else {
		return "", fmt.Errorf("client_id is required or switch to a client context first")
	}

	// Get note content (required)
	noteContent, ok := input["note"].(string)
	if !ok || noteContent == "" {
		return "", fmt.Errorf("note content is required")
	}

	// Verify advisor has access to this client
	var accessLevel string
	err := db.DB.QueryRow(`
		SELECT access_level FROM advisor_clients
		WHERE advisor_id = ? AND client_id = ? AND status = 'active'
	`, e.UserID, clientID).Scan(&accessLevel)
	if err != nil {
		return "", fmt.Errorf("you don't have access to this client")
	}

	// Get optional parameters
	category := "general"
	if cat, ok := input["category"].(string); ok && cat != "" {
		validCategories := map[string]bool{
			"general": true, "meeting": true, "goal": true,
			"concern": true, "action_item": true, "personal": true,
		}
		if validCategories[cat] {
			category = cat
		}
	}

	isPinned := false
	if pinned, ok := input["is_pinned"].(bool); ok {
		isPinned = pinned
	}

	// Insert the note
	result, err := db.DB.Exec(`
		INSERT INTO client_notes (advisor_id, client_id, note, category, is_pinned)
		VALUES (?, ?, ?, ?, ?)
	`, e.UserID, clientID, noteContent, category, isPinned)
	if err != nil {
		return "", fmt.Errorf("failed to create note: %w", err)
	}

	noteID, _ := result.LastInsertId()

	// Get client name for response
	var clientName string
	db.DB.QueryRow(`SELECT name FROM users WHERE id = ?`, clientID).Scan(&clientName)

	response := map[string]interface{}{
		"status":      "created",
		"note_id":     noteID,
		"client_id":   clientID,
		"client_name": clientName,
		"note":        noteContent,
		"category":    category,
		"is_pinned":   isPinned,
		"message":     fmt.Sprintf("Note added to client %s", clientName),
	}

	jsonBytes, _ := json.MarshalIndent(response, "", "  ")
	return string(jsonBytes), nil
}

// updateClientNote updates an existing note (advisor only)
func (e *ToolExecutor) updateClientNote(input map[string]interface{}) (string, error) {
	if !e.IsAdvisor {
		return "", fmt.Errorf("this tool is only available to advisors")
	}

	// Get note ID (required)
	noteID, ok := input["note_id"].(float64)
	if !ok {
		return "", fmt.Errorf("note_id is required")
	}

	// Verify advisor owns this note
	var existingNote, existingCategory string
	var existingPinned bool
	var clientID int
	err := db.DB.QueryRow(`
		SELECT note, category, is_pinned, client_id
		FROM client_notes
		WHERE id = ? AND advisor_id = ?
	`, int(noteID), e.UserID).Scan(&existingNote, &existingCategory, &existingPinned, &clientID)
	if err != nil {
		return "", fmt.Errorf("note not found or you don't have permission to edit it")
	}

	// Apply updates
	if newNote, ok := input["note"].(string); ok && newNote != "" {
		existingNote = newNote
	}
	if newCategory, ok := input["category"].(string); ok && newCategory != "" {
		validCategories := map[string]bool{
			"general": true, "meeting": true, "goal": true,
			"concern": true, "action_item": true, "personal": true,
		}
		if validCategories[newCategory] {
			existingCategory = newCategory
		}
	}
	if newPinned, ok := input["is_pinned"].(bool); ok {
		existingPinned = newPinned
	}

	// Update the note
	_, err = db.DB.Exec(`
		UPDATE client_notes
		SET note = ?, category = ?, is_pinned = ?
		WHERE id = ?
	`, existingNote, existingCategory, existingPinned, int(noteID))
	if err != nil {
		return "", fmt.Errorf("failed to update note: %w", err)
	}

	// Get client name for response
	var clientName string
	db.DB.QueryRow(`SELECT name FROM users WHERE id = ?`, clientID).Scan(&clientName)

	response := map[string]interface{}{
		"status":      "updated",
		"note_id":     int(noteID),
		"client_id":   clientID,
		"client_name": clientName,
		"note":        existingNote,
		"category":    existingCategory,
		"is_pinned":   existingPinned,
		"message":     "Note updated successfully",
	}

	jsonBytes, _ := json.MarshalIndent(response, "", "  ")
	return string(jsonBytes), nil
}

// deleteClientNote deletes a note (advisor only)
func (e *ToolExecutor) deleteClientNote(input map[string]interface{}) (string, error) {
	if !e.IsAdvisor {
		return "", fmt.Errorf("this tool is only available to advisors")
	}

	// Get note ID (required)
	noteID, ok := input["note_id"].(float64)
	if !ok {
		return "", fmt.Errorf("note_id is required")
	}

	// Get client info before deleting
	var clientID int
	var clientName string
	err := db.DB.QueryRow(`
		SELECT cn.client_id, u.name
		FROM client_notes cn
		JOIN users u ON cn.client_id = u.id
		WHERE cn.id = ? AND cn.advisor_id = ?
	`, int(noteID), e.UserID).Scan(&clientID, &clientName)
	if err != nil {
		return "", fmt.Errorf("note not found or you don't have permission to delete it")
	}

	// Delete the note
	result, err := db.DB.Exec(`
		DELETE FROM client_notes
		WHERE id = ? AND advisor_id = ?
	`, int(noteID), e.UserID)
	if err != nil {
		return "", fmt.Errorf("failed to delete note: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return "", fmt.Errorf("note not found")
	}

	response := map[string]interface{}{
		"status":      "deleted",
		"note_id":     int(noteID),
		"client_id":   clientID,
		"client_name": clientName,
		"message":     fmt.Sprintf("Note deleted from client %s", clientName),
	}

	jsonBytes, _ := json.MarshalIndent(response, "", "  ")
	return string(jsonBytes), nil
}
