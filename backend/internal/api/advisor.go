package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/finviz/backend/internal/auth"
	"github.com/finviz/backend/internal/db"
	"github.com/finviz/backend/internal/models"
)

// InviteClientRequest is the request body for inviting a client
type InviteClientRequest struct {
	Email       string `json:"email"`
	AccessLevel string `json:"accessLevel,omitempty"` // defaults to "full"
}

// CreateClientRequest is the request body for creating a client directly
type CreateClientRequest struct {
	Email       string `json:"email"`
	Name        string `json:"name"`
	Password    string `json:"password,omitempty"` // Optional - generate if not provided
	AccessLevel string `json:"accessLevel,omitempty"`
}

// UpdateClientRequest is the request body for updating client relationship
type UpdateClientRequest struct {
	AccessLevel string `json:"accessLevel,omitempty"`
	Status      string `json:"status,omitempty"`
}

// ClientSummary is the response for client list with summary info
type ClientSummary struct {
	models.User
	RelationshipID int       `json:"relationshipId"`
	AccessLevel    string    `json:"accessLevel"`
	Status         string    `json:"status"`
	AcceptedAt     *time.Time `json:"acceptedAt,omitempty"`
	TotalAssets    float64   `json:"totalAssets"`
	TotalDebts     float64   `json:"totalDebts"`
	NetWorth       float64   `json:"netWorth"`
	LastSimulation *time.Time `json:"lastSimulation,omitempty"`
}

// handleListClients returns list of advisor's clients with summary info
func handleListClients(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	rows, err := db.DB.Query(`
		SELECT
			u.id, u.email, u.name, u.role, u.created_at, u.updated_at,
			ac.id as relationship_id, ac.access_level, ac.status, ac.accepted_at,
			COALESCE(SUM(a.current_value), 0) as total_assets,
			COALESCE((SELECT SUM(current_balance) FROM debts WHERE user_id = u.id), 0) as total_debts,
			(SELECT MAX(created_at) FROM simulation_history WHERE user_id = u.id) as last_simulation
		FROM advisor_clients ac
		JOIN users u ON ac.client_id = u.id
		LEFT JOIN assets a ON a.user_id = u.id
		WHERE ac.advisor_id = ? AND ac.status != 'revoked'
		GROUP BY u.id, ac.id
		ORDER BY u.name
	`, user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch clients")
		return
	}
	defer rows.Close()

	clients := []ClientSummary{}
	for rows.Next() {
		var client ClientSummary
		var lastSim *time.Time
		err := rows.Scan(
			&client.ID, &client.Email, &client.Name, &client.Role,
			&client.CreatedAt, &client.UpdatedAt,
			&client.RelationshipID, &client.AccessLevel, &client.Status, &client.AcceptedAt,
			&client.TotalAssets, &client.TotalDebts, &lastSim,
		)
		if err != nil {
			continue
		}
		client.NetWorth = client.TotalAssets - client.TotalDebts
		client.LastSimulation = lastSim
		clients = append(clients, client)
	}

	respondJSON(w, http.StatusOK, clients)
}

