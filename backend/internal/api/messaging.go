package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/finviz/backend/internal/db"
	"github.com/finviz/backend/internal/models"
)

// handleListConversations lists all conversations for the current user
func handleListConversations(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var rows *sql.Rows
	var err error

	if user.IsAdvisor() {
		// Advisor sees conversations with their clients
		rows, err = db.DB.Query(`
			SELECT c.id, c.advisor_id, c.client_id, c.last_message_at,
			       c.unread_count_advisor, c.unread_count_client, c.created_at, c.updated_at,
			       u.name as client_name, u.email as client_email
			FROM conversations c
			JOIN users u ON c.client_id = u.id
			WHERE c.advisor_id = ?
			ORDER BY COALESCE(c.last_message_at, c.created_at) DESC
		`, user.ID)
	} else {
		// Client sees conversations with their advisors
		rows, err = db.DB.Query(`
			SELECT c.id, c.advisor_id, c.client_id, c.last_message_at,
			       c.unread_count_advisor, c.unread_count_client, c.created_at, c.updated_at,
			       u.name as advisor_name
			FROM conversations c
			JOIN users u ON c.advisor_id = u.id
			WHERE c.client_id = ?
			ORDER BY COALESCE(c.last_message_at, c.created_at) DESC
		`, user.ID)
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch conversations")
		return
	}
	defer rows.Close()

	var conversations []models.Conversation
	for rows.Next() {
		var c models.Conversation
		if user.IsAdvisor() {
			err = rows.Scan(&c.ID, &c.AdvisorID, &c.ClientID, &c.LastMessageAt,
				&c.UnreadCountAdvisor, &c.UnreadCountClient, &c.CreatedAt, &c.UpdatedAt,
				&c.ClientName, &c.ClientEmail)
		} else {
			err = rows.Scan(&c.ID, &c.AdvisorID, &c.ClientID, &c.LastMessageAt,
				&c.UnreadCountAdvisor, &c.UnreadCountClient, &c.CreatedAt, &c.UpdatedAt,
				&c.AdvisorName)
		}
		if err != nil {
			continue
		}
		conversations = append(conversations, c)
	}

	if conversations == nil {
		conversations = []models.Conversation{}
	}

	respondJSON(w, http.StatusOK, conversations)
}

