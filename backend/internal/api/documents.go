package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/finviz/backend/internal/db"
	"github.com/finviz/backend/internal/models"
	"github.com/finviz/backend/internal/storage"
)

// Maximum file size: 25MB
const maxFileSize = 25 << 20

// Allowed MIME types
var allowedMimeTypes = map[string]bool{
	"application/pdf":                                                        true,
	"image/jpeg":                                                              true,
	"image/png":                                                               true,
	"image/gif":                                                               true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":       true, // xlsx
	"application/vnd.ms-excel":                                                true, // xls
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true, // docx
	"application/msword":                                                      true, // doc
	"text/csv":                                                                true,
	"text/plain":                                                              true,
}

// HandleDocumentUpload handles file uploads
func HandleDocumentUpload(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		http.Error(w, "File too large (max 25MB)", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check file size
	if header.Size > maxFileSize {
		http.Error(w, "File too large (max 25MB)", http.StatusBadRequest)
		return
	}

	// Get metadata from form
	name := r.FormValue("name")
	if name == "" {
		name = header.Filename
	}
	category := r.FormValue("category")
	if category == "" {
		category = models.DocCategoryOther
	}
	if !models.IsValidCategory(category) {
		http.Error(w, "Invalid category", http.StatusBadRequest)
		return
	}

	description := r.FormValue("description")
	yearStr := r.FormValue("year")
	clientIDStr := r.FormValue("client_id")

	// Determine target user ID (for advisor uploading to client)
	targetUserID := user.ID
	uploadedBy := user.ID

	if clientIDStr != "" && user.Role == "advisor" {
		clientID, err := strconv.Atoi(clientIDStr)
		if err == nil {
			// Verify advisor has access to client
			var accessLevel string
			err = db.DB.QueryRow(`
				SELECT access_level FROM advisor_clients
				WHERE advisor_id = ? AND client_id = ? AND status = 'active'
			`, user.ID, clientID).Scan(&accessLevel)
			if err == nil && (accessLevel == "edit" || accessLevel == "full") {
				targetUserID = clientID
			}
		}
	}

	// Detect MIME type
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" || mimeType == "application/octet-stream" {
		// Try to detect from extension
		ext := strings.ToLower(strings.TrimPrefix(header.Filename[strings.LastIndex(header.Filename, "."):], "."))
		switch ext {
		case "pdf":
			mimeType = "application/pdf"
		case "jpg", "jpeg":
			mimeType = "image/jpeg"
		case "png":
			mimeType = "image/png"
		case "xlsx":
			mimeType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		case "csv":
			mimeType = "text/csv"
		case "docx":
			mimeType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		default:
			mimeType = "application/octet-stream"
		}
	}

	if !allowedMimeTypes[mimeType] {
		http.Error(w, "File type not allowed", http.StatusBadRequest)
		return
	}

	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Save to storage (encrypted)
	storagePath, err := storage.DefaultStorage.Save(fileBytes, header.Filename, true)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Parse year if provided
	var year *int
	if yearStr != "" {
		y, err := strconv.Atoi(yearStr)
		if err == nil {
			year = &y
		}
	}

	// Save document record to database
	var descPtr *string
	if description != "" {
		descPtr = &description
	}

	result, err := db.DB.Exec(`
		INSERT INTO documents (user_id, uploaded_by, name, original_name, mime_type, size, category, storage_path, encrypted, description, year)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, TRUE, ?, ?)
	`, targetUserID, uploadedBy, name, header.Filename, mimeType, header.Size, category, storagePath, descPtr, year)

	if err != nil {
		// Clean up stored file on DB error
		storage.DefaultStorage.Delete(storagePath)
		http.Error(w, "Failed to save document record", http.StatusInternalServerError)
		return
	}

	docID, _ := result.LastInsertId()

	// If advisor uploaded for client, auto-share with client for download
	if targetUserID != uploadedBy {
		db.DB.Exec(`
			INSERT INTO document_shares (document_id, shared_with_id, shared_by_id, permission)
			VALUES (?, ?, ?, 'download')
		`, docID, targetUserID, uploadedBy)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       docID,
		"name":     name,
		"category": category,
		"size":     header.Size,
		"message":  "Document uploaded successfully",
	})
}