// handleInviteClient sends an invitation to a client
func handleInviteClient(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req InviteClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" {
		respondError(w, http.StatusBadRequest, "Email is required")
		return
	}

	accessLevel := models.AccessLevelFull
	if req.AccessLevel != "" {
		if req.AccessLevel != models.AccessLevelView &&
		   req.AccessLevel != models.AccessLevelEdit &&
		   req.AccessLevel != models.AccessLevelFull {
			respondError(w, http.StatusBadRequest, "Invalid access level")
			return
		}
		accessLevel = req.AccessLevel
	}

	// Generate invitation token
	token := generateToken()
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days

	// Check if client already exists
	var existingUserID int
	err := db.DB.QueryRow("SELECT id FROM users WHERE email = ?", req.Email).Scan(&existingUserID)

	if err == nil {
		// User exists - create direct relationship
		// Check if relationship already exists
		var existingRelID int
		err = db.DB.QueryRow(
			"SELECT id FROM advisor_clients WHERE advisor_id = ? AND client_id = ?",
			user.ID, existingUserID,
		).Scan(&existingRelID)

		if err == nil {
			respondError(w, http.StatusConflict, "Client relationship already exists")
			return
		}

		// Create pending relationship
		_, err = db.DB.Exec(`
			INSERT INTO advisor_clients (advisor_id, client_id, status, access_level, invitation_token, invitation_expires_at)
			VALUES (?, ?, 'pending', ?, ?, ?)
		`, user.ID, existingUserID, accessLevel, token, expiresAt)

		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to create relationship")
			return
		}

		// TODO: Send email notification to existing user
		respondJSON(w, http.StatusCreated, map[string]interface{}{
			"message":     "Invitation sent to existing user",
			"clientId":    existingUserID,
			"status":      "pending",
		})
		return
	}

	// User doesn't exist - create invitation
	_, err = db.DB.Exec(`
		INSERT INTO client_invitations (advisor_id, client_email, invitation_token, expires_at)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE invitation_token = ?, expires_at = ?, status = 'pending'
	`, user.ID, req.Email, token, expiresAt, token, expiresAt)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create invitation")
		return
	}

	// TODO: Send invitation email
	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message":         "Invitation sent",
		"email":           req.Email,
		"invitationToken": token, // In production, send via email instead
		"expiresAt":       expiresAt,
	})
}

// handleCreateClient creates a new client account directly
func handleCreateClient(w http.ResponseWriter, r *http.Request) {
	advisor := getUserFromContext(r)
	if advisor == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req CreateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Name == "" {
		respondError(w, http.StatusBadRequest, "Email and name are required")
		return
	}

	// Check if user already exists
	var exists int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", req.Email).Scan(&exists)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if exists > 0 {
		respondError(w, http.StatusConflict, "Email already registered")
		return
	}

	// Generate password if not provided
	password := req.Password
	if password == "" {
		password = generateToken()[:16] // 16 char random password
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	accessLevel := models.AccessLevelFull
	if req.AccessLevel != "" {
		accessLevel = req.AccessLevel
	}

	// Create client user
	result, err := db.DB.Exec(
		`INSERT INTO users (email, password_hash, name, role, created_by_advisor_id)
		 VALUES (?, ?, ?, 'client', ?)`,
		req.Email, hashedPassword, req.Name, advisor.ID,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	clientID, _ := result.LastInsertId()

	// Create active relationship
	_, err = db.DB.Exec(`
		INSERT INTO advisor_clients (advisor_id, client_id, status, access_level, accepted_at)
		VALUES (?, ?, 'active', ?, NOW())
	`, advisor.ID, clientID, accessLevel)

	if err != nil {
		// Rollback user creation
		db.DB.Exec("DELETE FROM users WHERE id = ?", clientID)
		respondError(w, http.StatusInternalServerError, "Failed to create relationship")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message":          "Client created successfully",
		"clientId":         clientID,
		"email":            req.Email,
		"temporaryPassword": password, // In production, send via email
	})
}

// handleUpdateClient updates the advisor-client relationship
func handleUpdateClient(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	clientIDStr := r.PathValue("id")
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid client ID")
		return
	}

	var req UpdateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Verify relationship exists
	var exists int
	err = db.DB.QueryRow(
		"SELECT COUNT(*) FROM advisor_clients WHERE advisor_id = ? AND client_id = ?",
		user.ID, clientID,
	).Scan(&exists)
	if err != nil || exists == 0 {
		respondError(w, http.StatusNotFound, "Client relationship not found")
		return
	}

	// Build update query
	updates := []string{}
	args := []interface{}{}

	if req.AccessLevel != "" {
		if req.AccessLevel != models.AccessLevelView &&
		   req.AccessLevel != models.AccessLevelEdit &&
		   req.AccessLevel != models.AccessLevelFull {
			respondError(w, http.StatusBadRequest, "Invalid access level")
			return
		}
		updates = append(updates, "access_level = ?")
		args = append(args, req.AccessLevel)
	}

	if req.Status != "" {
		if req.Status != models.RelationshipStatusActive &&
		   req.Status != models.RelationshipStatusRevoked {
			respondError(w, http.StatusBadRequest, "Invalid status")
			return
		}
		updates = append(updates, "status = ?")
		args = append(args, req.Status)
	}

	if len(updates) == 0 {
		respondError(w, http.StatusBadRequest, "No updates provided")
		return
	}

	query := "UPDATE advisor_clients SET "
	for i, u := range updates {
		if i > 0 {
			query += ", "
		}
		query += u
	}
	query += " WHERE advisor_id = ? AND client_id = ?"
	args = append(args, user.ID, clientID)

	_, err = db.DB.Exec(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update relationship")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Client updated"})
}

