package models

import "time"

// Document represents a file stored in the document vault
type Document struct {
	ID           int        `json:"id"`
	UserID       int        `json:"user_id"`
	UploadedBy   int        `json:"uploaded_by"` // Can differ from UserID if advisor uploads for client
	Name         string     `json:"name"`
	OriginalName string     `json:"original_name"`
	MimeType     string     `json:"mime_type"`
	Size         int64      `json:"size"`
	Category     string     `json:"category"` // tax_returns, statements, estate_docs, insurance, other
	StoragePath  string     `json:"-"`        // Internal path, not exposed to API
	Encrypted    bool       `json:"encrypted"`
	Description  *string    `json:"description,omitempty"`
	Year         *int       `json:"year,omitempty"` // For tax documents, statements
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"` // Soft delete
}

// DocumentShare represents sharing permissions for a document
type DocumentShare struct {
	ID           int       `json:"id"`
	DocumentID   int       `json:"document_id"`
	SharedWithID int       `json:"shared_with_id"` // User ID of recipient
	SharedByID   int       `json:"shared_by_id"`   // User ID who shared
	Permission   string    `json:"permission"`     // view, download
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"` // Optional expiration
}

// DocumentVersion tracks version history
type DocumentVersion struct {
	ID          int       `json:"id"`
	DocumentID  int       `json:"document_id"`
	VersionNum  int       `json:"version_num"`
	StoragePath string    `json:"-"`
	Size        int64     `json:"size"`
	UploadedBy  int       `json:"uploaded_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// DocumentCategory constants
const (
	DocCategoryTaxReturns  = "tax_returns"
	DocCategoryStatements  = "statements"
	DocCategoryEstateDocs  = "estate_docs"
	DocCategoryInsurance   = "insurance"
	DocCategoryInvestments = "investments"
	DocCategoryReports     = "reports" // Auto-generated financial plan reports
	DocCategoryOther       = "other"
)

// Valid document categories
var ValidDocumentCategories = []string{
	DocCategoryTaxReturns,
	DocCategoryStatements,
	DocCategoryEstateDocs,
	DocCategoryInsurance,
	DocCategoryInvestments,
	DocCategoryReports,
	DocCategoryOther,
}

// DocumentUploadRequest represents a document upload request
type DocumentUploadRequest struct {
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Description *string `json:"description,omitempty"`
	Year        *int    `json:"year,omitempty"`
	ClientID    *int    `json:"client_id,omitempty"` // For advisor uploading to client
}

// DocumentListResponse represents the response for listing documents
type DocumentListResponse struct {
	Documents  []DocumentWithShares `json:"documents"`
	TotalCount int                  `json:"total_count"`
	Categories map[string]int       `json:"categories"` // Count per category
}

// DocumentWithShares includes sharing info
type DocumentWithShares struct {
	Document
	SharedWith   []DocumentShareInfo `json:"shared_with,omitempty"`
	UploadedByName string            `json:"uploaded_by_name,omitempty"`
	CanEdit      bool                `json:"can_edit"`
	CanDelete    bool                `json:"can_delete"`
	CanShare     bool                `json:"can_share"`
}

// DocumentShareInfo is a simplified share record for responses
type DocumentShareInfo struct {
	UserID     int       `json:"user_id"`
	UserName   string    `json:"user_name"`
	UserEmail  string    `json:"user_email"`
	Permission string    `json:"permission"`
	SharedAt   time.Time `json:"shared_at"`
}

// ShareDocumentRequest represents a request to share a document
type ShareDocumentRequest struct {
	ShareWithID int     `json:"share_with_id"`
	Permission  string  `json:"permission"` // view, download
	ExpiresIn   *int    `json:"expires_in,omitempty"` // Hours until expiration
}

// IsValidCategory checks if a category is valid
func IsValidCategory(category string) bool {
	for _, c := range ValidDocumentCategories {
		if c == category {
			return true
		}
	}
	return false
}