// HandleDocumentList lists documents for a user
func HandleDocumentList(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get optional filters
	category := r.URL.Query().Get("category")
	clientIDStr := r.URL.Query().Get("client_id")

	targetUserID := user.ID

	// For advisors viewing client documents
	if clientIDStr != "" && user.Role == "advisor" {
		clientID, err := strconv.Atoi(clientIDStr)
		if err == nil {
			// Verify advisor has access
			var accessLevel string
			err = db.DB.QueryRow(`
				SELECT access_level FROM advisor_clients
				WHERE advisor_id = ? AND client_id = ? AND status = 'active'
			`, user.ID, clientID).Scan(&accessLevel)
			if err == nil {
				targetUserID = clientID
			}
		}
	}

	// Build query
	query := `
		SELECT d.id, d.user_id, d.uploaded_by, d.name, d.original_name, d.mime_type,
		       d.size, d.category, d.encrypted, d.description, d.year, d.created_at, d.updated_at,
		       u.name as uploader_name
		FROM documents d
		LEFT JOIN users u ON d.uploaded_by = u.id
		WHERE d.user_id = ? AND d.deleted_at IS NULL
	`
	args := []interface{}{targetUserID}

	if category != "" && models.IsValidCategory(category) {
		query += " AND d.category = ?"
		args = append(args, category)
	}

	query += " ORDER BY d.created_at DESC"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to fetch documents", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var documents []models.DocumentWithShares
	categoryCount := make(map[string]int)

	for rows.Next() {
		var doc models.DocumentWithShares
		var uploaderName *string
		var description *string
		var year *int

		err := rows.Scan(
			&doc.ID, &doc.UserID, &doc.UploadedBy, &doc.Name, &doc.OriginalName,
			&doc.MimeType, &doc.Size, &doc.Category, &doc.Encrypted,
			&description, &year, &doc.CreatedAt, &doc.UpdatedAt, &uploaderName,
		)
		if err != nil {
			continue
		}

		doc.Description = description
		doc.Year = year
		if uploaderName != nil {
			doc.UploadedByName = *uploaderName
		}

		// Set permissions
		doc.CanEdit = doc.UploadedBy == user.ID || user.Role == "advisor"
		doc.CanDelete = doc.UploadedBy == user.ID || user.Role == "advisor"
		doc.CanShare = doc.UploadedBy == user.ID || user.Role == "advisor"

		documents = append(documents, doc)
		categoryCount[doc.Category]++
	}

	// Also include documents shared with this user
	if clientIDStr == "" {
		sharedRows, err := db.DB.Query(`
			SELECT d.id, d.user_id, d.uploaded_by, d.name, d.original_name, d.mime_type,
			       d.size, d.category, d.encrypted, d.description, d.year, d.created_at, d.updated_at,
			       u.name as uploader_name, ds.permission
			FROM documents d
			JOIN document_shares ds ON d.id = ds.document_id
			LEFT JOIN users u ON d.uploaded_by = u.id
			WHERE ds.shared_with_id = ? AND d.deleted_at IS NULL
			  AND (ds.expires_at IS NULL OR ds.expires_at > NOW())
			ORDER BY d.created_at DESC
		`, user.ID)

		if err == nil {
			defer sharedRows.Close()
			for sharedRows.Next() {
				var doc models.DocumentWithShares
				var uploaderName *string
				var description *string
				var year *int
				var permission string

				err := sharedRows.Scan(
					&doc.ID, &doc.UserID, &doc.UploadedBy, &doc.Name, &doc.OriginalName,
					&doc.MimeType, &doc.Size, &doc.Category, &doc.Encrypted,
					&description, &year, &doc.CreatedAt, &doc.UpdatedAt, &uploaderName, &permission,
				)
				if err != nil {
					continue
				}

				doc.Description = description
				doc.Year = year
				if uploaderName != nil {
					doc.UploadedByName = *uploaderName
				}

				// Shared docs have limited permissions
				doc.CanEdit = false
				doc.CanDelete = false
				doc.CanShare = false

				documents = append(documents, doc)
				categoryCount[doc.Category]++
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.DocumentListResponse{
		Documents:  documents,
		TotalCount: len(documents),
		Categories: categoryCount,
	})
}

// HandleDocumentDownload handles document download
func HandleDocumentDownload(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get document ID from path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	docID, err := strconv.Atoi(parts[len(parts)-2]) // /api/documents/{id}/download
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	// Check access
	var doc models.Document
	err = db.DB.QueryRow(`
		SELECT id, user_id, uploaded_by, name, original_name, mime_type, size, storage_path, encrypted
		FROM documents
		WHERE id = ? AND deleted_at IS NULL
	`, docID).Scan(&doc.ID, &doc.UserID, &doc.UploadedBy, &doc.Name, &doc.OriginalName, &doc.MimeType, &doc.Size, &doc.StoragePath, &doc.Encrypted)

	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Check permission: owner, uploader, shared, or advisor with access
	hasAccess := doc.UserID == user.ID || doc.UploadedBy == user.ID

	if !hasAccess {
		// Check if shared
		var shareCount int
		db.DB.QueryRow(`
			SELECT COUNT(*) FROM document_shares
			WHERE document_id = ? AND shared_with_id = ?
			  AND (expires_at IS NULL OR expires_at > NOW())
		`, docID, user.ID).Scan(&shareCount)
		hasAccess = shareCount > 0
	}

	if !hasAccess && user.Role == "advisor" {
		// Check if advisor has access to document owner
		var accessLevel string
		db.DB.QueryRow(`
			SELECT access_level FROM advisor_clients
			WHERE advisor_id = ? AND client_id = ? AND status = 'active'
		`, user.ID, doc.UserID).Scan(&accessLevel)
		hasAccess = accessLevel != ""
	}

	if !hasAccess {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Load file from storage
	data, err := storage.DefaultStorage.Load(doc.StoragePath, doc.Encrypted)
	if err != nil {
		http.Error(w, "Failed to load file", http.StatusInternalServerError)
		return
	}

	// Set headers for download
	w.Header().Set("Content-Type", doc.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, doc.OriginalName))
	w.Header().Set("Content-Length", strconv.FormatInt(doc.Size, 10))
	w.Write(data)
}

// HandleDocumentDelete handles document deletion (soft delete)
func HandleDocumentDelete(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get document ID from path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	docID, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	// Check ownership or advisor access
	var doc models.Document
	err = db.DB.QueryRow(`
		SELECT id, user_id, uploaded_by FROM documents WHERE id = ? AND deleted_at IS NULL
	`, docID).Scan(&doc.ID, &doc.UserID, &doc.UploadedBy)

	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	canDelete := doc.UploadedBy == user.ID

	if !canDelete && user.Role == "advisor" {
		var accessLevel string
		db.DB.QueryRow(`
			SELECT access_level FROM advisor_clients
			WHERE advisor_id = ? AND client_id = ? AND status = 'active'
		`, user.ID, doc.UserID).Scan(&accessLevel)
		canDelete = accessLevel == "full"
	}

	if !canDelete {
		http.Error(w, "Cannot delete this document", http.StatusForbidden)
		return
	}

	// Soft delete
	_, err = db.DB.Exec(`UPDATE documents SET deleted_at = NOW() WHERE id = ?`, docID)
	if err != nil {
		http.Error(w, "Failed to delete document", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Document deleted"})
}

// HandleDocumentShare handles sharing a document
func HandleDocumentShare(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get document ID from path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	docID, err := strconv.Atoi(parts[len(parts)-2]) // /api/documents/{id}/share
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	// Check ownership
	var doc models.Document
	err = db.DB.QueryRow(`
		SELECT id, user_id, uploaded_by FROM documents WHERE id = ? AND deleted_at IS NULL
	`, docID).Scan(&doc.ID, &doc.UserID, &doc.UploadedBy)

	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	canShare := doc.UploadedBy == user.ID || doc.UserID == user.ID

	if !canShare && user.Role == "advisor" {
		var accessLevel string
		db.DB.QueryRow(`
			SELECT access_level FROM advisor_clients
			WHERE advisor_id = ? AND client_id = ? AND status = 'active'
		`, user.ID, doc.UserID).Scan(&accessLevel)
		canShare = accessLevel == "full" || accessLevel == "edit"
	}

	if !canShare {
		http.Error(w, "Cannot share this document", http.StatusForbidden)
		return
	}

	// Parse request
	var req models.ShareDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Permission != "view" && req.Permission != "download" {
		http.Error(w, "Permission must be 'view' or 'download'", http.StatusBadRequest)
		return
	}

	// Calculate expiration
	var expiresAt *time.Time
	if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
		exp := time.Now().Add(time.Duration(*req.ExpiresIn) * time.Hour)
		expiresAt = &exp
	}

	// Create share record (upsert)
	_, err = db.DB.Exec(`
		INSERT INTO document_shares (document_id, shared_with_id, shared_by_id, permission, expires_at)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE permission = VALUES(permission), expires_at = VALUES(expires_at)
	`, docID, req.ShareWithID, user.ID, req.Permission, expiresAt)

	if err != nil {
		http.Error(w, "Failed to share document", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Document shared successfully"})
}

// SaveDocumentFromBytes saves a document directly from bytes (for internal use like report generation)
func SaveDocumentFromBytes(userID int, uploadedBy int, name string, category string, mimeType string, data []byte) (int64, error) {
	// Save to storage (encrypted)
	storagePath, err := storage.DefaultStorage.Save(data, name, true)
	if err != nil {
		return 0, fmt.Errorf("failed to save file: %w", err)
	}

	// Save document record
	result, err := db.DB.Exec(`
		INSERT INTO documents (user_id, uploaded_by, name, original_name, mime_type, size, category, storage_path, encrypted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, TRUE)
	`, userID, uploadedBy, name, name, mimeType, len(data), category, storagePath)

	if err != nil {
		storage.DefaultStorage.Delete(storagePath)
		return 0, fmt.Errorf("failed to save document record: %w", err)
	}

	docID, _ := result.LastInsertId()
	return docID, nil
}