// handleRemoveClient removes (revokes) the advisor-client relationship
func handleRemoveClient(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	clientIDStr := r.PathValue("id")
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid client ID")
		return
	}

	// Revoke instead of delete to maintain history
	result, err := db.DB.Exec(
		"UPDATE advisor_clients SET status = 'revoked' WHERE advisor_id = ? AND client_id = ?",
		user.ID, clientID,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to remove client")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		respondError(w, http.StatusNotFound, "Client relationship not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Client removed"})
}

// handleAddExistingClient adds an existing user as a client
func handleAddExistingClient(w http.ResponseWriter, r *http.Request) {
	advisor := getUserFromContext(r)
	if advisor == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req struct {
		ClientID    int    `json:"clientId"`
		AccessLevel string `json:"accessLevel,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ClientID == 0 {
		respondError(w, http.StatusBadRequest, "Client ID is required")
		return
	}

	// Verify client exists
	var clientRole string
	err := db.DB.QueryRow("SELECT role FROM users WHERE id = ?", req.ClientID).Scan(&clientRole)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	// Check if relationship already exists
	var existingID int
	err = db.DB.QueryRow(
		"SELECT id FROM advisor_clients WHERE advisor_id = ? AND client_id = ?",
		advisor.ID, req.ClientID,
	).Scan(&existingID)
	if err == nil {
		respondError(w, http.StatusConflict, "Relationship already exists")
		return
	}

	accessLevel := models.AccessLevelFull
	if req.AccessLevel != "" {
		accessLevel = req.AccessLevel
	}

	// Create active relationship
	_, err = db.DB.Exec(`
		INSERT INTO advisor_clients (advisor_id, client_id, status, access_level, accepted_at)
		VALUES (?, ?, 'active', ?, NOW())
	`, advisor.ID, req.ClientID, accessLevel)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to add client")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"message": "Client added successfully"})
}

// generateToken creates a secure random token
func generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// ==================== Admin Functions (Advisor Only) ====================

// CreateAdvisorRequest is the request body for creating a new advisor
type CreateAdvisorRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password,omitempty"` // Optional - generate if not provided
}

// UpdateAdvisorRequest is the request body for updating an advisor
type UpdateAdvisorRequest struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"` // Only set if changing password
}

// AdvisorSummary is the response for advisor list
type AdvisorSummary struct {
	ID          int       `json:"id"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"createdAt"`
	ClientCount int       `json:"clientCount"`
}

// handleListAdvisors returns list of all advisors (admin function)
func handleListAdvisors(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	rows, err := db.DB.Query(`
		SELECT
			u.id, u.email, u.name, u.role, u.created_at,
			(SELECT COUNT(*) FROM advisor_clients ac WHERE ac.advisor_id = u.id AND ac.status = 'active') as client_count
		FROM users u
		WHERE u.role = 'advisor'
		ORDER BY u.name
	`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch advisors")
		return
	}
	defer rows.Close()

	advisors := []AdvisorSummary{}
	for rows.Next() {
		var advisor AdvisorSummary
		err := rows.Scan(
			&advisor.ID, &advisor.Email, &advisor.Name, &advisor.Role,
			&advisor.CreatedAt, &advisor.ClientCount,
		)
		if err != nil {
			continue
		}
		advisors = append(advisors, advisor)
	}

	respondJSON(w, http.StatusOK, advisors)
}

