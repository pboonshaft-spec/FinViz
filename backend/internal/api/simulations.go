package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/finviz/backend/internal/db"
	"github.com/finviz/backend/internal/models"
)

// SimulationSaveRequest is the request body for saving a simulation
type SimulationSaveRequest struct {
	Params           models.SimulationParams   `json:"params"`
	Results          models.MonteCarloResponse `json:"results"`
	Name             *string                   `json:"name,omitempty"`
	Notes            *string                   `json:"notes,omitempty"`
}

// SimulationUpdateRequest is the request body for updating a simulation
type SimulationUpdateRequest struct {
	Name       *string `json:"name,omitempty"`
	Notes      *string `json:"notes,omitempty"`
	IsFavorite *bool   `json:"isFavorite,omitempty"`
}

// handleListSimulations returns a list of saved simulations for the user
func handleListSimulations(w http.ResponseWriter, r *http.Request) {
	userID := getEffectiveUserID(r)
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Parse query params
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offsetStr := r.URL.Query().Get("offset")
	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	favoritesOnly := r.URL.Query().Get("favorites") == "true"

	// Build query
	query := `
		SELECT sh.id, sh.name, sh.starting_net_worth, sh.final_p50, sh.success_rate,
		       sh.time_horizon_years, sh.is_favorite, sh.created_at,
		       COALESCE(u.name, '') as run_by_user_name
		FROM simulation_history sh
		LEFT JOIN users u ON sh.run_by_user_id = u.id
		WHERE sh.user_id = ?
	`
	args := []interface{}{userID}

	if favoritesOnly {
		query += " AND sh.is_favorite = TRUE"
	}

	query += " ORDER BY sh.created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch simulations")
		return
	}
	defer rows.Close()

	simulations := []models.SimulationHistorySummary{}
	for rows.Next() {
		var sim models.SimulationHistorySummary
		err := rows.Scan(
			&sim.ID, &sim.Name, &sim.StartingNetWorth, &sim.FinalP50,
			&sim.SuccessRate, &sim.TimeHorizonYears, &sim.IsFavorite,
			&sim.CreatedAt, &sim.RunByUserName,
		)
		if err != nil {
			continue
		}
		simulations = append(simulations, sim)
	}

	respondJSON(w, http.StatusOK, simulations)
}

// handleGetSimulation returns the full details of a specific simulation
func handleGetSimulation(w http.ResponseWriter, r *http.Request) {
	userID := getEffectiveUserID(r)
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	simIDStr := r.PathValue("id")
	simID, err := strconv.Atoi(simIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid simulation ID")
		return
	}

	var sim models.SimulationHistory
	var runByUserName string
	err = db.DB.QueryRow(`
		SELECT sh.id, sh.user_id, sh.run_by_user_id, sh.name, sh.notes,
		       sh.params, sh.results, sh.starting_net_worth, sh.final_p50,
		       sh.success_rate, sh.time_horizon_years, sh.is_favorite, sh.created_at,
		       COALESCE(u.name, '') as run_by_user_name
		FROM simulation_history sh
		LEFT JOIN users u ON sh.run_by_user_id = u.id
		WHERE sh.id = ? AND sh.user_id = ?
	`, simID, userID).Scan(
		&sim.ID, &sim.UserID, &sim.RunByUserID, &sim.Name, &sim.Notes,
		&sim.Params, &sim.Results, &sim.StartingNetWorth, &sim.FinalP50,
		&sim.SuccessRate, &sim.TimeHorizonYears, &sim.IsFavorite, &sim.CreatedAt,
		&runByUserName,
	)

	if err != nil {
		respondError(w, http.StatusNotFound, "Simulation not found")
		return
	}

	// Parse the JSON fields
	var params models.SimulationParams
	var results models.MonteCarloResponse
	json.Unmarshal([]byte(sim.Params), &params)
	json.Unmarshal([]byte(sim.Results), &results)

	response := models.SimulationHistoryFull{
		SimulationHistory: sim,
		ParsedParams:      &params,
		ParsedResults:     &results,
	}
	if runByUserName != "" {
		response.RunByUser = &models.User{Name: runByUserName}
	}

	respondJSON(w, http.StatusOK, response)
}

// handleSaveSimulation saves a new simulation to history
func handleSaveSimulation(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	targetUserID := getEffectiveUserID(r)

	var req SimulationSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Serialize params and results to JSON
	paramsJSON, err := json.Marshal(req.Params)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to serialize params")
		return
	}

	resultsJSON, err := json.Marshal(req.Results)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to serialize results")
		return
	}

	// Insert into database
	result, err := db.DB.Exec(`
		INSERT INTO simulation_history
		(user_id, run_by_user_id, name, notes, params, results,
		 starting_net_worth, final_p50, success_rate, time_horizon_years)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		targetUserID,
		user.ID, // The person running it (could be advisor)
		req.Name,
		req.Notes,
		string(paramsJSON),
		string(resultsJSON),
		req.Results.Summary.StartingNetWorth,
		req.Results.Summary.FinalP50,
		req.Results.Summary.SuccessRate,
		req.Params.TimeHorizonYears,
	)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save simulation")
		return
	}

	id, _ := result.LastInsertId()

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"id":      id,
		"message": "Simulation saved successfully",
	})
}

// handleUpdateSimulation updates a simulation's metadata
func handleUpdateSimulation(w http.ResponseWriter, r *http.Request) {
	userID := getEffectiveUserID(r)
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	simIDStr := r.PathValue("id")
	simID, err := strconv.Atoi(simIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid simulation ID")
		return
	}

	var req SimulationUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Verify ownership
	var exists int
	err = db.DB.QueryRow(
		"SELECT COUNT(*) FROM simulation_history WHERE id = ? AND user_id = ?",
		simID, userID,
	).Scan(&exists)
	if err != nil || exists == 0 {
		respondError(w, http.StatusNotFound, "Simulation not found")
		return
	}

	// Build update query dynamically
	updates := []string{}
	args := []interface{}{}

	if req.Name != nil {
		updates = append(updates, "name = ?")
		args = append(args, *req.Name)
	}
	if req.Notes != nil {
		updates = append(updates, "notes = ?")
		args = append(args, *req.Notes)
	}
	if req.IsFavorite != nil {
		updates = append(updates, "is_favorite = ?")
		args = append(args, *req.IsFavorite)
	}

	if len(updates) == 0 {
		respondError(w, http.StatusBadRequest, "No updates provided")
		return
	}

	query := "UPDATE simulation_history SET "
	for i, u := range updates {
		if i > 0 {
			query += ", "
		}
		query += u
	}
	query += " WHERE id = ? AND user_id = ?"
	args = append(args, simID, userID)

	_, err = db.DB.Exec(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update simulation")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Simulation updated"})
}

// handleDeleteSimulation deletes a simulation from history
func handleDeleteSimulation(w http.ResponseWriter, r *http.Request) {
	userID := getEffectiveUserID(r)
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	simIDStr := r.PathValue("id")
	simID, err := strconv.Atoi(simIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid simulation ID")
		return
	}

	result, err := db.DB.Exec(
		"DELETE FROM simulation_history WHERE id = ? AND user_id = ?",
		simID, userID,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete simulation")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		respondError(w, http.StatusNotFound, "Simulation not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Simulation deleted"})
}
