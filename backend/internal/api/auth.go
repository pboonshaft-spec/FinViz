package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/finviz/backend/internal/auth"
	"github.com/finviz/backend/internal/db"
	"github.com/finviz/backend/internal/models"
)

type contextKey string

const userContextKey contextKey = "user"

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
		"INSERT INTO users (email, password_hash, name) VALUES (?, ?, ?)",
		req.Email, hashedPassword, req.Name,
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
		"SELECT id, email, password_hash, name FROM users WHERE email = ?",
		req.Email,
	).Scan(&user.ID, &user.Email, &passwordHash, &user.Name)

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
			"SELECT id, email, name, created_at, updated_at FROM users WHERE id = ?",
			token.UserID,
		).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt)

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