// handleCreateAdvisor creates a new advisor account (admin function)
func handleCreateAdvisor(w http.ResponseWriter, r *http.Request) {
	currentUser := getUserFromContext(r)
	if currentUser == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req CreateAdvisorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Name == "" {
		respondError(w, http.StatusBadRequest, "Email and name are required")
		return
	}

	// Check if user already exists
	var exists int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", req.Email).Scan(&exists)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if exists > 0 {
		respondError(w, http.StatusConflict, "Email already registered")
		return
	}

	// Generate password if not provided
	password := req.Password
	generatedPassword := ""
	if password == "" {
		generatedPassword = generateToken()[:16] // 16 char random password
		password = generatedPassword
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Create advisor user
	result, err := db.DB.Exec(
		`INSERT INTO users (email, password_hash, name, role)
		 VALUES (?, ?, ?, 'advisor')`,
		req.Email, hashedPassword, req.Name,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create advisor")
		return
	}

	advisorID, _ := result.LastInsertId()

	response := map[string]interface{}{
		"message":   "Advisor created successfully",
		"advisorId": advisorID,
		"email":     req.Email,
		"name":      req.Name,
	}

	// Only include temporary password if we generated one
	if generatedPassword != "" {
		response["temporaryPassword"] = generatedPassword
	}

	respondJSON(w, http.StatusCreated, response)
}

// handleGetAdvisor returns details for a specific advisor
func handleGetAdvisor(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	advisorIDStr := r.PathValue("id")
	advisorID, err := strconv.Atoi(advisorIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid advisor ID")
		return
	}

	var advisor AdvisorSummary
	err = db.DB.QueryRow(`
		SELECT
			u.id, u.email, u.name, u.role, u.created_at,
			(SELECT COUNT(*) FROM advisor_clients ac WHERE ac.advisor_id = u.id AND ac.status = 'active') as client_count
		FROM users u
		WHERE u.id = ? AND u.role = 'advisor'
	`, advisorID).Scan(
		&advisor.ID, &advisor.Email, &advisor.Name, &advisor.Role,
		&advisor.CreatedAt, &advisor.ClientCount,
	)

	if err != nil {
		respondError(w, http.StatusNotFound, "Advisor not found")
		return
	}

	respondJSON(w, http.StatusOK, advisor)
}

// handleUpdateAdvisor updates an advisor account (admin function)
func handleUpdateAdvisor(w http.ResponseWriter, r *http.Request) {
	currentUser := getUserFromContext(r)
	if currentUser == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	advisorIDStr := r.PathValue("id")
	advisorID, err := strconv.Atoi(advisorIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid advisor ID")
		return
	}

	// Verify target is an advisor
	var targetRole string
	err = db.DB.QueryRow("SELECT role FROM users WHERE id = ?", advisorID).Scan(&targetRole)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}
	if targetRole != models.RoleAdvisor {
		respondError(w, http.StatusBadRequest, "User is not an advisor")
		return
	}

	var req UpdateAdvisorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Build update query dynamically
	updates := []string{}
	args := []interface{}{}

	if req.Name != "" {
		updates = append(updates, "name = ?")
		args = append(args, req.Name)
	}

	if req.Email != "" {
		// Check if email is already taken by another user
		var existingID int
		err := db.DB.QueryRow("SELECT id FROM users WHERE email = ? AND id != ?", req.Email, advisorID).Scan(&existingID)
		if err == nil {
			respondError(w, http.StatusConflict, "Email already in use")
			return
		}
		updates = append(updates, "email = ?")
		args = append(args, req.Email)
	}

	if req.Password != "" {
		hashedPassword, err := auth.HashPassword(req.Password)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to hash password")
			return
		}
		updates = append(updates, "password_hash = ?")
		args = append(args, hashedPassword)
	}

	if len(updates) == 0 {
		respondError(w, http.StatusBadRequest, "No updates provided")
		return
	}

	query := "UPDATE users SET "
	for i, u := range updates {
		if i > 0 {
			query += ", "
		}
		query += u
	}
	query += ", updated_at = NOW() WHERE id = ?"
	args = append(args, advisorID)

	_, err = db.DB.Exec(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update advisor")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Advisor updated successfully"})
}

