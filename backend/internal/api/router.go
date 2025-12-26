package api

import (
	"encoding/json"
	"net/http"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()

	// Public routes (no auth required)
	mux.HandleFunc("POST /api/auth/register", handleRegister)
	mux.HandleFunc("POST /api/auth/login", handleLogin)
	mux.HandleFunc("GET /api/health", handleHealth)

	// Asset types (public - needed for registration form)
	mux.HandleFunc("GET /api/asset-types", handleGetAssetTypes)

	// Plaid status (public - to check if configured)
	mux.HandleFunc("GET /api/plaid/status", handlePlaidStatus)

	// Protected routes - wrap with auth middleware
	protectedMux := http.NewServeMux()

	// User info
	protectedMux.HandleFunc("GET /api/auth/me", handleGetMe)

	// Assets CRUD
	protectedMux.HandleFunc("GET /api/assets", handleGetAssets)
	protectedMux.HandleFunc("POST /api/assets", handleCreateAsset)
	protectedMux.HandleFunc("PUT /api/assets/{id}", handleUpdateAsset)
	protectedMux.HandleFunc("DELETE /api/assets/{id}", handleDeleteAsset)

	// Debts CRUD
	protectedMux.HandleFunc("GET /api/debts", handleGetDebts)
	protectedMux.HandleFunc("POST /api/debts", handleCreateDebt)
	protectedMux.HandleFunc("PUT /api/debts/{id}", handleUpdateDebt)
	protectedMux.HandleFunc("DELETE /api/debts/{id}", handleDeleteDebt)

	// Monte Carlo
	protectedMux.HandleFunc("POST /api/monte-carlo", handleMonteCarlo)

	// CSV Import
	protectedMux.HandleFunc("POST /api/import/csv", handleCSVImport)

	// Plaid endpoints
	protectedMux.HandleFunc("POST /api/plaid/link-token", handleCreateLinkToken)
	protectedMux.HandleFunc("POST /api/plaid/exchange-token", handleExchangeToken)
	protectedMux.HandleFunc("GET /api/plaid/items", handleGetPlaidItems)
	protectedMux.HandleFunc("DELETE /api/plaid/items/{id}", handleDeletePlaidItem)
	protectedMux.HandleFunc("GET /api/plaid/accounts", handleGetPlaidAccounts)
	protectedMux.HandleFunc("POST /api/plaid/sync", handleSyncAccounts)

	// Transactions endpoints
	protectedMux.HandleFunc("GET /api/transactions", handleGetTransactions)
	protectedMux.HandleFunc("GET /api/transactions/summary", handleGetTransactionSummary)
	protectedMux.HandleFunc("GET /api/transactions/categories", handleGetCategories)
	protectedMux.HandleFunc("POST /api/transactions/sync", handleSyncTransactions)

	// Apply auth middleware to protected routes
	mux.Handle("/api/auth/me", AuthMiddleware(protectedMux))
	mux.Handle("/api/assets", AuthMiddleware(protectedMux))
	mux.Handle("/api/assets/", AuthMiddleware(protectedMux))
	mux.Handle("/api/debts", AuthMiddleware(protectedMux))
	mux.Handle("/api/debts/", AuthMiddleware(protectedMux))
	mux.Handle("/api/monte-carlo", AuthMiddleware(protectedMux))
	mux.Handle("/api/import/", AuthMiddleware(protectedMux))
	mux.Handle("/api/plaid/", AuthMiddleware(protectedMux))
	mux.Handle("/api/transactions", AuthMiddleware(protectedMux))
	mux.Handle("/api/transactions/", AuthMiddleware(protectedMux))

	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
