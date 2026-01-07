package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/finviz/backend/internal/auth"
	"github.com/finviz/backend/internal/db"
	"github.com/finviz/backend/internal/models"
)

// InvitationDetailsResponse is the response for getting invitation details
type InvitationDetailsResponse struct {
	AdvisorName  string    `json:"advisorName"`
	AdvisorEmail string    `json:"advisorEmail"`
	ClientEmail  string    `json:"clientEmail"`
	ExpiresAt    time.Time `json:"expiresAt"`
	Status       string    `json:"status"`
	IsExpired    bool      `json:"isExpired"`
}

// AcceptInvitationRequest is the request body for accepting an invitation
type AcceptInvitationRequest struct {
	// For new users
	Name     string `json:"name,omitempty"`
	Password string `json:"password,omitempty"`
	// For existing users (login to accept)
	Email            string `json:"email,omitempty"`
	ExistingPassword string `json:"existingPassword,omitempty"`
}

// handleGetInvitation returns details about an invitation
func handleGetInvitation(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		respondError(w, http.StatusBadRequest, "Token is required")
		return
	}

	var invitation struct {
		ID           int
		AdvisorID    int
		ClientEmail  string
		Status       string
		ExpiresAt    time.Time
		AdvisorName  string
		AdvisorEmail string
	}

	err := db.DB.QueryRow(`
		SELECT ci.id, ci.advisor_id, ci.client_email, ci.status, ci.expires_at,
		       u.name, u.email
		FROM client_invitations ci
		JOIN users u ON ci.advisor_id = u.id
		WHERE ci.invitation_token = ?
	`, token).Scan(
		&invitation.ID, &invitation.AdvisorID, &invitation.ClientEmail,
		&invitation.Status, &invitation.ExpiresAt,
		&invitation.AdvisorName, &invitation.AdvisorEmail,
	)

	if err != nil {
		respondError(w, http.StatusNotFound, "Invitation not found")
		return
	}

	isExpired := time.Now().After(invitation.ExpiresAt)
	if invitation.Status != models.InvitationStatusPending {
		respondJSON(w, http.StatusOK, InvitationDetailsResponse{
			AdvisorName:  invitation.AdvisorName,
			AdvisorEmail: invitation.AdvisorEmail,
			ClientEmail:  invitation.ClientEmail,
			ExpiresAt:    invitation.ExpiresAt,
			Status:       invitation.Status,
			IsExpired:    isExpired,
		})
		return
	}

	if isExpired {
		// Mark as expired
		db.DB.Exec("UPDATE client_invitations SET status = 'expired' WHERE id = ?", invitation.ID)
	}

	respondJSON(w, http.StatusOK, InvitationDetailsResponse{
		AdvisorName:  invitation.AdvisorName,
		AdvisorEmail: invitation.AdvisorEmail,
		ClientEmail:  invitation.ClientEmail,
		ExpiresAt:    invitation.ExpiresAt,
		Status:       invitation.Status,
		IsExpired:    isExpired,
	})
}

// handleAcceptInvitation accepts an invitation (new or existing user)
func handleAcceptInvitation(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		respondError(w, http.StatusBadRequest, "Token is required")
		return
	}

	var req AcceptInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get invitation details
	var invitation struct {
		ID          int
		AdvisorID   int
		ClientEmail string
		Status      string
		ExpiresAt   time.Time
	}

	err := db.DB.QueryRow(`
		SELECT id, advisor_id, client_email, status, expires_at
		FROM client_invitations
		WHERE invitation_token = ?
	`, token).Scan(
		&invitation.ID, &invitation.AdvisorID, &invitation.ClientEmail,
		&invitation.Status, &invitation.ExpiresAt,
	)

	if err != nil {
		respondError(w, http.StatusNotFound, "Invitation not found")
		return
	}

	// Check status
	if invitation.Status != models.InvitationStatusPending {
		respondError(w, http.StatusBadRequest, "Invitation is no longer valid")
		return
	}

	// Check expiration
	if time.Now().After(invitation.ExpiresAt) {
		db.DB.Exec("UPDATE client_invitations SET status = 'expired' WHERE id = ?", invitation.ID)
		respondError(w, http.StatusBadRequest, "Invitation has expired")
		return
	}

	var clientID int
	var clientUser models.User

	// Check if user already exists
	err = db.DB.QueryRow(
		"SELECT id, email, password_hash, name, role FROM users WHERE email = ?",
		invitation.ClientEmail,
	).Scan(&clientUser.ID, &clientUser.Email, &clientUser.Password, &clientUser.Name, &clientUser.Role)

	if err == nil {
		// Existing user - verify password if provided
		if req.ExistingPassword != "" {
			if !auth.CheckPassword(req.ExistingPassword, clientUser.Password) {
				respondError(w, http.StatusUnauthorized, "Invalid password")
				return
			}
		}
		clientID = clientUser.ID
	} else {
		// New user - create account
		if req.Name == "" || req.Password == "" {
			respondError(w, http.StatusBadRequest, "Name and password are required for new users")
			return
		}

		if len(req.Password) < 8 {
			respondError(w, http.StatusBadRequest, "Password must be at least 8 characters")
			return
		}

		hashedPassword, err := auth.HashPassword(req.Password)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to hash password")
			return
		}

		result, err := db.DB.Exec(
			`INSERT INTO users (email, password_hash, name, role, created_by_advisor_id)
			 VALUES (?, ?, ?, 'client', ?)`,
			invitation.ClientEmail, hashedPassword, req.Name, invitation.AdvisorID,
		)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to create user")
			return
		}

		id, _ := result.LastInsertId()
		clientID = int(id)
		clientUser = models.User{
			ID:    clientID,
			Email: invitation.ClientEmail,
			Name:  req.Name,
			Role:  models.RoleClient,
		}
	}

	// Create or update advisor-client relationship
	_, err = db.DB.Exec(`
		INSERT INTO advisor_clients (advisor_id, client_id, status, access_level, accepted_at)
		VALUES (?, ?, 'active', 'full', NOW())
		ON DUPLICATE KEY UPDATE status = 'active', accepted_at = NOW()
	`, invitation.AdvisorID, clientID)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create relationship")
		return
	}

	// Mark invitation as accepted
	db.DB.Exec(
		"UPDATE client_invitations SET status = 'accepted', accepted_at = NOW() WHERE id = ?",
		invitation.ID,
	)

	// Generate auth token
	authToken, err := auth.GenerateToken(clientID, clientUser.Email)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondJSON(w, http.StatusOK, models.AuthResponse{
		Token: authToken,
		User:  clientUser,
	})
}

