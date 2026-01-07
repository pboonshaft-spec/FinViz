package api

import (
	"encoding/json"
	"net/http"
	"regexp"
)

// clientIDPattern matches paths with a numeric client ID followed by more path segments
// e.g., /api/advisor/clients/123/assets
var clientIDPattern = regexp.MustCompile(`^/api/advisor/clients/\d+/.+`)

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

	// Chat status (public - to check if configured)
	mux.HandleFunc("GET /api/chat/status", handleChatStatus)

	// Public invitation endpoints
	mux.HandleFunc("GET /api/invitation/{token}", handleGetInvitation)
	mux.HandleFunc("POST /api/invitation/{token}/accept", handleAcceptInvitation)

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
	protectedMux.HandleFunc("POST /api/monte-carlo/scenarios", handleScenarioComparison)

	// Simulation History
	protectedMux.HandleFunc("GET /api/simulations", handleListSimulations)
	protectedMux.HandleFunc("GET /api/simulations/{id}", handleGetSimulation)
	protectedMux.HandleFunc("POST /api/simulations", handleSaveSimulation)
	protectedMux.HandleFunc("PUT /api/simulations/{id}", handleUpdateSimulation)
	protectedMux.HandleFunc("DELETE /api/simulations/{id}", handleDeleteSimulation)

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
	protectedMux.HandleFunc("GET /api/transactions/debug", handleGetTransactionDebug)
	protectedMux.HandleFunc("POST /api/transactions/sync", handleSyncTransactions)

	// Chat endpoint
	protectedMux.HandleFunc("POST /api/chat", handleChat)

	// Report generation
	protectedMux.HandleFunc("POST /api/reports/generate", handleGenerateReport)

	// Messaging endpoints
	protectedMux.HandleFunc("GET /api/messages/conversations", handleListConversations)
	protectedMux.HandleFunc("POST /api/messages/conversations", handleStartConversation)
	protectedMux.HandleFunc("GET /api/messages/conversations/{id}", handleGetConversation)
	protectedMux.HandleFunc("GET /api/messages/conversations/{id}/messages", handleGetMessages)
	protectedMux.HandleFunc("POST /api/messages/conversations/{id}/messages", handleSendMessage)
	protectedMux.HandleFunc("POST /api/messages/conversations/{id}/read", handleMarkAsRead)
	protectedMux.HandleFunc("GET /api/messages/unread", handleGetUnreadCounts)
	protectedMux.HandleFunc("POST /api/messages/keys", handleRegisterPublicKey)
	protectedMux.HandleFunc("GET /api/messages/keys/{userId}", handleGetPublicKey)

	// Client pending invitations
	protectedMux.HandleFunc("GET /api/invitations/pending", handleListPendingInvitations)
	protectedMux.HandleFunc("POST /api/invitations/{token}/accept", handleAcceptPendingRelationship)
	protectedMux.HandleFunc("POST /api/invitations/{token}/reject", handleRejectPendingRelationship)

	// Document vault endpoints
	protectedMux.HandleFunc("POST /api/documents/upload", HandleDocumentUpload)
	protectedMux.HandleFunc("GET /api/documents", HandleDocumentList)
	protectedMux.HandleFunc("GET /api/documents/{id}/download", HandleDocumentDownload)
	protectedMux.HandleFunc("DELETE /api/documents/{id}", HandleDocumentDelete)
	protectedMux.HandleFunc("POST /api/documents/{id}/share", HandleDocumentShare)

	// Client goals endpoints (for clients viewing their own goals)
	protectedMux.HandleFunc("GET /api/goals", handleGetMyGoals)
	protectedMux.HandleFunc("PUT /api/goals/{goalId}/progress", handleUpdateMyGoalProgress)

	// Advisor-only routes (handled in advisor mux)
	advisorMux := http.NewServeMux()
	advisorMux.HandleFunc("GET /api/advisor/clients", handleListClients)
	advisorMux.HandleFunc("POST /api/advisor/clients/invite", handleInviteClient)
	advisorMux.HandleFunc("POST /api/advisor/clients/create", handleCreateClient)
	advisorMux.HandleFunc("POST /api/advisor/clients/add", handleAddExistingClient)
	advisorMux.HandleFunc("PUT /api/advisor/clients/{id}", handleUpdateClient)
	advisorMux.HandleFunc("DELETE /api/advisor/clients/{id}", handleRemoveClient)

	// Client notes (advisor-only)
	advisorMux.HandleFunc("GET /api/advisor/notes", handleGetAllClientNotes)

	// Admin routes (advisor-only) for managing advisors and users
	advisorMux.HandleFunc("GET /api/advisor/admin/advisors", handleListAdvisors)
	advisorMux.HandleFunc("POST /api/advisor/admin/advisors", handleCreateAdvisor)
	advisorMux.HandleFunc("GET /api/advisor/admin/advisors/{id}", handleGetAdvisor)
	advisorMux.HandleFunc("PUT /api/advisor/admin/advisors/{id}", handleUpdateAdvisor)
	advisorMux.HandleFunc("DELETE /api/advisor/admin/advisors/{id}", handleDeleteAdvisor)
	advisorMux.HandleFunc("GET /api/advisor/admin/users", handleListAllUsers)
	advisorMux.HandleFunc("POST /api/advisor/admin/assign-client", handleAssignClient)
	advisorMux.HandleFunc("POST /api/advisor/admin/claim-client", handleClaimClient)

	// Advisor client context routes (for viewing/managing specific client's data)
	clientContextMux := http.NewServeMux()
	clientContextMux.HandleFunc("GET /api/advisor/clients/{clientId}/assets", handleGetAssets)
	clientContextMux.HandleFunc("POST /api/advisor/clients/{clientId}/assets", handleCreateAsset)
	clientContextMux.HandleFunc("PUT /api/advisor/clients/{clientId}/assets/{id}", handleUpdateAsset)
	clientContextMux.HandleFunc("DELETE /api/advisor/clients/{clientId}/assets/{id}", handleDeleteAsset)
	clientContextMux.HandleFunc("GET /api/advisor/clients/{clientId}/debts", handleGetDebts)
	clientContextMux.HandleFunc("POST /api/advisor/clients/{clientId}/debts", handleCreateDebt)
	clientContextMux.HandleFunc("PUT /api/advisor/clients/{clientId}/debts/{id}", handleUpdateDebt)
	clientContextMux.HandleFunc("DELETE /api/advisor/clients/{clientId}/debts/{id}", handleDeleteDebt)
	clientContextMux.HandleFunc("POST /api/advisor/clients/{clientId}/monte-carlo", handleMonteCarlo)
	clientContextMux.HandleFunc("POST /api/advisor/clients/{clientId}/monte-carlo/scenarios", handleScenarioComparison)
	clientContextMux.HandleFunc("GET /api/advisor/clients/{clientId}/simulations", handleListSimulations)
	clientContextMux.HandleFunc("GET /api/advisor/clients/{clientId}/simulations/{id}", handleGetSimulation)
	clientContextMux.HandleFunc("POST /api/advisor/clients/{clientId}/simulations", handleSaveSimulation)
	clientContextMux.HandleFunc("POST /api/advisor/clients/{clientId}/chat", handleChat)
	clientContextMux.HandleFunc("GET /api/advisor/clients/{clientId}/transactions", handleGetTransactions)
	clientContextMux.HandleFunc("GET /api/advisor/clients/{clientId}/transactions/summary", handleGetTransactionSummary)
	clientContextMux.HandleFunc("GET /api/advisor/clients/{clientId}/transactions/categories", handleGetCategories)
	clientContextMux.HandleFunc("POST /api/advisor/clients/{clientId}/reports/generate", handleGenerateReport)
	// Client notes routes (advisor-only, not visible to clients)
	clientContextMux.HandleFunc("GET /api/advisor/clients/{clientId}/notes", handleListClientNotes)
	clientContextMux.HandleFunc("POST /api/advisor/clients/{clientId}/notes", handleCreateClientNote)
	clientContextMux.HandleFunc("PUT /api/advisor/clients/{clientId}/notes/{noteId}", handleUpdateClientNote)
	clientContextMux.HandleFunc("DELETE /api/advisor/clients/{clientId}/notes/{noteId}", handleDeleteClientNote)
	// Client goals routes (visible to both advisors and clients)
	clientContextMux.HandleFunc("GET /api/advisor/clients/{clientId}/goals", handleListGoals)
	clientContextMux.HandleFunc("POST /api/advisor/clients/{clientId}/goals", handleCreateGoal)
	clientContextMux.HandleFunc("PUT /api/advisor/clients/{clientId}/goals/{goalId}", handleUpdateGoal)
	clientContextMux.HandleFunc("DELETE /api/advisor/clients/{clientId}/goals/{goalId}", handleDeleteGoal)

	// Apply auth middleware to protected routes
	mux.Handle("/api/auth/me", AuthMiddleware(protectedMux))
	mux.Handle("/api/assets", AuthMiddleware(protectedMux))
	mux.Handle("/api/assets/", AuthMiddleware(protectedMux))
	mux.Handle("/api/debts", AuthMiddleware(protectedMux))
	mux.Handle("/api/debts/", AuthMiddleware(protectedMux))
	mux.Handle("/api/monte-carlo", AuthMiddleware(protectedMux))
	mux.Handle("/api/simulations", AuthMiddleware(protectedMux))
	mux.Handle("/api/simulations/", AuthMiddleware(protectedMux))
	mux.Handle("/api/import/", AuthMiddleware(protectedMux))
	mux.Handle("/api/plaid/", AuthMiddleware(protectedMux))
	mux.Handle("/api/transactions", AuthMiddleware(protectedMux))
	mux.Handle("/api/transactions/", AuthMiddleware(protectedMux))
	mux.Handle("/api/chat", AuthMiddleware(protectedMux))
	mux.Handle("/api/invitations/", AuthMiddleware(protectedMux))
	mux.Handle("/api/reports/", AuthMiddleware(protectedMux))
	mux.Handle("/api/messages/", AuthMiddleware(protectedMux))
	mux.Handle("/api/documents", AuthMiddleware(protectedMux))
	mux.Handle("/api/documents/", AuthMiddleware(protectedMux))
	mux.Handle("/api/goals", AuthMiddleware(protectedMux))
	mux.Handle("/api/goals/", AuthMiddleware(protectedMux))

	// Apply auth + advisor middleware to advisor routes
	mux.Handle("/api/advisor/clients", AuthMiddleware(AdvisorMiddleware(advisorMux)))
	mux.Handle("/api/advisor/clients/", AuthMiddleware(AdvisorMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a client context route (has clientId in path)
		// Routes like /api/advisor/clients/{clientId}/assets
		if clientIDPattern.MatchString(r.URL.Path) {
			ClientAccessMiddleware(clientContextMux).ServeHTTP(w, r)
		} else {
			advisorMux.ServeHTTP(w, r)
		}
	}))))

	// Admin routes (advisor-only) for managing advisors
	mux.Handle("/api/advisor/admin/", AuthMiddleware(AdvisorMiddleware(advisorMux)))

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
