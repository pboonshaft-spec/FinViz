package ingestion

import (
	"github.com/finviz/backend/internal/models"
)

// Source represents the data source type
type Source string

const (
	SourceManual Source = "manual"
	SourceCSV    Source = "csv"
	SourcePlaid  Source = "plaid"
)

// Ingester defines the interface for data ingestion
type Ingester interface {
	// Name returns the ingester name
	Name() string

	// Source returns the source type
	Source() Source

	// ImportAssets imports assets and returns the count imported
	ImportAssets(data []byte) (int, error)

	// ImportDebts imports debts and returns the count imported
	ImportDebts(data []byte) (int, error)

	// SyncAccounts syncs accounts (for Plaid-like integrations)
	SyncAccounts() error
}

// AssetImport represents an asset to be imported
type AssetImport struct {
	Name             string
	TypeID           int
	CurrentValue     float64
	CustomReturn     *float64
	CustomVolatility *float64
}

// DebtImport represents a debt to be imported
type DebtImport struct {
	Name           string
	CurrentBalance float64
	InterestRate   *float64
	MinimumPayment *float64
}

// ImportResult contains the results of an import operation
type ImportResult struct {
	Imported int
	Skipped  int
	Errors   []string
	Assets   []models.Asset
	Debts    []models.Debt
}