// handleAcceptPendingRelationship accepts a pending advisor relationship (for existing users)
func handleAcceptPendingRelationship(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	token := r.PathValue("token")
	if token == "" {
		respondError(w, http.StatusBadRequest, "Token is required")
		return
	}

	// Find the pending relationship
	var relationship struct {
		ID        int
		AdvisorID int
		ClientID  int
		ExpiresAt time.Time
	}

	err := db.DB.QueryRow(`
		SELECT id, advisor_id, client_id, invitation_expires_at
		FROM advisor_clients
		WHERE invitation_token = ? AND client_id = ? AND status = 'pending'
	`, token, user.ID).Scan(
		&relationship.ID, &relationship.AdvisorID, &relationship.ClientID,
		&relationship.ExpiresAt,
	)

	if err != nil {
		respondError(w, http.StatusNotFound, "Pending relationship not found")
		return
	}

	// Check expiration
	if time.Now().After(relationship.ExpiresAt) {
		respondError(w, http.StatusBadRequest, "Invitation has expired")
		return
	}

	// Accept the relationship
	_, err = db.DB.Exec(
		"UPDATE advisor_clients SET status = 'active', accepted_at = NOW() WHERE id = ?",
		relationship.ID,
	)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to accept relationship")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Relationship accepted"})
}

// handleRejectPendingRelationship rejects a pending advisor relationship
func handleRejectPendingRelationship(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	token := r.PathValue("token")
	if token == "" {
		respondError(w, http.StatusBadRequest, "Token is required")
		return
	}

	result, err := db.DB.Exec(`
		UPDATE advisor_clients
		SET status = 'revoked'
		WHERE invitation_token = ? AND client_id = ? AND status = 'pending'
	`, token, user.ID)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to reject relationship")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		respondError(w, http.StatusNotFound, "Pending relationship not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Relationship rejected"})
}

// handleListPendingInvitations lists pending invitations for a client
func handleListPendingInvitations(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	rows, err := db.DB.Query(`
		SELECT ac.id, ac.invitation_token, ac.invitation_expires_at, ac.created_at,
		       u.name as advisor_name, u.email as advisor_email
		FROM advisor_clients ac
		JOIN users u ON ac.advisor_id = u.id
		WHERE ac.client_id = ? AND ac.status = 'pending'
		ORDER BY ac.created_at DESC
	`, user.ID)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch invitations")
		return
	}
	defer rows.Close()

	type PendingInvitation struct {
		ID           int       `json:"id"`
		Token        string    `json:"token"`
		ExpiresAt    time.Time `json:"expiresAt"`
		CreatedAt    time.Time `json:"createdAt"`
		AdvisorName  string    `json:"advisorName"`
		AdvisorEmail string    `json:"advisorEmail"`
	}

	invitations := []PendingInvitation{}
	for rows.Next() {
		var inv PendingInvitation
		err := rows.Scan(
			&inv.ID, &inv.Token, &inv.ExpiresAt, &inv.CreatedAt,
			&inv.AdvisorName, &inv.AdvisorEmail,
		)
		if err != nil {
			continue
		}
		invitations = append(invitations, inv)
	}

	respondJSON(w, http.StatusOK, invitations)
}
