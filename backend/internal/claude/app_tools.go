package claude

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/finviz/backend/internal/db"
)

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	return math.Abs(x)
}

// ToolExecutor handles execution of tools for a specific user
type ToolExecutor struct {
	UserID int
}

// NewToolExecutor creates a new tool executor for a user
func NewToolExecutor(userID int) *ToolExecutor {
	return &ToolExecutor{UserID: userID}
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
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

// getUserAssets fetches all assets for the user
func (e *ToolExecutor) getUserAssets() (string, error) {
	rows, err := db.DB.Query(`
		SELECT a.id, a.name, at.name as type_name, a.current_value,
			   COALESCE(a.custom_return, at.default_return) as expected_return,
			   COALESCE(a.custom_volatility, at.default_volatility) as volatility
		FROM assets a
		LEFT JOIN asset_types at ON a.type_id = at.id
		WHERE a.user_id = ?
		ORDER BY a.current_value DESC
	`, e.UserID)
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
	rows, err := db.DB.Query(`
		SELECT id, name, current_balance, interest_rate, minimum_payment
		FROM debts
		WHERE user_id = ?
		ORDER BY current_balance DESC
	`, e.UserID)
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
	args := []interface{}{e.UserID, startDate, endDate}

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
	// Get total assets by type
	assetRows, err := db.DB.Query(`
		SELECT COALESCE(at.name, 'Other') as type_name, SUM(a.current_value) as total
		FROM assets a
		LEFT JOIN asset_types at ON a.type_id = at.id
		WHERE a.user_id = ?
		GROUP BY at.name
		ORDER BY total DESC
	`, e.UserID)
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
	err = db.DB.QueryRow(`SELECT COALESCE(SUM(current_balance), 0) FROM debts WHERE user_id = ?`, e.UserID).Scan(&totalDebts)
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
	`, e.UserID, startDate)
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
	`, e.UserID, startDate)
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
