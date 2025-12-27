package models

import "time"

// PlaidItem represents a linked Plaid institution
type PlaidItem struct {
	ID              int       `json:"id" db:"id"`
	UserID          int       `json:"userId" db:"user_id"`
	ItemID          string    `json:"itemId" db:"item_id"`
	AccessToken     string    `json:"-" db:"access_token"` // Never expose
	InstitutionID   string    `json:"institutionId" db:"institution_id"`
	InstitutionName string    `json:"institutionName" db:"institution_name"`
	Status          string    `json:"status" db:"status"`
	CreatedAt       time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time `json:"updatedAt" db:"updated_at"`
}

// PlaidAccount represents a synced account from Plaid
type PlaidAccount struct {
	ID               int        `json:"id" db:"id"`
	PlaidItemID      int        `json:"plaidItemId" db:"plaid_item_id"`
	UserID           int        `json:"userId" db:"user_id"`
	AccountID        string     `json:"accountId" db:"account_id"`
	Name             string     `json:"name" db:"name"`
	OfficialName     *string    `json:"officialName,omitempty" db:"official_name"`
	Type             string     `json:"type" db:"type"`
	Subtype          *string    `json:"subtype,omitempty" db:"subtype"`
	CurrentBalance   *float64   `json:"currentBalance,omitempty" db:"current_balance"`
	AvailableBalance *float64   `json:"availableBalance,omitempty" db:"available_balance"`
	CreditLimit      *float64   `json:"creditLimit,omitempty" db:"credit_limit"`
	ISOCurrencyCode  *string    `json:"isoCurrencyCode,omitempty" db:"iso_currency_code"`
	LastSyncedAt     *time.Time `json:"lastSyncedAt,omitempty" db:"last_synced_at"`
	CreatedAt        time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt        time.Time  `json:"updatedAt" db:"updated_at"`
}

// LinkTokenRequest is the request to create a Plaid Link token
type LinkTokenRequest struct {
	// No body needed, user ID comes from auth context
}

// LinkTokenResponse is returned when creating a link token
type LinkTokenResponse struct {
	LinkToken  string    `json:"linkToken"`
	Expiration time.Time `json:"expiration"`
}

// ExchangeTokenRequest is the request to exchange a public token
type ExchangeTokenRequest struct {
	PublicToken string `json:"publicToken"`
}

// PlaidItemResponse is returned after linking an institution
type PlaidItemResponse struct {
	Item     PlaidItem      `json:"item"`
	Accounts []PlaidAccount `json:"accounts"`
}

// SyncResponse is returned after syncing accounts
type SyncResponse struct {
	SyncedAccounts int `json:"syncedAccounts"`
	NewAssets      int `json:"newAssets"`
	NewDebts       int `json:"newDebts"`
	UpdatedAssets  int `json:"updatedAssets"`
	UpdatedDebts   int `json:"updatedDebts"`
}
