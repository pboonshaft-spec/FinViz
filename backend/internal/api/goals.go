package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/finviz/backend/internal/db"
	"github.com/finviz/backend/internal/models"
)

// handleListGoals returns goals for a client (accessible by both advisor and client)
func handleListGoals(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var clientID int
	var err error

	if user.IsAdvisor() {
		// Advisor viewing client goals
		clientID, err = strconv.Atoi(r.PathValue("clientId"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid client ID")
			return
		}
		// Verify advisor has access to this client
		if !advisorHasClientAccess(user.ID, clientID) {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
	} else {
		// Client viewing their own goals
		clientID = user.ID
	}

	// Optional status filter
	status := r.URL.Query().Get("status")

	var goals []models.ClientGoal
	var query string
	var args []interface{}

	if status != "" {
		query = `SELECT id, advisor_id, client_id, title, description, category, status, priority,
			target_amount, current_amount, target_date, completed_at, created_at, updated_at
			FROM client_goals
			WHERE client_id = ? AND status = ?
			ORDER BY
				CASE priority WHEN 'high' THEN 1 WHEN 'medium' THEN 2 ELSE 3 END,
				created_at DESC`
		args = []interface{}{clientID, status}
	} else {
		query = `SELECT id, advisor_id, client_id, title, description, category, status, priority,
			target_amount, current_amount, target_date, completed_at, created_at, updated_at
			FROM client_goals
			WHERE client_id = ?
			ORDER BY
				CASE status WHEN 'in_progress' THEN 1 WHEN 'pending' THEN 2 WHEN 'on_hold' THEN 3 ELSE 4 END,
				CASE priority WHEN 'high' THEN 1 WHEN 'medium' THEN 2 ELSE 3 END,
				created_at DESC`
		args = []interface{}{clientID}
	}

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch goals")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var goal models.ClientGoal
		var description, targetDate sql.NullString
		var targetAmount, currentAmount sql.NullFloat64
		var completedAt sql.NullTime

		err := rows.Scan(
			&goal.ID, &goal.AdvisorID, &goal.ClientID, &goal.Title,
			&description, &goal.Category, &goal.Status, &goal.Priority,
			&targetAmount, &currentAmount, &targetDate, &completedAt,
			&goal.CreatedAt, &goal.UpdatedAt,
		)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to parse goals")
			return
		}

		if description.Valid {
			goal.Description = &description.String
		}
		if targetDate.Valid {
			goal.TargetDate = &targetDate.String
		}
		if targetAmount.Valid {
			goal.TargetAmount = &targetAmount.Float64
		}
		if currentAmount.Valid {
			goal.CurrentAmount = &currentAmount.Float64
		}
		if completedAt.Valid {
			goal.CompletedAt = &completedAt.Time
		}

		goals = append(goals, goal)
	}

	if goals == nil {
		goals = []models.ClientGoal{}
	}

	respondJSON(w, http.StatusOK, goals)
}