// handleGetConversation gets a specific conversation and its messages
func handleGetConversation(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Extract conversation ID from path
	convIDStr := r.PathValue("id")
	if convIDStr == "" {
		// Try to extract from URL path manually
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 3 {
			convIDStr = parts[len(parts)-1]
		}
	}

	convID, err := strconv.Atoi(convIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	// Verify user has access to this conversation
	var conv models.Conversation
	err = db.DB.QueryRow(`
		SELECT id, advisor_id, client_id, last_message_at,
		       unread_count_advisor, unread_count_client, created_at, updated_at
		FROM conversations
		WHERE id = ? AND (advisor_id = ? OR client_id = ?)
	`, convID, user.ID, user.ID).Scan(&conv.ID, &conv.AdvisorID, &conv.ClientID,
		&conv.LastMessageAt, &conv.UnreadCountAdvisor, &conv.UnreadCountClient,
		&conv.CreatedAt, &conv.UpdatedAt)

	if err != nil {
		respondError(w, http.StatusNotFound, "Conversation not found")
		return
	}

	// Get the other party's name
	var otherUserID int
	if user.ID == conv.AdvisorID {
		otherUserID = conv.ClientID
	} else {
		otherUserID = conv.AdvisorID
	}

	var otherName string
	db.DB.QueryRow(`SELECT name FROM users WHERE id = ?`, otherUserID).Scan(&otherName)

	if user.IsAdvisor() {
		conv.ClientName = otherName
	} else {
		conv.AdvisorName = otherName
	}

	respondJSON(w, http.StatusOK, conv)
}

// handleGetMessages gets messages for a conversation
func handleGetMessages(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Extract conversation ID from path
	convIDStr := r.PathValue("id")
	if convIDStr == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		for i, p := range parts {
			if p == "conversations" && i+1 < len(parts) {
				convIDStr = parts[i+1]
				break
			}
		}
	}

	convID, err := strconv.Atoi(convIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	// Verify user has access
	var accessCheck int
	err = db.DB.QueryRow(`
		SELECT 1 FROM conversations
		WHERE id = ? AND (advisor_id = ? OR client_id = ?)
	`, convID, user.ID, user.ID).Scan(&accessCheck)
	if err != nil {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Get pagination params
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	beforeID := 0
	if b := r.URL.Query().Get("before"); b != "" {
		if parsed, err := strconv.Atoi(b); err == nil {
			beforeID = parsed
		}
	}

	// Fetch messages
	var rows *sql.Rows
	if beforeID > 0 {
		rows, err = db.DB.Query(`
			SELECT m.id, m.conversation_id, m.sender_id, m.encrypted_content, m.nonce,
			       m.read_at, m.created_at, u.name as sender_name
			FROM messages m
			JOIN users u ON m.sender_id = u.id
			WHERE m.conversation_id = ? AND m.id < ?
			ORDER BY m.created_at DESC
			LIMIT ?
		`, convID, beforeID, limit)
	} else {
		rows, err = db.DB.Query(`
			SELECT m.id, m.conversation_id, m.sender_id, m.encrypted_content, m.nonce,
			       m.read_at, m.created_at, u.name as sender_name
			FROM messages m
			JOIN users u ON m.sender_id = u.id
			WHERE m.conversation_id = ?
			ORDER BY m.created_at DESC
			LIMIT ?
		`, convID, limit)
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch messages")
		return
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.EncryptedContent,
			&m.Nonce, &m.ReadAt, &m.CreatedAt, &m.SenderName); err != nil {
			continue
		}
		m.IsOwn = m.SenderID == user.ID
		messages = append(messages, m)
	}

	if messages == nil {
		messages = []models.Message{}
	}

	// Mark messages as read (from the other party)
	go markMessagesAsRead(convID, user.ID)

	respondJSON(w, http.StatusOK, messages)
}

// handleSendMessage sends a new encrypted message
func handleSendMessage(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Extract conversation ID
	convIDStr := r.PathValue("id")
	if convIDStr == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		for i, p := range parts {
			if p == "conversations" && i+1 < len(parts) {
				convIDStr = parts[i+1]
				break
			}
		}
	}

	convID, err := strconv.Atoi(convIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	// Verify user has access and get conversation details
	var conv struct {
		AdvisorID int
		ClientID  int
	}
	err = db.DB.QueryRow(`
		SELECT advisor_id, client_id FROM conversations
		WHERE id = ? AND (advisor_id = ? OR client_id = ?)
	`, convID, user.ID, user.ID).Scan(&conv.AdvisorID, &conv.ClientID)
	if err != nil {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	var req models.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.EncryptedContent == "" || req.Nonce == "" {
		respondError(w, http.StatusBadRequest, "Encrypted content and nonce are required")
		return
	}

	// Insert message
	result, err := db.DB.Exec(`
		INSERT INTO messages (conversation_id, sender_id, encrypted_content, nonce)
		VALUES (?, ?, ?, ?)
	`, convID, user.ID, req.EncryptedContent, req.Nonce)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to send message")
		return
	}

	msgID, _ := result.LastInsertId()

	// Update conversation last_message_at and increment unread count
	if user.ID == conv.AdvisorID {
		// Advisor sent message, increment client's unread count
		db.DB.Exec(`
			UPDATE conversations
			SET last_message_at = NOW(), unread_count_client = unread_count_client + 1
			WHERE id = ?
		`, convID)
	} else {
		// Client sent message, increment advisor's unread count
		db.DB.Exec(`
			UPDATE conversations
			SET last_message_at = NOW(), unread_count_advisor = unread_count_advisor + 1
			WHERE id = ?
		`, convID)
	}

	// Return the created message
	var msg models.Message
	db.DB.QueryRow(`
		SELECT m.id, m.conversation_id, m.sender_id, m.encrypted_content, m.nonce,
		       m.read_at, m.created_at, u.name as sender_name
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.id = ?
	`, msgID).Scan(&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.EncryptedContent,
		&msg.Nonce, &msg.ReadAt, &msg.CreatedAt, &msg.SenderName)
	msg.IsOwn = true

	respondJSON(w, http.StatusCreated, msg)
}

// handleStartConversation starts a new conversation (advisor to client)
func handleStartConversation(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req struct {
		ClientID int `json:"clientId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var advisorID, clientID int

	if user.IsAdvisor() {
		advisorID = user.ID
		clientID = req.ClientID

		// Verify advisor has access to this client
		var relationID int
		err := db.DB.QueryRow(`
			SELECT id FROM advisor_clients
			WHERE advisor_id = ? AND client_id = ? AND status = 'active'
		`, advisorID, clientID).Scan(&relationID)
		if err != nil {
			respondError(w, http.StatusForbidden, "You don't have access to this client")
			return
		}
	} else {
		// Client starting conversation - find their advisor
		clientID = user.ID
		err := db.DB.QueryRow(`
			SELECT advisor_id FROM advisor_clients
			WHERE client_id = ? AND status = 'active'
			LIMIT 1
		`, clientID).Scan(&advisorID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "No advisor found")
			return
		}
	}

	// Check if conversation already exists
	var existingID int
	err := db.DB.QueryRow(`
		SELECT id FROM conversations
		WHERE advisor_id = ? AND client_id = ?
	`, advisorID, clientID).Scan(&existingID)

	if err == nil {
		// Conversation exists, return it
		var conv models.Conversation
		db.DB.QueryRow(`
			SELECT id, advisor_id, client_id, last_message_at,
			       unread_count_advisor, unread_count_client, created_at, updated_at
			FROM conversations WHERE id = ?
		`, existingID).Scan(&conv.ID, &conv.AdvisorID, &conv.ClientID,
			&conv.LastMessageAt, &conv.UnreadCountAdvisor, &conv.UnreadCountClient,
			&conv.CreatedAt, &conv.UpdatedAt)

		respondJSON(w, http.StatusOK, conv)
		return
	}

	// Create new conversation
	result, err := db.DB.Exec(`
		INSERT INTO conversations (advisor_id, client_id)
		VALUES (?, ?)
	`, advisorID, clientID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create conversation")
		return
	}

	convID, _ := result.LastInsertId()

	var conv models.Conversation
	db.DB.QueryRow(`
		SELECT id, advisor_id, client_id, last_message_at,
		       unread_count_advisor, unread_count_client, created_at, updated_at
		FROM conversations WHERE id = ?
	`, convID).Scan(&conv.ID, &conv.AdvisorID, &conv.ClientID,
		&conv.LastMessageAt, &conv.UnreadCountAdvisor, &conv.UnreadCountClient,
		&conv.CreatedAt, &conv.UpdatedAt)

	// Get names
	if user.IsAdvisor() {
		db.DB.QueryRow(`SELECT name FROM users WHERE id = ?`, clientID).Scan(&conv.ClientName)
	} else {
		db.DB.QueryRow(`SELECT name FROM users WHERE id = ?`, advisorID).Scan(&conv.AdvisorName)
	}

	respondJSON(w, http.StatusCreated, conv)
}

// handleMarkAsRead marks messages in a conversation as read
func handleMarkAsRead(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	convIDStr := r.PathValue("id")
	if convIDStr == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		for i, p := range parts {
			if p == "conversations" && i+1 < len(parts) {
				convIDStr = parts[i+1]
				break
			}
		}
	}

	convID, err := strconv.Atoi(convIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	markMessagesAsRead(convID, user.ID)

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleGetUnreadCounts returns unread message counts for the user
func handleGetUnreadCounts(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var counts models.UnreadCounts

	if user.IsAdvisor() {
		db.DB.QueryRow(`
			SELECT COALESCE(SUM(unread_count_advisor), 0), COUNT(CASE WHEN unread_count_advisor > 0 THEN 1 END)
			FROM conversations WHERE advisor_id = ?
		`, user.ID).Scan(&counts.TotalUnread, &counts.Conversations)
	} else {
		db.DB.QueryRow(`
			SELECT COALESCE(SUM(unread_count_client), 0), COUNT(CASE WHEN unread_count_client > 0 THEN 1 END)
			FROM conversations WHERE client_id = ?
		`, user.ID).Scan(&counts.TotalUnread, &counts.Conversations)
	}

	respondJSON(w, http.StatusOK, counts)
}

// handleRegisterPublicKey registers a user's public key for E2E encryption
func handleRegisterPublicKey(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req models.RegisterPublicKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.PublicKey == "" || req.KeyID == "" {
		respondError(w, http.StatusBadRequest, "Public key and key ID are required")
		return
	}

	// Upsert the public key
	_, err := db.DB.Exec(`
		INSERT INTO user_public_keys (user_id, public_key, key_id)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE public_key = VALUES(public_key)
	`, user.ID, req.PublicKey, req.KeyID)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to register public key")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleGetPublicKey gets a user's public key
func handleGetPublicKey(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	userIDStr := r.PathValue("userId")
	if userIDStr == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		userIDStr = parts[len(parts)-1]
	}

	targetUserID, err := strconv.Atoi(userIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Verify caller has a relationship with target user
	var hasAccess bool
	if user.IsAdvisor() {
		var count int
		db.DB.QueryRow(`
			SELECT COUNT(*) FROM advisor_clients
			WHERE advisor_id = ? AND client_id = ? AND status = 'active'
		`, user.ID, targetUserID).Scan(&count)
		hasAccess = count > 0 || user.ID == targetUserID
	} else {
		var count int
		db.DB.QueryRow(`
			SELECT COUNT(*) FROM advisor_clients
			WHERE client_id = ? AND advisor_id = ? AND status = 'active'
		`, user.ID, targetUserID).Scan(&count)
		hasAccess = count > 0 || user.ID == targetUserID
	}

	if !hasAccess {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	var key models.UserPublicKey
	err = db.DB.QueryRow(`
		SELECT id, user_id, public_key, key_id, created_at
		FROM user_public_keys
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, targetUserID).Scan(&key.ID, &key.UserID, &key.PublicKey, &key.KeyID, &key.CreatedAt)

	if err != nil {
		respondError(w, http.StatusNotFound, "Public key not found")
		return
	}

	respondJSON(w, http.StatusOK, key)
}

// markMessagesAsRead marks all messages from the other party as read
func markMessagesAsRead(convID, userID int) {
	now := time.Now()

	// Mark messages from others as read
	db.DB.Exec(`
		UPDATE messages
		SET read_at = ?
		WHERE conversation_id = ? AND sender_id != ? AND read_at IS NULL
	`, now, convID, userID)

	// Reset unread count
	var advisorID, clientID int
	db.DB.QueryRow(`
		SELECT advisor_id, client_id FROM conversations WHERE id = ?
	`, convID).Scan(&advisorID, &clientID)

	if userID == advisorID {
		db.DB.Exec(`UPDATE conversations SET unread_count_advisor = 0 WHERE id = ?`, convID)
	} else {
		db.DB.Exec(`UPDATE conversations SET unread_count_client = 0 WHERE id = ?`, convID)
	}
}