// handleDeleteAdvisor deactivates an advisor account (admin function)
// Note: We don't actually delete to preserve data integrity
func handleDeleteAdvisor(w http.ResponseWriter, r *http.Request) {
	currentUser := getUserFromContext(r)
	if currentUser == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	advisorIDStr := r.PathValue("id")
	advisorID, err := strconv.Atoi(advisorIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid advisor ID")
		return
	}

	// Prevent self-deletion
	if advisorID == currentUser.ID {
		respondError(w, http.StatusBadRequest, "Cannot delete your own account")
		return
	}

	// Verify target is an advisor
	var targetRole string
	err = db.DB.QueryRow("SELECT role FROM users WHERE id = ?", advisorID).Scan(&targetRole)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}
	if targetRole != models.RoleAdvisor {
		respondError(w, http.StatusBadRequest, "User is not an advisor")
		return
	}

	// Check if advisor has active clients
	var clientCount int
	err = db.DB.QueryRow(
		"SELECT COUNT(*) FROM advisor_clients WHERE advisor_id = ? AND status = 'active'",
		advisorID,
	).Scan(&clientCount)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if clientCount > 0 {
		respondError(w, http.StatusConflict, "Cannot delete advisor with active clients. Reassign or remove clients first.")
		return
	}

	// Revoke all advisor-client relationships
	_, err = db.DB.Exec(
		"UPDATE advisor_clients SET status = 'revoked' WHERE advisor_id = ?",
		advisorID,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to revoke client relationships")
		return
	}

	// Delete the advisor (or convert to client if you want to preserve account)
	_, err = db.DB.Exec("DELETE FROM users WHERE id = ?", advisorID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete advisor")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Advisor deleted successfully"})
}

// handleListAllUsers returns list of all users (admin function)
func handleListAllUsers(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	rows, err := db.DB.Query(`
		SELECT
			u.id, u.email, u.name, u.role, u.created_at,
			(SELECT GROUP_CONCAT(CONCAT(adv.id, ':', adv.name) SEPARATOR '|')
			 FROM advisor_clients ac
			 JOIN users adv ON ac.advisor_id = adv.id
			 WHERE ac.client_id = u.id AND ac.status = 'active') as advisors
		FROM users u
		ORDER BY u.role, u.name
	`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}
	defer rows.Close()

	type AdvisorRef struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	type UserInfo struct {
		ID        int          `json:"id"`
		Email     string       `json:"email"`
		Name      string       `json:"name"`
		Role      string       `json:"role"`
		CreatedAt time.Time    `json:"createdAt"`
		Advisors  []AdvisorRef `json:"advisors,omitempty"`
	}

	users := []UserInfo{}
	for rows.Next() {
		var u UserInfo
		var advisorsStr *string
		err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.CreatedAt, &advisorsStr)
		if err != nil {
			continue
		}

		// Parse advisors string
		if advisorsStr != nil && *advisorsStr != "" {
			parts := splitString(*advisorsStr, "|")
			for _, part := range parts {
				idName := splitString(part, ":")
				if len(idName) == 2 {
					id, _ := strconv.Atoi(idName[0])
					u.Advisors = append(u.Advisors, AdvisorRef{ID: id, Name: idName[1]})
				}
			}
		}

		users = append(users, u)
	}

	respondJSON(w, http.StatusOK, users)
}