// handleCreateGoal creates a new goal (advisor only)
func handleCreateGoal(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil || !user.IsAdvisor() {
		respondError(w, http.StatusUnauthorized, "Only advisors can create goals")
		return
	}

	clientID, err := strconv.Atoi(r.PathValue("clientId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid client ID")
		return
	}

	// Verify advisor has access to this client
	if !advisorHasClientAccess(user.ID, clientID) {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	var req models.CreateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Title == "" {
		respondError(w, http.StatusBadRequest, "Goal title is required")
		return
	}

	// Default category and priority
	if req.Category == "" {
		req.Category = models.GoalCategoryOther
	}
	if req.Priority == "" {
		req.Priority = models.GoalPriorityMedium
	}

	// Validate category
	validCategories := map[string]bool{
		models.GoalCategoryRetirement:    true,
		models.GoalCategorySavings:       true,
		models.GoalCategoryDebt:          true,
		models.GoalCategoryInvestment:    true,
		models.GoalCategoryEducation:     true,
		models.GoalCategoryEmergency:     true,
		models.GoalCategoryMajorPurchase: true,
		models.GoalCategoryOther:         true,
	}
	if !validCategories[req.Category] {
		respondError(w, http.StatusBadRequest, "Invalid category")
		return
	}

	// Validate priority
	validPriorities := map[string]bool{
		models.GoalPriorityLow:    true,
		models.GoalPriorityMedium: true,
		models.GoalPriorityHigh:   true,
	}
	if !validPriorities[req.Priority] {
		respondError(w, http.StatusBadRequest, "Invalid priority")
		return
	}

	// Parse target date if provided
	var targetDate *string
	if req.TargetDate != "" {
		targetDate = &req.TargetDate
	}

	result, err := db.DB.Exec(
		`INSERT INTO client_goals (advisor_id, client_id, title, description, category, priority, target_amount, current_amount, target_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID, clientID, req.Title, req.Description, req.Category, req.Priority,
		req.TargetAmount, req.CurrentAmount, targetDate,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create goal")
		return
	}

	goalID, _ := result.LastInsertId()

	// Fetch the created goal
	goal, err := getGoalByID(int(goalID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch created goal")
		return
	}

	respondJSON(w, http.StatusCreated, goal)
}

// handleUpdateGoal updates an existing goal (advisor only for most fields, client can update progress)
func handleUpdateGoal(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var clientID int
	var err error

	if user.IsAdvisor() {
		clientID, err = strconv.Atoi(r.PathValue("clientId"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid client ID")
			return
		}
		if !advisorHasClientAccess(user.ID, clientID) {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
	} else {
		clientID = user.ID
	}

	goalID, err := strconv.Atoi(r.PathValue("goalId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid goal ID")
		return
	}

	// Fetch existing goal
	existingGoal, err := getGoalByID(goalID)
	if err != nil || existingGoal.ClientID != clientID {
		respondError(w, http.StatusNotFound, "Goal not found")
		return
	}

	var req models.UpdateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Clients can only update current_amount (progress)
	if !user.IsAdvisor() {
		if req.CurrentAmount != nil {
			_, err = db.DB.Exec(
				`UPDATE client_goals SET current_amount = ? WHERE id = ?`,
				*req.CurrentAmount, goalID,
			)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to update goal")
				return
			}
		}
	} else {
		// Advisors can update all fields
		if req.Title != "" {
			existingGoal.Title = req.Title
		}
		if req.Description != nil {
			existingGoal.Description = req.Description
		}
		if req.Category != "" {
			existingGoal.Category = req.Category
		}
		if req.Priority != "" {
			existingGoal.Priority = req.Priority
		}
		if req.TargetAmount != nil {
			existingGoal.TargetAmount = req.TargetAmount
		}
		if req.CurrentAmount != nil {
			existingGoal.CurrentAmount = req.CurrentAmount
		}
		if req.TargetDate != nil {
			existingGoal.TargetDate = req.TargetDate
		}

		// Handle status change
		if req.Status != "" && req.Status != existingGoal.Status {
			existingGoal.Status = req.Status
			if req.Status == models.GoalStatusCompleted {
				now := time.Now()
				_, err = db.DB.Exec(
					`UPDATE client_goals SET title = ?, description = ?, category = ?, status = ?, priority = ?,
					target_amount = ?, current_amount = ?, target_date = ?, completed_at = ? WHERE id = ?`,
					existingGoal.Title, existingGoal.Description, existingGoal.Category, existingGoal.Status,
					existingGoal.Priority, existingGoal.TargetAmount, existingGoal.CurrentAmount,
					existingGoal.TargetDate, now, goalID,
				)
			} else {
				_, err = db.DB.Exec(
					`UPDATE client_goals SET title = ?, description = ?, category = ?, status = ?, priority = ?,
					target_amount = ?, current_amount = ?, target_date = ?, completed_at = NULL WHERE id = ?`,
					existingGoal.Title, existingGoal.Description, existingGoal.Category, existingGoal.Status,
					existingGoal.Priority, existingGoal.TargetAmount, existingGoal.CurrentAmount,
					existingGoal.TargetDate, goalID,
				)
			}
		} else {
			_, err = db.DB.Exec(
				`UPDATE client_goals SET title = ?, description = ?, category = ?, priority = ?,
				target_amount = ?, current_amount = ?, target_date = ? WHERE id = ?`,
				existingGoal.Title, existingGoal.Description, existingGoal.Category,
				existingGoal.Priority, existingGoal.TargetAmount, existingGoal.CurrentAmount,
				existingGoal.TargetDate, goalID,
			)
		}

		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to update goal")
			return
		}
	}

	// Fetch updated goal
	updatedGoal, err := getGoalByID(goalID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch updated goal")
		return
	}

	respondJSON(w, http.StatusOK, updatedGoal)
}

// handleDeleteGoal deletes a goal (advisor only)
func handleDeleteGoal(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil || !user.IsAdvisor() {
		respondError(w, http.StatusUnauthorized, "Only advisors can delete goals")
		return
	}

	clientID, err := strconv.Atoi(r.PathValue("clientId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid client ID")
		return
	}

	goalID, err := strconv.Atoi(r.PathValue("goalId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid goal ID")
		return
	}

	// Verify advisor owns this goal (created it for their client)
	result, err := db.DB.Exec(
		`DELETE FROM client_goals WHERE id = ? AND advisor_id = ? AND client_id = ?`,
		goalID, user.ID, clientID,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete goal")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		respondError(w, http.StatusNotFound, "Goal not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Goal deleted successfully"})
}

// handleGetMyGoals returns goals for the current user (client endpoint)
func handleGetMyGoals(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	status := r.URL.Query().Get("status")

	var goals []models.ClientGoal
	var query string
	var args []interface{}

	if status != "" {
		query = `SELECT id, advisor_id, client_id, title, description, category, status, priority,
			target_amount, current_amount, target_date, completed_at, created_at, updated_at
			FROM client_goals
			WHERE client_id = ? AND status = ?
			ORDER BY
				CASE priority WHEN 'high' THEN 1 WHEN 'medium' THEN 2 ELSE 3 END,
				created_at DESC`
		args = []interface{}{user.ID, status}
	} else {
		query = `SELECT id, advisor_id, client_id, title, description, category, status, priority,
			target_amount, current_amount, target_date, completed_at, created_at, updated_at
			FROM client_goals
			WHERE client_id = ?
			ORDER BY
				CASE status WHEN 'in_progress' THEN 1 WHEN 'pending' THEN 2 WHEN 'on_hold' THEN 3 ELSE 4 END,
				CASE priority WHEN 'high' THEN 1 WHEN 'medium' THEN 2 ELSE 3 END,
				created_at DESC`
		args = []interface{}{user.ID}
	}

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch goals")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var goal models.ClientGoal
		var description, targetDate sql.NullString
		var targetAmount, currentAmount sql.NullFloat64
		var completedAt sql.NullTime

		err := rows.Scan(
			&goal.ID, &goal.AdvisorID, &goal.ClientID, &goal.Title,
			&description, &goal.Category, &goal.Status, &goal.Priority,
			&targetAmount, &currentAmount, &targetDate, &completedAt,
			&goal.CreatedAt, &goal.UpdatedAt,
		)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to parse goals")
			return
		}

		if description.Valid {
			goal.Description = &description.String
		}
		if targetDate.Valid {
			goal.TargetDate = &targetDate.String
		}
		if targetAmount.Valid {
			goal.TargetAmount = &targetAmount.Float64
		}
		if currentAmount.Valid {
			goal.CurrentAmount = &currentAmount.Float64
		}
		if completedAt.Valid {
			goal.CompletedAt = &completedAt.Time
		}

		goals = append(goals, goal)
	}

	if goals == nil {
		goals = []models.ClientGoal{}
	}

	respondJSON(w, http.StatusOK, goals)
}

// handleUpdateMyGoalProgress allows a client to update their goal progress
func handleUpdateMyGoalProgress(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	goalID, err := strconv.Atoi(r.PathValue("goalId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid goal ID")
		return
	}

	// Verify goal belongs to this user
	existingGoal, err := getGoalByID(goalID)
	if err != nil || existingGoal.ClientID != user.ID {
		respondError(w, http.StatusNotFound, "Goal not found")
		return
	}

	var req struct {
		CurrentAmount float64 `json:"currentAmount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	_, err = db.DB.Exec(
		`UPDATE client_goals SET current_amount = ? WHERE id = ?`,
		req.CurrentAmount, goalID,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update goal progress")
		return
	}

	updatedGoal, err := getGoalByID(goalID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch updated goal")
		return
	}

	respondJSON(w, http.StatusOK, updatedGoal)
}

// getGoalByID fetches a goal by ID
func getGoalByID(goalID int) (*models.ClientGoal, error) {
	var goal models.ClientGoal
	var description, targetDate sql.NullString
	var targetAmount, currentAmount sql.NullFloat64
	var completedAt sql.NullTime

	err := db.DB.QueryRow(
		`SELECT id, advisor_id, client_id, title, description, category, status, priority,
		target_amount, current_amount, target_date, completed_at, created_at, updated_at
		FROM client_goals WHERE id = ?`,
		goalID,
	).Scan(
		&goal.ID, &goal.AdvisorID, &goal.ClientID, &goal.Title,
		&description, &goal.Category, &goal.Status, &goal.Priority,
		&targetAmount, &currentAmount, &targetDate, &completedAt,
		&goal.CreatedAt, &goal.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if description.Valid {
		goal.Description = &description.String
	}
	if targetDate.Valid {
		goal.TargetDate = &targetDate.String
	}
	if targetAmount.Valid {
		goal.TargetAmount = &targetAmount.Float64
	}
	if currentAmount.Valid {
		goal.CurrentAmount = &currentAmount.Float64
	}
	if completedAt.Valid {
		goal.CompletedAt = &completedAt.Time
	}

	return &goal, nil
}
