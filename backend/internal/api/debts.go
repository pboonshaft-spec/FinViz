package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/finviz/backend/internal/db"
	"github.com/finviz/backend/internal/models"
)

func handleGetDebts(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	rows, err := db.DB.Query(`
		SELECT id, user_id, name, current_balance, interest_rate, minimum_payment, plaid_account_id, created_at, updated_at
		FROM debts
		WHERE user_id = ?
		ORDER BY name
	`, user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var debts []models.Debt
	for rows.Next() {
		var d models.Debt
		var interestRate, minimumPayment sql.NullFloat64
		var plaidAccountID sql.NullString
		if err := rows.Scan(&d.ID, &d.UserID, &d.Name, &d.CurrentBalance, &interestRate, &minimumPayment, &plaidAccountID, &d.CreatedAt, &d.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if interestRate.Valid {
			d.InterestRate = &interestRate.Float64
		}
		if minimumPayment.Valid {
			d.MinimumPayment = &minimumPayment.Float64
		}
		if plaidAccountID.Valid {
			d.PlaidAccountID = &plaidAccountID.String
		}
		debts = append(debts, d)
	}

	if debts == nil {
		debts = []models.Debt{}
	}

	respondJSON(w, http.StatusOK, debts)
}

func handleCreateDebt(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req models.CreateDebtRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Name is required")
		return
	}

	result, err := db.DB.Exec(
		`INSERT INTO debts (user_id, name, current_balance, interest_rate, minimum_payment) VALUES (?, ?, ?, ?, ?)`,
		user.ID, req.Name, req.CurrentBalance, req.InterestRate, req.MinimumPayment,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	respondJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func handleUpdateDebt(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var req models.UpdateDebtRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Build dynamic update query
	query := "UPDATE debts SET updated_at = NOW()"
	args := []interface{}{}

	if req.Name != nil {
		query += ", name = ?"
		args = append(args, *req.Name)
	}
	if req.CurrentBalance != nil {
		query += ", current_balance = ?"
		args = append(args, *req.CurrentBalance)
	}
	if req.InterestRate != nil {
		query += ", interest_rate = ?"
		args = append(args, *req.InterestRate)
	}
	if req.MinimumPayment != nil {
		query += ", minimum_payment = ?"
		args = append(args, *req.MinimumPayment)
	}

	query += " WHERE id = ? AND user_id = ?"
	args = append(args, id, user.ID)

	result, err := db.DB.Exec(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		respondError(w, http.StatusNotFound, "Debt not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func handleDeleteDebt(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	result, err := db.DB.Exec("DELETE FROM debts WHERE id = ? AND user_id = ?", id, user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		respondError(w, http.StatusNotFound, "Debt not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
