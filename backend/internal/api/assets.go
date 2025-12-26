package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/finviz/backend/internal/db"
	"github.com/finviz/backend/internal/models"
)

func handleGetAssetTypes(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`SELECT id, name, default_return, default_volatility, created_at FROM asset_types ORDER BY name`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var types []models.AssetType
	for rows.Next() {
		var t models.AssetType
		if err := rows.Scan(&t.ID, &t.Name, &t.DefaultReturn, &t.DefaultVolatility, &t.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		types = append(types, t)
	}

	if types == nil {
		types = []models.AssetType{}
	}

	respondJSON(w, http.StatusOK, types)
}

func handleGetAssets(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	rows, err := db.DB.Query(`
		SELECT a.id, a.user_id, a.name, a.type_id, a.current_value, a.custom_return, a.custom_volatility,
		       a.plaid_account_id, a.created_at, a.updated_at, t.id, t.name, t.default_return, t.default_volatility
		FROM assets a
		JOIN asset_types t ON a.type_id = t.id
		WHERE a.user_id = ?
		ORDER BY a.name
	`, user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var assets []models.Asset
	for rows.Next() {
		var a models.Asset
		var t models.AssetType
		var customReturn, customVolatility sql.NullFloat64
		var plaidAccountID sql.NullString
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.Name, &a.TypeID, &a.CurrentValue, &customReturn, &customVolatility,
			&plaidAccountID, &a.CreatedAt, &a.UpdatedAt, &t.ID, &t.Name, &t.DefaultReturn, &t.DefaultVolatility,
		); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if customReturn.Valid {
			a.CustomReturn = &customReturn.Float64
		}
		if customVolatility.Valid {
			a.CustomVolatility = &customVolatility.Float64
		}
		if plaidAccountID.Valid {
			a.PlaidAccountID = &plaidAccountID.String
		}
		a.AssetType = &t
		assets = append(assets, a)
	}

	if assets == nil {
		assets = []models.Asset{}
	}

	respondJSON(w, http.StatusOK, assets)
}

func handleCreateAsset(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req models.CreateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Name is required")
		return
	}

	result, err := db.DB.Exec(
		`INSERT INTO assets (user_id, name, type_id, current_value, custom_return, custom_volatility) VALUES (?, ?, ?, ?, ?, ?)`,
		user.ID, req.Name, req.TypeID, req.CurrentValue, req.CustomReturn, req.CustomVolatility,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	respondJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func handleUpdateAsset(w http.ResponseWriter, r *http.Request) {
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

	var req models.UpdateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Build dynamic update query
	query := "UPDATE assets SET updated_at = NOW()"
	args := []interface{}{}

	if req.Name != nil {
		query += ", name = ?"
		args = append(args, *req.Name)
	}
	if req.TypeID != nil {
		query += ", type_id = ?"
		args = append(args, *req.TypeID)
	}
	if req.CurrentValue != nil {
		query += ", current_value = ?"
		args = append(args, *req.CurrentValue)
	}
	if req.CustomReturn != nil {
		query += ", custom_return = ?"
		args = append(args, *req.CustomReturn)
	}
	if req.CustomVolatility != nil {
		query += ", custom_volatility = ?"
		args = append(args, *req.CustomVolatility)
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
		respondError(w, http.StatusNotFound, "Asset not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func handleDeleteAsset(w http.ResponseWriter, r *http.Request) {
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

	result, err := db.DB.Exec("DELETE FROM assets WHERE id = ? AND user_id = ?", id, user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		respondError(w, http.StatusNotFound, "Asset not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
