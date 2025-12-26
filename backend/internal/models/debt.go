package models

import "time"

type Debt struct {
	ID             int       `json:"id" db:"id"`
	UserID         int       `json:"userId" db:"user_id"`
	Name           string    `json:"name" db:"name"`
	CurrentBalance float64   `json:"currentBalance" db:"current_balance"`
	InterestRate   *float64  `json:"interestRate" db:"interest_rate"`
	MinimumPayment *float64  `json:"minimumPayment" db:"minimum_payment"`
	PlaidAccountID *string   `json:"plaidAccountId,omitempty" db:"plaid_account_id"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
}

type CreateDebtRequest struct {
	Name           string  `json:"name"`
	CurrentBalance float64 `json:"currentBalance"`
	InterestRate   float64 `json:"interestRate"`
	MinimumPayment float64 `json:"minimumPayment"`
}

type UpdateDebtRequest struct {
	Name           *string  `json:"name,omitempty"`
	CurrentBalance *float64 `json:"currentBalance,omitempty"`
	InterestRate   *float64 `json:"interestRate,omitempty"`
	MinimumPayment *float64 `json:"minimumPayment,omitempty"`
}
