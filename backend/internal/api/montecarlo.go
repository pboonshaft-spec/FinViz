package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/finviz/backend/internal/db"
	"github.com/finviz/backend/internal/models"
	"github.com/finviz/backend/internal/simulation"
)

func handleMonteCarlo(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req models.MonteCarloRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Initialize params with defaults if not provided
	params := req.Params
	if params == nil {
		defaultParams := models.DefaultSimulationParams()
		params = &defaultParams
	}

	// Validate time horizon
	if params.TimeHorizonYears > 80 {
		respondError(w, http.StatusBadRequest, "Time horizon must be 80 years or less")
		return
	}

	// Validate ages if provided
	if params.CurrentAge > 0 && params.RetirementAge > 0 {
		if params.RetirementAge < params.CurrentAge {
			respondError(w, http.StatusBadRequest, "Retirement age must be greater than current age")
			return
		}
	}

	// Fetch all assets with their types for this user
	assets, err := fetchAssetsWithTypesForUser(user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Fetch all debts for this user
	debts, err := fetchDebtsForUser(user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Filter out credit card debt if requested
	if params.ExcludeCreditCardDebt {
		debts = filterOutCreditCardDebt(debts)
	}

	result := simulation.RunMonteCarloWithParams(assets, debts, params)
	respondJSON(w, http.StatusOK, result)
}

func fetchAssetsWithTypesForUser(userID int) ([]models.Asset, error) {
	rows, err := db.DB.Query(`
		SELECT a.id, a.name, a.type_id, a.current_value, a.custom_return, a.custom_volatility,
		       a.created_at, a.updated_at, t.id, t.name, t.default_return, t.default_volatility
		FROM assets a
		JOIN asset_types t ON a.type_id = t.id
		WHERE a.user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []models.Asset
	for rows.Next() {
		var a models.Asset
		var t models.AssetType
		if err := rows.Scan(
			&a.ID, &a.Name, &a.TypeID, &a.CurrentValue, &a.CustomReturn, &a.CustomVolatility,
			&a.CreatedAt, &a.UpdatedAt, &t.ID, &t.Name, &t.DefaultReturn, &t.DefaultVolatility,
		); err != nil {
			return nil, err
		}
		a.AssetType = &t
		assets = append(assets, a)
	}

	return assets, nil
}

func fetchDebtsForUser(userID int) ([]models.Debt, error) {
	rows, err := db.DB.Query(`
		SELECT id, name, current_balance, interest_rate, minimum_payment, created_at, updated_at
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
		if err := rows.Scan(&d.ID, &d.Name, &d.CurrentBalance, &d.InterestRate, &d.MinimumPayment, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		debts = append(debts, d)
	}

	return debts, nil
}

// filterOutCreditCardDebt removes credit card debt from the list
// Credit cards are identified by keywords in the name
func filterOutCreditCardDebt(debts []models.Debt) []models.Debt {
	filtered := make([]models.Debt, 0, len(debts))
	for _, d := range debts {
		if !isCreditCardDebt(d.Name) {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

// isCreditCardDebt checks if a debt name indicates a credit card
func isCreditCardDebt(name string) bool {
	lower := strings.ToLower(name)
	creditKeywords := []string{
		"credit card", "credit", "card", "visa", "mastercard",
		"amex", "american express", "discover", "chase sapphire",
		"capital one", "citi", "barclays",
	}
	for _, kw := range creditKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}
