package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/finviz/backend/internal/api"
	"github.com/finviz/backend/internal/db"
	"github.com/finviz/backend/internal/storage"
)

func main() {
	// Connect to database
	if err := db.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize document storage
	storagePath := os.Getenv("STORAGE_PATH")
	if storagePath == "" {
		// Default to ./data/documents relative to working directory
		storagePath = filepath.Join(".", "data", "documents")
	}
	encryptionKey := os.Getenv("STORAGE_ENCRYPTION_KEY")
	if encryptionKey == "" {
		encryptionKey = "default-encryption-key-change-in-production"
		log.Println("WARNING: Using default encryption key. Set STORAGE_ENCRYPTION_KEY in production!")
	}
	if err := storage.InitStorage(storagePath, encryptionKey); err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	log.Printf("Document storage initialized at: %s", storagePath)

	// Create router
	router := api.NewRouter()

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
