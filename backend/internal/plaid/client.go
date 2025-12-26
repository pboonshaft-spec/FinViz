package plaid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Client handles Plaid API requests
type Client struct {
	clientID   string
	secret     string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Plaid client
func NewClient() *Client {
	env := os.Getenv("PLAID_ENV")
	if env == "" {
		env = "sandbox"
	}

	var baseURL string
	switch env {
	case "sandbox":
		baseURL = "https://sandbox.plaid.com"
	case "development":
		baseURL = "https://development.plaid.com"
	case "production":
		baseURL = "https://production.plaid.com"
	default:
		baseURL = "https://sandbox.plaid.com"
	}

	return &Client{
		clientID:   os.Getenv("PLAID_CLIENT_ID"),
		secret:     os.Getenv("PLAID_SECRET"),
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// IsConfigured returns true if Plaid credentials are set
func (c *Client) IsConfigured() bool {
	return c.clientID != "" && c.secret != ""
}

func (c *Client) post(endpoint string, body interface{}) ([]byte, error) {
	// Add credentials to body
	bodyMap := make(map[string]interface{})

	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(bodyBytes, &bodyMap); err != nil {
			return nil, err
		}
	}

	bodyMap["client_id"] = c.clientID
	bodyMap["secret"] = c.secret

	jsonBody, err := json.Marshal(bodyMap)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		var plaidErr PlaidError
		if err := json.Unmarshal(respBody, &plaidErr); err == nil && plaidErr.ErrorMessage != "" {
			return nil, fmt.Errorf("plaid error: %s - %s", plaidErr.ErrorCode, plaidErr.ErrorMessage)
		}
		return nil, fmt.Errorf("plaid API error: %d - %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// PlaidError represents a Plaid API error
type PlaidError struct {
	ErrorType    string `json:"error_type"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	DisplayMsg   string `json:"display_message"`
}

// CreateLinkToken creates a Link token for initializing Plaid Link
func (c *Client) CreateLinkToken(userID string) (*LinkTokenResponse, error) {
	body := map[string]interface{}{
		"user": map[string]string{
			"client_user_id": userID,
		},
		"client_name":   "FinViz",
		"products":      []string{"transactions"},
		"country_codes": []string{"US"},
		"language":      "en",
	}

	resp, err := c.post("/link/token/create", body)
	if err != nil {
		return nil, err
	}

	var result LinkTokenResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// LinkTokenResponse from Plaid
type LinkTokenResponse struct {
	LinkToken  string `json:"link_token"`
	Expiration string `json:"expiration"`
}

// ExchangePublicToken exchanges a public token for an access token
func (c *Client) ExchangePublicToken(publicToken string) (*ExchangeResponse, error) {
	body := map[string]interface{}{
		"public_token": publicToken,
	}

	resp, err := c.post("/item/public_token/exchange", body)
	if err != nil {
		return nil, err
	}

	var result ExchangeResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ExchangeResponse from Plaid
type ExchangeResponse struct {
	AccessToken string `json:"access_token"`
	ItemID      string `json:"item_id"`
}

// GetAccounts retrieves accounts for an item
func (c *Client) GetAccounts(accessToken string) (*AccountsResponse, error) {
	body := map[string]interface{}{
		"access_token": accessToken,
	}

	resp, err := c.post("/accounts/get", body)
	if err != nil {
		return nil, err
	}

	var result AccountsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// AccountsResponse from Plaid
type AccountsResponse struct {
	Accounts []Account `json:"accounts"`
	Item     Item      `json:"item"`
}

// Account represents a Plaid account
type Account struct {
	AccountID    string   `json:"account_id"`
	Name         string   `json:"name"`
	OfficialName string   `json:"official_name"`
	Type         string   `json:"type"`
	Subtype      string   `json:"subtype"`
	Balances     Balances `json:"balances"`
}

// Balances for an account
type Balances struct {
	Current          *float64 `json:"current"`
	Available        *float64 `json:"available"`
	Limit            *float64 `json:"limit"`
	ISOCurrencyCode  string   `json:"iso_currency_code"`
}

// Item represents a Plaid item
type Item struct {
	ItemID        string `json:"item_id"`
	InstitutionID string `json:"institution_id"`
}

// GetInstitution gets institution details
func (c *Client) GetInstitution(institutionID string) (*InstitutionResponse, error) {
	body := map[string]interface{}{
		"institution_id": institutionID,
		"country_codes":  []string{"US"},
	}

	resp, err := c.post("/institutions/get_by_id", body)
	if err != nil {
		return nil, err
	}

	var result InstitutionResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// InstitutionResponse from Plaid
type InstitutionResponse struct {
	Institution Institution `json:"institution"`
}

// Institution details
type Institution struct {
	InstitutionID string `json:"institution_id"`
	Name          string `json:"name"`
}

// GetLiabilities retrieves liabilities for an item
func (c *Client) GetLiabilities(accessToken string) (*LiabilitiesResponse, error) {
	body := map[string]interface{}{
		"access_token": accessToken,
	}

	resp, err := c.post("/liabilities/get", body)
	if err != nil {
		return nil, err
	}

	var result LiabilitiesResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// LiabilitiesResponse from Plaid
type LiabilitiesResponse struct {
	Accounts    []Account   `json:"accounts"`
	Liabilities Liabilities `json:"liabilities"`
}

// Liabilities container
type Liabilities struct {
	Credit   []CreditLiability   `json:"credit"`
	Mortgage []MortgageLiability `json:"mortgage"`
	Student  []StudentLiability  `json:"student"`
}

// CreditLiability for credit cards
type CreditLiability struct {
	AccountID         string   `json:"account_id"`
	APRs              []APR    `json:"aprs"`
	IsOverdue         bool     `json:"is_overdue"`
	LastPaymentAmount *float64 `json:"last_payment_amount"`
	LastPaymentDate   string   `json:"last_payment_date"`
	MinimumPayment    *float64 `json:"minimum_payment_amount"`
}

// APR details
type APR struct {
	APRPercentage float64 `json:"apr_percentage"`
	APRType       string  `json:"apr_type"`
}

// MortgageLiability for mortgages
type MortgageLiability struct {
	AccountID                  string   `json:"account_id"`
	CurrentLateFee             *float64 `json:"current_late_fee"`
	InterestRatePercentage     float64  `json:"interest_rate_percentage"`
	LastPaymentAmount          *float64 `json:"last_payment_amount"`
	NextMonthlyPayment         *float64 `json:"next_monthly_payment"`
	OriginationPrincipalAmount *float64 `json:"origination_principal_amount"`
}

// StudentLiability for student loans
type StudentLiability struct {
	AccountID              string   `json:"account_id"`
	InterestRatePercentage float64  `json:"interest_rate_percentage"`
	MinimumPaymentAmount   *float64 `json:"minimum_payment_amount"`
	OriginationDate        string   `json:"origination_date"`
}

// GetTransactions retrieves transactions for an item
func (c *Client) GetTransactions(accessToken string, startDate, endDate string) (*TransactionsResponse, error) {
	body := map[string]interface{}{
		"access_token": accessToken,
		"start_date":   startDate,
		"end_date":     endDate,
		"options": map[string]interface{}{
			"count":  500,
			"offset": 0,
		},
	}

	resp, err := c.post("/transactions/get", body)
	if err != nil {
		return nil, err
	}

	var result TransactionsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// TransactionsResponse from Plaid
type TransactionsResponse struct {
	Accounts          []Account     `json:"accounts"`
	Transactions      []Transaction `json:"transactions"`
	TotalTransactions int           `json:"total_transactions"`
}

// Transaction represents a Plaid transaction
type Transaction struct {
	TransactionID       string   `json:"transaction_id"`
	AccountID           string   `json:"account_id"`
	Amount              float64  `json:"amount"`
	Date                string   `json:"date"`
	Name                string   `json:"name"`
	MerchantName        *string  `json:"merchant_name"`
	Category            []string `json:"category"`
	PersonalFinanceCat  *PFCat   `json:"personal_finance_category"`
	Pending             bool     `json:"pending"`
	TransactionType     string   `json:"transaction_type"`
	ISOCurrencyCode     string   `json:"iso_currency_code"`
	PaymentChannel      string   `json:"payment_channel"`
}

// PFCat is Plaid's personal finance category
type PFCat struct {
	Primary   string `json:"primary"`
	Detailed  string `json:"detailed"`
}
