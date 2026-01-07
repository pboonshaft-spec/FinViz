package api

import (
	"encoding/json"
	"fmt"
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

	// Check if advisor has permission to run simulations
	if isActingAsAdvisor(r) && !canRunSimulations(r) {
		respondError(w, http.StatusForbidden, "No permission to run simulations for this client")
		return
	}

	// Get the effective user ID (client ID if advisor is acting on behalf of client)
	targetUserID := getEffectiveUserID(r)

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

	// Fetch all assets with their types for the target user
	assets, err := fetchAssetsWithTypesForUser(targetUserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Fetch all debts for the target user
	debts, err := fetchDebtsForUser(targetUserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Filter out credit card debt if requested
	if params.ExcludeCreditCardDebt {
		debts = filterOutCreditCardDebt(debts)
	}

	result := simulation.RunMonteCarloWithParams(assets, debts, params)

	// Save the simulation if requested
	if req.SaveResult {
		paramsJSON, _ := json.Marshal(params)
		resultsJSON, _ := json.Marshal(result)

		_, err := db.DB.Exec(`
			INSERT INTO simulation_history
			(user_id, run_by_user_id, name, notes, params, results,
			 starting_net_worth, final_p50, success_rate, time_horizon_years)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			targetUserID,
			user.ID,
			req.Name,
			req.Notes,
			string(paramsJSON),
			string(resultsJSON),
			result.Summary.StartingNetWorth,
			result.Summary.FinalP50,
			result.Summary.SuccessRate,
			params.TimeHorizonYears,
		)

		if err != nil {
			// Log but don't fail the request - simulation was successful
			// Just couldn't save to history
		}
	}

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

// handleScenarioComparison runs multiple scenarios and compares them
func handleScenarioComparison(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Check if advisor has permission to run simulations
	if isActingAsAdvisor(r) && !canRunSimulations(r) {
		respondError(w, http.StatusForbidden, "No permission to run simulations for this client")
		return
	}

	// Get the effective user ID
	targetUserID := getEffectiveUserID(r)

	var req models.ScenarioComparisonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if len(req.Scenarios) < 2 {
		respondError(w, http.StatusBadRequest, "At least 2 scenarios are required for comparison")
		return
	}
	if len(req.Scenarios) > 5 {
		respondError(w, http.StatusBadRequest, "Maximum 5 scenarios allowed")
		return
	}

	// Fetch assets and debts once
	assets, err := fetchAssetsWithTypesForUser(targetUserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	debts, err := fetchDebtsForUser(targetUserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Run each scenario
	results := make([]models.ScenarioResult, len(req.Scenarios))
	for i, scenario := range req.Scenarios {
		params := scenario.Params
		if params == nil {
			defaultParams := models.DefaultSimulationParams()
			params = &defaultParams
		}

		// Filter debts if requested
		scenarioDebts := debts
		if params.ExcludeCreditCardDebt {
			scenarioDebts = filterOutCreditCardDebt(debts)
		}

		result := simulation.RunMonteCarloWithParams(assets, scenarioDebts, params)
		results[i] = models.ScenarioResult{
			Name:        scenario.Name,
			Summary:     result.Summary,
			Projections: result.Projections,
		}
	}

	// Generate comparisons between all pairs
	comparisons := generateScenarioComparisons(results)

	// Find best scenario (highest success rate)
	bestScenario := results[0].Name
	bestRate := results[0].Summary.SuccessRate
	for _, r := range results[1:] {
		if r.Summary.SuccessRate > bestRate {
			bestRate = r.Summary.SuccessRate
			bestScenario = r.Name
		}
	}

	response := models.ScenarioComparisonResponse{
		Scenarios:    results,
		Comparisons:  comparisons,
		BestScenario: bestScenario,
	}

	respondJSON(w, http.StatusOK, response)
}

// generateScenarioComparisons creates comparison objects for all scenario pairs
func generateScenarioComparisons(results []models.ScenarioResult) []models.ScenarioDiff {
	var comparisons []models.ScenarioDiff

	// Compare each pair (first scenario vs all others)
	baseline := results[0]
	for i := 1; i < len(results); i++ {
		alt := results[i]

		successDiff := alt.Summary.SuccessRate - baseline.Summary.SuccessRate
		p50Diff := alt.Summary.FinalP50 - baseline.Summary.FinalP50
		contribDiff := alt.Summary.TotalContributions - baseline.Summary.TotalContributions

		rec := generateRecommendation(baseline.Name, alt.Name, successDiff, p50Diff, contribDiff)

		comparisons = append(comparisons, models.ScenarioDiff{
			ScenarioA:         baseline.Name,
			ScenarioB:         alt.Name,
			SuccessRateDiff:   successDiff,
			FinalP50Diff:      p50Diff,
			ContributionsDiff: contribDiff,
			Recommendation:    rec,
		})
	}

	return comparisons
}

// generateRecommendation creates a human-readable recommendation based on differences
func generateRecommendation(nameA, nameB string, successDiff, p50Diff, contribDiff float64) string {
	if successDiff > 10 {
		return nameB + " significantly improves your success rate by " + formatPercent(successDiff) + ". Strongly consider this option."
	} else if successDiff > 5 {
		return nameB + " improves your success rate by " + formatPercent(successDiff) + ". Worth considering."
	} else if successDiff > 0 {
		return nameB + " slightly improves success rate (" + formatPercent(successDiff) + "). Minor improvement."
	} else if successDiff < -10 {
		return nameB + " reduces success rate by " + formatPercent(-successDiff) + ". Not recommended."
	} else if successDiff < -5 {
		return nameB + " moderately reduces success rate. Consider trade-offs carefully."
	} else if successDiff < 0 {
		return nameB + " has slightly lower success rate, but may have other benefits."
	}

	// Similar success rates - look at wealth
	if p50Diff > 100000 {
		return "Similar success rates, but " + nameB + " results in significantly higher expected wealth."
	} else if p50Diff < -100000 {
		return "Similar success rates, but " + nameA + " results in higher expected wealth."
	}

	return "Both scenarios have similar outcomes. Choose based on personal preference."
}

// formatPercent formats a percentage for display
func formatPercent(val float64) string {
	if val >= 0 {
		return fmt.Sprintf("+%.1f%%", val)
	}
	return fmt.Sprintf("%.1f%%", val)
}
