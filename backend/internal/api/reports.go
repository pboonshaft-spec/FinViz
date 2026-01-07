package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/finviz/backend/internal/db"
	"github.com/finviz/backend/internal/models"
	"github.com/finviz/backend/internal/reports"
	"github.com/finviz/backend/internal/simulation"
)

// ReportRequest contains parameters for generating a report
type ReportRequest struct {
	IncludeSimulation bool                     `json:"includeSimulation"`
	SimulationParams  *models.SimulationParams `json:"simulationParams,omitempty"`
}

// handleGenerateReport generates a PDF financial plan report
func handleGenerateReport(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	userID := getEffectiveUserID(r)
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Unable to determine user")
		return
	}

	// Parse request body
	var req ReportRequest
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
	}

	// Get client name (either the client being viewed or the user themselves)
	clientName := user.Name
	advisorName := ""
	if client := getClientContext(r); client != nil {
		clientName = client.Name
		advisorName = user.Name
	}

	// Fetch assets with types
	assets, err := fetchUserAssets(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch assets")
		return
	}

	// Fetch debts
	debts, err := fetchUserDebts(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch debts")
		return
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
		ClientName:  clientName,
		AdvisorName: advisorName,
		GeneratedAt: time.Now(),
		Assets:      assets,
		Debts:       debts,
		TotalAssets: totalAssets,
		TotalDebts:  totalDebts,
		NetWorth:    netWorth,
	}

	// Run simulation if requested
	if req.IncludeSimulation {
		params := models.DefaultSimulationParams()
		if req.SimulationParams != nil {
			params = *req.SimulationParams
		}
		params.ApplyDefaults()

		simResult := simulation.RunMonteCarloWithParams(assets, debts, &params)
		reportData.Simulation = &simResult
		reportData.Params = &params
	}

	// Generate PDF
	pdfBytes, err := reports.GenerateFinancialPlanReport(reportData)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate PDF: %v", err))
		return
	}

	// Generate filename
	filename := fmt.Sprintf("financial_plan_%s_%s.pdf",
		sanitizeFilename(clientName),
		time.Now().Format("2006-01-02"))

	// Send PDF response
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))
	w.WriteHeader(http.StatusOK)
	w.Write(pdfBytes)
}

// fetchUserAssets retrieves all assets with their types for a user
func fetchUserAssets(userID int) ([]models.Asset, error) {
	rows, err := db.DB.Query(`
		SELECT a.id, a.user_id, a.name, a.type_id, a.current_value,
			   a.custom_return, a.custom_volatility, a.plaid_account_id, a.created_at, a.updated_at,
			   t.id, t.name, t.default_return, t.default_volatility
		FROM assets a
		LEFT JOIN asset_types t ON a.type_id = t.id
		WHERE a.user_id = ?
		ORDER BY a.current_value DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []models.Asset
	for rows.Next() {
		var a models.Asset
		var t models.AssetType
		var typeCreatedAt time.Time

		err := rows.Scan(
			&a.ID, &a.UserID, &a.Name, &a.TypeID, &a.CurrentValue,
			&a.CustomReturn, &a.CustomVolatility, &a.PlaidAccountID, &a.CreatedAt, &a.UpdatedAt,
			&t.ID, &t.Name, &t.DefaultReturn, &t.DefaultVolatility,
		)
		if err != nil {
			return nil, err
		}
		// Skip type created_at since we don't need it and it simplifies the query
		_ = typeCreatedAt
		a.AssetType = &t
		assets = append(assets, a)
	}

	return assets, nil
}

// fetchUserDebts retrieves all debts for a user
func fetchUserDebts(userID int) ([]models.Debt, error) {
	rows, err := db.DB.Query(`
		SELECT id, user_id, name, current_balance, interest_rate, minimum_payment,
			   plaid_account_id, created_at, updated_at
		FROM debts
		WHERE user_id = ?
		ORDER BY current_balance DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var debts []models.Debt
	for rows.Next() {
		var d models.Debt
		err := rows.Scan(
			&d.ID, &d.UserID, &d.Name, &d.CurrentBalance, &d.InterestRate,
			&d.MinimumPayment, &d.PlaidAccountID, &d.CreatedAt, &d.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		debts = append(debts, d)
	}

	return debts, nil
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
