package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Storage interface for file operations
type Storage interface {
	// Save stores a file and returns the storage path
	Save(data []byte, filename string, encrypt bool) (string, error)
	// Load retrieves a file by path
	Load(path string, encrypted bool) ([]byte, error)
	// Delete removes a file
	Delete(path string) error
	// GetURL returns a download URL (for S3) or empty for local
	GetURL(path string) string
}

// LocalStorage implements Storage for local filesystem
type LocalStorage struct {
	BasePath      string
	EncryptionKey []byte
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage(basePath string, encryptionKeyStr string) (*LocalStorage, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Derive encryption key from string using SHA-256
	key := sha256.Sum256([]byte(encryptionKeyStr))

	return &LocalStorage{
		BasePath:      basePath,
		EncryptionKey: key[:],
	}, nil
}

// Save stores a file with optional encryption
func (s *LocalStorage) Save(data []byte, filename string, encrypt bool) (string, error) {
	// Generate unique path based on date and random suffix
	now := time.Now()
	dirPath := filepath.Join(s.BasePath, fmt.Sprintf("%d/%02d", now.Year(), now.Month()))

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate unique filename
	randBytes := make([]byte, 8)
	rand.Read(randBytes)
	uniqueName := fmt.Sprintf("%d_%x_%s", now.UnixNano(), randBytes, sanitizeFilename(filename))
	fullPath := filepath.Join(dirPath, uniqueName)

	var dataToWrite []byte
	if encrypt {
		var err error
		dataToWrite, err = s.encrypt(data)
		if err != nil {
			return "", fmt.Errorf("encryption failed: %w", err)
		}
	} else {
		dataToWrite = data
	}

	if err := os.WriteFile(fullPath, dataToWrite, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Return relative path from base
	relPath, _ := filepath.Rel(s.BasePath, fullPath)
	return relPath, nil
}

// Load retrieves a file with optional decryption
func (s *LocalStorage) Load(path string, encrypted bool) ([]byte, error) {
	fullPath := filepath.Join(s.BasePath, path)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if encrypted {
		decrypted, err := s.decrypt(data)
		if err != nil {
			return nil, fmt.Errorf("decryption failed: %w", err)
		}
		return decrypted, nil
	}

	return data, nil
}

// Delete removes a file
func (s *LocalStorage) Delete(path string) error {
	fullPath := filepath.Join(s.BasePath, path)
	return os.Remove(fullPath)
}

// GetURL returns empty for local storage (direct file access)
func (s *LocalStorage) GetURL(path string) string {
	return ""
}

// encrypt data using AES-GCM
func (s *LocalStorage) encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.EncryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decrypt data using AES-GCM
func (s *LocalStorage) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.EncryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// sanitizeFilename removes/replaces characters that are unsafe for filenames
func sanitizeFilename(name string) string {
	result := make([]byte, 0, len(name))
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' {
			result = append(result, byte(c))
		} else if c == ' ' {
			result = append(result, '_')
		}
	}
	if len(result) == 0 {
		return "file"
	}
	// Limit length
	if len(result) > 100 {
		result = result[:100]
	}
	return string(result)
}

// Global storage instance
var DefaultStorage Storage

// InitStorage initializes the default storage
func InitStorage(basePath string, encryptionKey string) error {
	storage, err := NewLocalStorage(basePath, encryptionKey)
	if err != nil {
		return err
	}
	DefaultStorage = storage
	return nil
}
