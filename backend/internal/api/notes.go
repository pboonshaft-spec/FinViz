package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/finviz/backend/internal/db"
	"github.com/finviz/backend/internal/models"
)

// handleListClientNotes returns all notes for a specific client (advisor only)
func handleListClientNotes(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil || !user.IsAdvisor() {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
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

	// Optional category filter
	category := r.URL.Query().Get("category")

	var notes []models.ClientNote
	var query string
	var args []interface{}

	if category != "" {
		query = `SELECT id, advisor_id, client_id, note, category, is_pinned, created_at, updated_at
			FROM client_notes
			WHERE advisor_id = ? AND client_id = ? AND category = ?
			ORDER BY is_pinned DESC, created_at DESC`
		args = []interface{}{user.ID, clientID, category}
	} else {
		query = `SELECT id, advisor_id, client_id, note, category, is_pinned, created_at, updated_at
			FROM client_notes
			WHERE advisor_id = ? AND client_id = ?
			ORDER BY is_pinned DESC, created_at DESC`
		args = []interface{}{user.ID, clientID}
	}

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch notes")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var note models.ClientNote
		err := rows.Scan(&note.ID, &note.AdvisorID, &note.ClientID, &note.Note, &note.Category, &note.IsPinned, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to parse notes")
			return
		}
		notes = append(notes, note)
	}

	if notes == nil {
		notes = []models.ClientNote{}
	}

	respondJSON(w, http.StatusOK, notes)
}

