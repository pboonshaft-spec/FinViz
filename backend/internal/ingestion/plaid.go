package ingestion

import (
	"errors"
)

// PlaidIngester is a stub for future Plaid integration
type PlaidIngester struct {
	clientID string
	secret   string
	env      string // sandbox, development, production
}

// NewPlaidIngester creates a new Plaid ingester
// This is a stub - actual Plaid integration will require:
// - github.com/plaid/plaid-go SDK
// - Plaid API credentials
// - Link token generation for frontend
// - Webhook handling for updates
func NewPlaidIngester(clientID, secret, env string) *PlaidIngester {
	return &PlaidIngester{
		clientID: clientID,
		secret:   secret,
		env:      env,
	}
}

func (p *PlaidIngester) Name() string {
	return "Plaid"
}

func (p *PlaidIngester) Source() Source {
	return SourcePlaid
}

// ImportAssets is not used for Plaid - use SyncAccounts instead
func (p *PlaidIngester) ImportAssets(data []byte) (int, error) {
	return 0, errors.New("plaid integration uses SyncAccounts, not ImportAssets")
}

// ImportDebts is not used for Plaid - use SyncAccounts instead
func (p *PlaidIngester) ImportDebts(data []byte) (int, error) {
	return 0, errors.New("plaid integration uses SyncAccounts, not ImportDebts")
}

// SyncAccounts syncs all connected accounts from Plaid
// Stub implementation - actual integration will:
// 1. Call Plaid's accounts/get endpoint
// 2. Map account types to asset/debt types
// 3. Update or create records in database
// 4. Handle investment accounts separately
func (p *PlaidIngester) SyncAccounts() error {
	return errors.New("plaid integration not yet implemented")
}

// Future methods to implement:
//
// CreateLinkToken generates a Plaid Link token for the frontend
// func (p *PlaidIngester) CreateLinkToken(userID string) (string, error)
//
// ExchangePublicToken exchanges a public token for an access token
// func (p *PlaidIngester) ExchangePublicToken(publicToken string) (string, error)
//
// GetInvestments fetches investment holdings
// func (p *PlaidIngester) GetInvestments(accessToken string) ([]AssetImport, error)
//
// GetLiabilities fetches credit cards, loans, mortgages
// func (p *PlaidIngester) GetLiabilities(accessToken string) ([]DebtImport, error)
