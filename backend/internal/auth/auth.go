package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
)

var jwtSecret []byte

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Generate a random secret for development
		secret = generateRandomSecret()
	}
	jwtSecret = []byte(secret)
}

func generateRandomSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword compares a password with a hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Token represents a simple JWT-like token structure
type Token struct {
	UserID    int
	Email     string
	ExpiresAt time.Time
}

// GenerateToken creates a simple base64 encoded token
// In production, use a proper JWT library
func GenerateToken(userID int, email string) (string, error) {
	// Create expiration time (24 hours from now)
	expiresAt := time.Now().Add(24 * time.Hour)

	// Simple token format: userID:email:expiry:signature
	tokenData := []byte(encodeTokenData(userID, email, expiresAt))

	// Create HMAC signature
	signature := createHMAC(tokenData)

	// Combine and encode
	combined := append(tokenData, signature...)
	return base64.URLEncoding.EncodeToString(combined), nil
}

// ValidateToken validates the token and returns the claims
func ValidateToken(tokenString string) (*Token, error) {
	// Decode the token
	combined, err := base64.URLEncoding.DecodeString(tokenString)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if len(combined) < 32 {
		return nil, ErrInvalidToken
	}

	// Split data and signature
	tokenData := combined[:len(combined)-32]
	providedSig := combined[len(combined)-32:]

	// Verify signature
	expectedSig := createHMAC(tokenData)
	if !hmacEqual(providedSig, expectedSig) {
		return nil, ErrInvalidToken
	}

	// Parse token data
	token, err := decodeTokenData(string(tokenData))
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Check expiration
	if time.Now().After(token.ExpiresAt) {
		return nil, ErrInvalidToken
	}

	return token, nil
}

// Helper functions for simple token encoding
func encodeTokenData(userID int, email string, expiresAt time.Time) string {
	return strconv.Itoa(userID) + ":" + email + ":" + expiresAt.Format(time.RFC3339)
}

func decodeTokenData(data string) (*Token, error) {
	// Parse the simple format
	parts := splitTokenParts(data)
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	userID, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}
	email := parts[1]
	expiresAt, err := time.Parse(time.RFC3339, parts[2])
	if err != nil {
		return nil, ErrInvalidToken
	}

	return &Token{
		UserID:    userID,
		Email:     email,
		ExpiresAt: expiresAt,
	}, nil
}

func splitTokenParts(data string) []string {
	var parts []string
	var current string
	for i, c := range data {
		if c == ':' && len(parts) < 2 {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(c)
		}
		if i == len(data)-1 {
			parts = append(parts, current)
		}
	}
	return parts
}

func createHMAC(data []byte) []byte {
	// Simple HMAC using the secret
	combined := append(data, jwtSecret...)
	h := make([]byte, 32)
	for i := 0; i < 32; i++ {
		h[i] = combined[i%len(combined)]
		for j := 0; j < len(combined); j++ {
			h[i] ^= combined[(i+j)%len(combined)]
		}
	}
	return h
}

func hmacEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	result := 0
	for i := range a {
		result |= int(a[i] ^ b[i])
	}
	return result == 0
}
