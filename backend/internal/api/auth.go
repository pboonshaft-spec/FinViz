package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/finviz/backend/internal/auth"
	"github.com/finviz/backend/internal/db"
	"github.com/finviz/backend/internal/models"
)

type contextKey string

const (
	userContextKey       contextKey = "user"
	clientContextKey     contextKey = "client"       // The client being acted upon (for advisors)
	actingAsAdvisorKey   contextKey = "actingAsAdvisor"
)

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" || req.Name == "" {
		respondError(w, http.StatusBadRequest, "Email, password, and name are required")
		return
	}

	if len(req.Password) < 8 {
		respondError(w, http.StatusBadRequest, "Password must be at least 8 characters")
		return
	}

	// Validate and default role
	role := models.RoleClient
	if req.Role != "" {
		if req.Role != models.RoleClient && req.Role != models.RoleAdvisor {
			respondError(w, http.StatusBadRequest, "Invalid role. Must be 'client' or 'advisor'")
			return
		}
		role = req.Role
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

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Create user
	result, err := db.DB.Exec(
		"INSERT INTO users (email, password_hash, name, role) VALUES (?, ?, ?, ?)",
		req.Email, hashedPassword, req.Name, role,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	userID, _ := result.LastInsertId()

	// Generate token
	token, err := auth.GenerateToken(int(userID), req.Email)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondJSON(w, http.StatusCreated, models.AuthResponse{
		Token: token,
		User: models.User{
			ID:    int(userID),
			Email: req.Email,
			Name:  req.Name,
			Role:  role,
		},
	})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Get user
	var user models.User
	var passwordHash string
	err := db.DB.QueryRow(
		"SELECT id, email, password_hash, name, role FROM users WHERE email = ?",
		req.Email,
	).Scan(&user.ID, &user.Email, &passwordHash, &user.Name, &user.Role)

	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Check password
	if !auth.CheckPassword(req.Password, passwordHash) {
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate token
	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondJSON(w, http.StatusOK, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

func handleGetMe(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}
	respondJSON(w, http.StatusOK, user)
}

// AuthMiddleware validates the JWT token and adds user to context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		// Extract token (Bearer <token>)
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		tokenString := parts[1]

		// Validate token
		token, err := auth.ValidateToken(tokenString)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// Get user from database
		var user models.User
		err = db.DB.QueryRow(
			"SELECT id, email, name, role, created_at, updated_at FROM users WHERE id = ?",
			token.UserID,
		).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)

		if err != nil {
			respondError(w, http.StatusUnauthorized, "User not found")
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), userContextKey, &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getUserFromContext retrieves the user from the request context
func getUserFromContext(r *http.Request) *models.User {
	user, ok := r.Context().Value(userContextKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}

// getClientContext retrieves the client being acted upon from the request context
func getClientContext(r *http.Request) *models.User {
	client, ok := r.Context().Value(clientContextKey).(*models.User)
	if !ok {
		return nil
	}
	return client
}

// getEffectiveUserID returns the user ID to use for data operations
// If advisor is acting on behalf of client, returns client ID; otherwise returns authenticated user's ID
func getEffectiveUserID(r *http.Request) int {
	if client := getClientContext(r); client != nil {
		return client.ID
	}
	user := getUserFromContext(r)
	if user != nil {
		return user.ID
	}
	return 0
}

// isActingAsAdvisor returns true if the current request is an advisor acting on behalf of a client
func isActingAsAdvisor(r *http.Request) bool {
	acting, ok := r.Context().Value(actingAsAdvisorKey).(bool)
	return ok && acting
}

// AdvisorMiddleware ensures the user is an advisor
func AdvisorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromContext(r)
		if user == nil {
			respondError(w, http.StatusUnauthorized, "Not authenticated")
			return
		}
		if !user.IsAdvisor() {
			respondError(w, http.StatusForbidden, "Advisor access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ClientAccessMiddleware validates advisor has access to specified client
// Extracts clientId from URL path: /api/advisor/clients/{clientId}/...
func ClientAccessMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromContext(r)
		if user == nil {
			respondError(w, http.StatusUnauthorized, "Not authenticated")
			return
		}
		if !user.IsAdvisor() {
			respondError(w, http.StatusForbidden, "Advisor access required")
			return
		}

		// Extract clientId from URL path: /api/advisor/clients/{clientId}/...
		// Path format: /api/advisor/clients/123/assets
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		var clientIDStr string
		// pathParts: ["api", "advisor", "clients", "123", "assets", ...]
		if len(pathParts) >= 4 && pathParts[0] == "api" && pathParts[1] == "advisor" && pathParts[2] == "clients" {
			clientIDStr = pathParts[3]
		}
		if clientIDStr == "" {
			respondError(w, http.StatusBadRequest, "Client ID required")
			return
		}

		clientID, err := strconv.Atoi(clientIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid client ID")
			return
		}

		// Verify relationship exists and is active
		var relationshipID int
		var accessLevel string
		err = db.DB.QueryRow(`
			SELECT id, access_level FROM advisor_clients
			WHERE advisor_id = ? AND client_id = ? AND status = 'active'
		`, user.ID, clientID).Scan(&relationshipID, &accessLevel)

		if err != nil {
			respondError(w, http.StatusForbidden, "No access to this client")
			return
		}

		// Load client user
		var client models.User
		err = db.DB.QueryRow(
			"SELECT id, email, name, role, created_at, updated_at FROM users WHERE id = ?",
			clientID,
		).Scan(&client.ID, &client.Email, &client.Name, &client.Role, &client.CreatedAt, &client.UpdatedAt)

		if err != nil {
			respondError(w, http.StatusNotFound, "Client not found")
			return
		}

		// Add client context
		ctx := context.WithValue(r.Context(), clientContextKey, &client)
		ctx = context.WithValue(ctx, actingAsAdvisorKey, true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getAccessLevel returns the access level for the current advisor-client relationship
func getAccessLevel(r *http.Request) string {
	user := getUserFromContext(r)
	client := getClientContext(r)
	if user == nil || client == nil {
		return ""
	}

	var accessLevel string
	err := db.DB.QueryRow(`
		SELECT access_level FROM advisor_clients
		WHERE advisor_id = ? AND client_id = ? AND status = 'active'
	`, user.ID, client.ID).Scan(&accessLevel)

	if err != nil {
		return ""
	}
	return accessLevel
}

// canEdit returns true if the current user can edit the target user's data
func canEdit(r *http.Request) bool {
	if !isActingAsAdvisor(r) {
		return true // User editing their own data
	}
	level := getAccessLevel(r)
	return level == models.AccessLevelEdit || level == models.AccessLevelFull
}

// canRunSimulations returns true if the current user can run simulations for the target
func canRunSimulations(r *http.Request) bool {
	if !isActingAsAdvisor(r) {
		return true // User running their own simulations
	}
	return getAccessLevel(r) == models.AccessLevelFull
}
