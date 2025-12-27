package models

import "time"

type Transaction struct {
	ID                 int       `json:"id" db:"id"`
	UserID             int       `json:"userId" db:"user_id"`
	PlaidTransactionID *string   `json:"plaidTransactionId,omitempty" db:"plaid_transaction_id"`
	PlaidAccountID     *string   `json:"plaidAccountId,omitempty" db:"plaid_account_id"`
	AccountName        *string   `json:"accountName,omitempty" db:"account_name"`
	Amount             float64   `json:"amount" db:"amount"`
	Date               string    `json:"date" db:"date"`
	Name               string    `json:"name" db:"name"`
	MerchantName       *string   `json:"merchantName,omitempty" db:"merchant_name"`
	Category           *string   `json:"category,omitempty" db:"category"`
	Subcategory        *string   `json:"subcategory,omitempty" db:"subcategory"`
	Pending            bool      `json:"pending" db:"pending"`
	TransactionType    *string   `json:"transactionType,omitempty" db:"transaction_type"`
	ISOCurrencyCode    *string   `json:"isoCurrencyCode,omitempty" db:"iso_currency_code"`
	CreatedAt          time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt          time.Time `json:"updatedAt" db:"updated_at"`
}

type TransactionSummary struct {
	TotalIncome   float64           `json:"totalIncome"`
	TotalExpenses float64           `json:"totalExpenses"`
	NetCashFlow   float64           `json:"netCashFlow"`
	ByCategory    []CategorySummary `json:"byCategory"`
	ByMonth       []MonthSummary    `json:"byMonth"`
}

type CategorySummary struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Count    int     `json:"count"`
}

type MonthSummary struct {
	Month    string  `json:"month"`
	Income   float64 `json:"income"`
	Expenses float64 `json:"expenses"`
	Net      float64 `json:"net"`
}

type SyncTransactionsResponse struct {
	NewTransactions     int `json:"newTransactions"`
	UpdatedTransactions int `json:"updatedTransactions"`
	RemovedTransactions int `json:"removedTransactions"`
}