// splitString splits a string by separator
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	result := []string{}
	for _, part := range strings.Split(s, sep) {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// AssignClientRequest is the request body for assigning a client to an advisor
type AssignClientRequest struct {
	ClientID    int    `json:"clientId"`
	AdvisorID   int    `json:"advisorId"`
	AccessLevel string `json:"accessLevel,omitempty"`
}

// handleAssignClient assigns a client to a specific advisor (admin function)
func handleAssignClient(w http.ResponseWriter, r *http.Request) {
	currentUser := getUserFromContext(r)
	if currentUser == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req AssignClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ClientID == 0 || req.AdvisorID == 0 {
		respondError(w, http.StatusBadRequest, "Client ID and Advisor ID are required")
		return
	}

	// Verify client exists and is a client
	var clientRole string
	err := db.DB.QueryRow("SELECT role FROM users WHERE id = ?", req.ClientID).Scan(&clientRole)
	if err != nil {
		respondError(w, http.StatusNotFound, "Client not found")
		return
	}
	if clientRole != models.RoleClient {
		respondError(w, http.StatusBadRequest, "User is not a client")
		return
	}

	// Verify advisor exists and is an advisor
	var advisorRole string
	err = db.DB.QueryRow("SELECT role FROM users WHERE id = ?", req.AdvisorID).Scan(&advisorRole)
	if err != nil {
		respondError(w, http.StatusNotFound, "Advisor not found")
		return
	}
	if advisorRole != models.RoleAdvisor {
		respondError(w, http.StatusBadRequest, "Target user is not an advisor")
		return
	}

	// Check if relationship already exists
	var existingID int
	var existingStatus string
	err = db.DB.QueryRow(
		"SELECT id, status FROM advisor_clients WHERE advisor_id = ? AND client_id = ?",
		req.AdvisorID, req.ClientID,
	).Scan(&existingID, &existingStatus)

	accessLevel := models.AccessLevelFull
	if req.AccessLevel != "" {
		accessLevel = req.AccessLevel
	}

	if err == nil {
		// Relationship exists - reactivate if revoked
		if existingStatus == models.RelationshipStatusRevoked {
			_, err = db.DB.Exec(
				"UPDATE advisor_clients SET status = 'active', access_level = ?, accepted_at = NOW() WHERE id = ?",
				accessLevel, existingID,
			)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to reactivate relationship")
				return
			}
			respondJSON(w, http.StatusOK, map[string]string{"message": "Client relationship reactivated"})
			return
		}
		respondError(w, http.StatusConflict, "Relationship already exists")
		return
	}

	// Create new relationship
	_, err = db.DB.Exec(`
		INSERT INTO advisor_clients (advisor_id, client_id, status, access_level, accepted_at)
		VALUES (?, ?, 'active', ?, NOW())
	`, req.AdvisorID, req.ClientID, accessLevel)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to assign client")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"message": "Client assigned successfully"})
}

// handleClaimClient claims a client for the current advisor (convenience endpoint)
func handleClaimClient(w http.ResponseWriter, r *http.Request) {
	currentUser := getUserFromContext(r)
	if currentUser == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req struct {
		ClientID    int    `json:"clientId"`
		AccessLevel string `json:"accessLevel,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ClientID == 0 {
		respondError(w, http.StatusBadRequest, "Client ID is required")
		return
	}

	// Verify client exists and is a client
	var clientRole string
	err := db.DB.QueryRow("SELECT role FROM users WHERE id = ?", req.ClientID).Scan(&clientRole)
	if err != nil {
		respondError(w, http.StatusNotFound, "Client not found")
		return
	}
	if clientRole != models.RoleClient {
		respondError(w, http.StatusBadRequest, "User is not a client")
		return
	}

	// Check if relationship already exists
	var existingID int
	var existingStatus string
	err = db.DB.QueryRow(
		"SELECT id, status FROM advisor_clients WHERE advisor_id = ? AND client_id = ?",
		currentUser.ID, req.ClientID,
	).Scan(&existingID, &existingStatus)

	accessLevel := models.AccessLevelFull
	if req.AccessLevel != "" {
		accessLevel = req.AccessLevel
	}

	if err == nil {
		// Relationship exists - reactivate if revoked
		if existingStatus == models.RelationshipStatusRevoked {
			_, err = db.DB.Exec(
				"UPDATE advisor_clients SET status = 'active', access_level = ?, accepted_at = NOW() WHERE id = ?",
				accessLevel, existingID,
			)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to reactivate relationship")
				return
			}
			respondJSON(w, http.StatusOK, map[string]string{"message": "Client claimed successfully"})
			return
		}
		respondError(w, http.StatusConflict, "You already have a relationship with this client")
		return
	}

	// Create new relationship
	_, err = db.DB.Exec(`
		INSERT INTO advisor_clients (advisor_id, client_id, status, access_level, accepted_at)
		VALUES (?, ?, 'active', ?, NOW())
	`, currentUser.ID, req.ClientID, accessLevel)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to claim client")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"message": "Client claimed successfully"})
}