// handleCreateClientNote creates a new note for a client (advisor only)
func handleCreateClientNote(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil || !user.IsAdvisor() {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
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

	var req models.CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Note == "" {
		respondError(w, http.StatusBadRequest, "Note content is required")
		return
	}

	// Default category
	if req.Category == "" {
		req.Category = models.NoteCategoryGeneral
	}

	// Validate category
	validCategories := map[string]bool{
		models.NoteCategoryGeneral:    true,
		models.NoteCategoryMeeting:    true,
		models.NoteCategoryGoal:       true,
		models.NoteCategoryConcern:    true,
		models.NoteCategoryActionItem: true,
		models.NoteCategoryPersonal:   true,
	}
	if !validCategories[req.Category] {
		respondError(w, http.StatusBadRequest, "Invalid category")
		return
	}

	result, err := db.DB.Exec(
		`INSERT INTO client_notes (advisor_id, client_id, note, category, is_pinned) VALUES (?, ?, ?, ?, ?)`,
		user.ID, clientID, req.Note, req.Category, req.IsPinned,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create note")
		return
	}

	noteID, _ := result.LastInsertId()

	// Fetch the created note
	var note models.ClientNote
	err = db.DB.QueryRow(
		`SELECT id, advisor_id, client_id, note, category, is_pinned, created_at, updated_at FROM client_notes WHERE id = ?`,
		noteID,
	).Scan(&note.ID, &note.AdvisorID, &note.ClientID, &note.Note, &note.Category, &note.IsPinned, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch created note")
		return
	}

	respondJSON(w, http.StatusCreated, note)
}

// handleUpdateClientNote updates an existing note (advisor only)
func handleUpdateClientNote(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil || !user.IsAdvisor() {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	clientID, err := strconv.Atoi(r.PathValue("clientId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid client ID")
		return
	}

	noteID, err := strconv.Atoi(r.PathValue("noteId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid note ID")
		return
	}

	// Verify advisor owns this note
	var existingNote models.ClientNote
	err = db.DB.QueryRow(
		`SELECT id, advisor_id, client_id, note, category, is_pinned, created_at, updated_at
		FROM client_notes WHERE id = ? AND advisor_id = ? AND client_id = ?`,
		noteID, user.ID, clientID,
	).Scan(&existingNote.ID, &existingNote.AdvisorID, &existingNote.ClientID, &existingNote.Note, &existingNote.Category, &existingNote.IsPinned, &existingNote.CreatedAt, &existingNote.UpdatedAt)
	if err != nil {
		respondError(w, http.StatusNotFound, "Note not found")
		return
	}

	var req models.UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Apply updates
	if req.Note != "" {
		existingNote.Note = req.Note
	}
	if req.Category != "" {
		validCategories := map[string]bool{
			models.NoteCategoryGeneral:    true,
			models.NoteCategoryMeeting:    true,
			models.NoteCategoryGoal:       true,
			models.NoteCategoryConcern:    true,
			models.NoteCategoryActionItem: true,
			models.NoteCategoryPersonal:   true,
		}
		if !validCategories[req.Category] {
			respondError(w, http.StatusBadRequest, "Invalid category")
			return
		}
		existingNote.Category = req.Category
	}
	if req.IsPinned != nil {
		existingNote.IsPinned = *req.IsPinned
	}

	_, err = db.DB.Exec(
		`UPDATE client_notes SET note = ?, category = ?, is_pinned = ? WHERE id = ?`,
		existingNote.Note, existingNote.Category, existingNote.IsPinned, noteID,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update note")
		return
	}

	// Fetch updated note
	var updatedNote models.ClientNote
	err = db.DB.QueryRow(
		`SELECT id, advisor_id, client_id, note, category, is_pinned, created_at, updated_at FROM client_notes WHERE id = ?`,
		noteID,
	).Scan(&updatedNote.ID, &updatedNote.AdvisorID, &updatedNote.ClientID, &updatedNote.Note, &updatedNote.Category, &updatedNote.IsPinned, &updatedNote.CreatedAt, &updatedNote.UpdatedAt)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch updated note")
		return
	}

	respondJSON(w, http.StatusOK, updatedNote)
}

// handleDeleteClientNote deletes a note (advisor only)
func handleDeleteClientNote(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil || !user.IsAdvisor() {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	clientID, err := strconv.Atoi(r.PathValue("clientId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid client ID")
		return
	}

	noteID, err := strconv.Atoi(r.PathValue("noteId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid note ID")
		return
	}

	// Verify advisor owns this note
	result, err := db.DB.Exec(
		`DELETE FROM client_notes WHERE id = ? AND advisor_id = ? AND client_id = ?`,
		noteID, user.ID, clientID,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete note")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		respondError(w, http.StatusNotFound, "Note not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Note deleted successfully"})
}

// handleGetAllClientNotes returns all notes for the advisor across all clients
func handleGetAllClientNotes(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil || !user.IsAdvisor() {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Optional category filter
	category := r.URL.Query().Get("category")
	limit := 50 // Default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	var notes []models.ClientNoteWithClient
	var query string
	var args []interface{}

	if category != "" {
		query = `SELECT n.id, n.advisor_id, n.client_id, n.note, n.category, n.is_pinned, n.created_at, n.updated_at, u.name
			FROM client_notes n
			JOIN users u ON n.client_id = u.id
			WHERE n.advisor_id = ? AND n.category = ?
			ORDER BY n.is_pinned DESC, n.created_at DESC
			LIMIT ?`
		args = []interface{}{user.ID, category, limit}
	} else {
		query = `SELECT n.id, n.advisor_id, n.client_id, n.note, n.category, n.is_pinned, n.created_at, n.updated_at, u.name
			FROM client_notes n
			JOIN users u ON n.client_id = u.id
			WHERE n.advisor_id = ?
			ORDER BY n.is_pinned DESC, n.created_at DESC
			LIMIT ?`
		args = []interface{}{user.ID, limit}
	}

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch notes")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var note models.ClientNoteWithClient
		err := rows.Scan(&note.ID, &note.AdvisorID, &note.ClientID, &note.Note, &note.Category, &note.IsPinned, &note.CreatedAt, &note.UpdatedAt, &note.ClientName)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to parse notes")
			return
		}
		notes = append(notes, note)
	}

	if notes == nil {
		notes = []models.ClientNoteWithClient{}
	}

	respondJSON(w, http.StatusOK, notes)
}

// advisorHasClientAccess checks if the advisor has an active relationship with the client
func advisorHasClientAccess(advisorID, clientID int) bool {
	var count int
	err := db.DB.QueryRow(
		`SELECT COUNT(*) FROM advisor_clients WHERE advisor_id = ? AND client_id = ? AND status = 'active'`,
		advisorID, clientID,
	).Scan(&count)
	return err == nil && count > 0
}
