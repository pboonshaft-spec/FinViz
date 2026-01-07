package models

import "time"

// Conversation represents a messaging thread between an advisor and client
type Conversation struct {
	ID                 int        `json:"id" db:"id"`
	AdvisorID          int        `json:"advisorId" db:"advisor_id"`
	ClientID           int        `json:"clientId" db:"client_id"`
	LastMessageAt      *time.Time `json:"lastMessageAt,omitempty" db:"last_message_at"`
	UnreadCountAdvisor int        `json:"unreadCountAdvisor" db:"unread_count_advisor"`
	UnreadCountClient  int        `json:"unreadCountClient" db:"unread_count_client"`
	CreatedAt          time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt          time.Time  `json:"updatedAt" db:"updated_at"`

	// Joined fields
	AdvisorName string `json:"advisorName,omitempty" db:"-"`
	ClientName  string `json:"clientName,omitempty" db:"-"`
	ClientEmail string `json:"clientEmail,omitempty" db:"-"`
}

// Message represents an E2E encrypted message
type Message struct {
	ID               int        `json:"id" db:"id"`
	ConversationID   int        `json:"conversationId" db:"conversation_id"`
	SenderID         int        `json:"senderId" db:"sender_id"`
	EncryptedContent string     `json:"encryptedContent" db:"encrypted_content"`
	Nonce            string     `json:"nonce" db:"nonce"`
	ReadAt           *time.Time `json:"readAt,omitempty" db:"read_at"`
	CreatedAt        time.Time  `json:"createdAt" db:"created_at"`

	// Joined/computed fields
	SenderName string `json:"senderName,omitempty" db:"-"`
	IsOwn      bool   `json:"isOwn,omitempty" db:"-"`
}

// UserPublicKey stores a user's public key for E2E encryption
type UserPublicKey struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"userId" db:"user_id"`
	PublicKey string    `json:"publicKey" db:"public_key"`
	KeyID     string    `json:"keyId" db:"key_id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// SendMessageRequest is the request body for sending a message
type SendMessageRequest struct {
	EncryptedContent string `json:"encryptedContent"`
	Nonce            string `json:"nonce"`
}

// ConversationWithLastMessage includes the last message preview
type ConversationWithLastMessage struct {
	Conversation
	LastMessagePreview  string `json:"lastMessagePreview,omitempty"`
	LastMessageSenderID int    `json:"lastMessageSenderId,omitempty"`
}

// RegisterPublicKeyRequest is the request body for registering a public key
type RegisterPublicKeyRequest struct {
	PublicKey string `json:"publicKey"`
	KeyID     string `json:"keyId"`
}

// UnreadCounts represents unread message counts for a user
type UnreadCounts struct {
	TotalUnread    int `json:"totalUnread"`
	Conversations  int `json:"conversations"`
}
