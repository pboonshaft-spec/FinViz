package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func Connect() error {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "3306")
	user := getEnv("DB_USER", "finviz")
	password := getEnv("DB_PASSWORD", "finviz")
	dbname := getEnv("DB_NAME", "finviz")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, password, host, port, dbname)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// Wait for database to be ready
	for i := 0; i < 30; i++ {
		err = DB.Ping()
		if err == nil {
			log.Println("Connected to MySQL database")
			return nil
		}
		log.Printf("Waiting for database... (%d/30)\n", i+1)
		time.Sleep(time.Second)
	}

	return fmt.Errorf("failed to connect to database after 30 attempts: %w", err)
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

func RunMigrations() error {
	migrations := []string{
		// Users table for multi-tenancy with role support
		`CREATE TABLE IF NOT EXISTS users (
			id INT PRIMARY KEY AUTO_INCREMENT,
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			role ENUM('client', 'advisor') NOT NULL DEFAULT 'client',
			created_by_advisor_id INT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS asset_types (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(50) NOT NULL UNIQUE,
			default_return DECIMAL(5,2) NOT NULL,
			default_volatility DECIMAL(5,2) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS assets (
			id INT PRIMARY KEY AUTO_INCREMENT,
			user_id INT NOT NULL,
			name VARCHAR(255) NOT NULL,
			type_id INT NOT NULL,
			current_value DECIMAL(15,2) NOT NULL,
			custom_return DECIMAL(5,2),
			custom_volatility DECIMAL(5,2),
			plaid_account_id VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (type_id) REFERENCES asset_types(id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS debts (
			id INT PRIMARY KEY AUTO_INCREMENT,
			user_id INT NOT NULL,
			name VARCHAR(255) NOT NULL,
			current_balance DECIMAL(15,2) NOT NULL,
			interest_rate DECIMAL(5,2),
			minimum_payment DECIMAL(10,2),
			plaid_account_id VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		// Plaid Items - stores access tokens for linked institutions
		`CREATE TABLE IF NOT EXISTS plaid_items (
			id INT PRIMARY KEY AUTO_INCREMENT,
			user_id INT NOT NULL,
			item_id VARCHAR(255) NOT NULL UNIQUE,
			access_token VARCHAR(255) NOT NULL,
			institution_id VARCHAR(255),
			institution_name VARCHAR(255),
			status VARCHAR(50) DEFAULT 'active',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		// Plaid Accounts - stores synced account info
		`CREATE TABLE IF NOT EXISTS plaid_accounts (
			id INT PRIMARY KEY AUTO_INCREMENT,
			plaid_item_id INT NOT NULL,
			user_id INT NOT NULL,
			account_id VARCHAR(255) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			official_name VARCHAR(255),
			type VARCHAR(50) NOT NULL,
			subtype VARCHAR(50),
			current_balance DECIMAL(15,2),
			available_balance DECIMAL(15,2),
			credit_limit DECIMAL(15,2),
			iso_currency_code VARCHAR(10),
			last_synced_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (plaid_item_id) REFERENCES plaid_items(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		// Transactions - stores synced transaction data from Plaid
		`CREATE TABLE IF NOT EXISTS transactions (
			id INT PRIMARY KEY AUTO_INCREMENT,
			user_id INT NOT NULL,
			plaid_transaction_id VARCHAR(255) UNIQUE,
			plaid_account_id VARCHAR(255),
			account_name VARCHAR(255),
			amount DECIMAL(15,2) NOT NULL,
			date DATE NOT NULL,
			name VARCHAR(255) NOT NULL,
			merchant_name VARCHAR(255),
			category VARCHAR(100),
			subcategory VARCHAR(100),
			pending BOOLEAN DEFAULT FALSE,
			transaction_type VARCHAR(50),
			iso_currency_code VARCHAR(10),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_user_date (user_id, date),
			INDEX idx_user_category (user_id, category)
		)`,
		// Advisor-Client relationships
		`CREATE TABLE IF NOT EXISTS advisor_clients (
			id INT PRIMARY KEY AUTO_INCREMENT,
			advisor_id INT NOT NULL,
			client_id INT NOT NULL,
			status ENUM('pending', 'active', 'revoked') NOT NULL DEFAULT 'pending',
			access_level ENUM('view', 'edit', 'full') NOT NULL DEFAULT 'full',
			invitation_token VARCHAR(255),
			invitation_expires_at TIMESTAMP NULL,
			accepted_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (advisor_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (client_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE KEY unique_relationship (advisor_id, client_id)
		)`,
		// Simulation history - persisted Monte Carlo results
		`CREATE TABLE IF NOT EXISTS simulation_history (
			id INT PRIMARY KEY AUTO_INCREMENT,
			user_id INT NOT NULL,
			run_by_user_id INT NOT NULL,
			name VARCHAR(255),
			notes TEXT,
			params JSON NOT NULL,
			results JSON NOT NULL,
			starting_net_worth DECIMAL(15,2) NOT NULL,
			final_p50 DECIMAL(15,2) NOT NULL,
			success_rate DECIMAL(5,2) NOT NULL,
			time_horizon_years INT NOT NULL,
			is_favorite BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_user_created (user_id, created_at DESC),
			INDEX idx_run_by (run_by_user_id)
		)`,
		// Client invitations for advisor-client linking
		`CREATE TABLE IF NOT EXISTS client_invitations (
			id INT PRIMARY KEY AUTO_INCREMENT,
			advisor_id INT NOT NULL,
			client_email VARCHAR(255) NOT NULL,
			invitation_token VARCHAR(255) NOT NULL,
			status ENUM('pending', 'accepted', 'expired', 'cancelled') NOT NULL DEFAULT 'pending',
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			accepted_at TIMESTAMP NULL,
			FOREIGN KEY (advisor_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE KEY unique_token (invitation_token),
			INDEX idx_email (client_email)
		)`,
		// User public keys for E2E encryption
		`CREATE TABLE IF NOT EXISTS user_public_keys (
			id INT PRIMARY KEY AUTO_INCREMENT,
			user_id INT NOT NULL,
			public_key TEXT NOT NULL,
			key_id VARCHAR(64) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE KEY unique_user_key (user_id, key_id),
			INDEX idx_user (user_id)
		)`,
		// Conversations between advisor and client
		`CREATE TABLE IF NOT EXISTS conversations (
			id INT PRIMARY KEY AUTO_INCREMENT,
			advisor_id INT NOT NULL,
			client_id INT NOT NULL,
			last_message_at TIMESTAMP NULL,
			unread_count_advisor INT DEFAULT 0,
			unread_count_client INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (advisor_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (client_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE KEY unique_conversation (advisor_id, client_id)
		)`,
		// E2E encrypted messages
		`CREATE TABLE IF NOT EXISTS messages (
			id INT PRIMARY KEY AUTO_INCREMENT,
			conversation_id INT NOT NULL,
			sender_id INT NOT NULL,
			encrypted_content TEXT NOT NULL,
			nonce VARCHAR(64) NOT NULL,
			read_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
			FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_conversation_created (conversation_id, created_at)
		)`,
		// Document vault - stores files with encryption at rest
		`CREATE TABLE IF NOT EXISTS documents (
			id INT PRIMARY KEY AUTO_INCREMENT,
			user_id INT NOT NULL,
			uploaded_by INT NOT NULL,
			name VARCHAR(255) NOT NULL,
			original_name VARCHAR(255) NOT NULL,
			mime_type VARCHAR(100) NOT NULL,
			size BIGINT NOT NULL,
			category ENUM('tax_returns', 'statements', 'estate_docs', 'insurance', 'investments', 'reports', 'other') NOT NULL DEFAULT 'other',
			storage_path VARCHAR(500) NOT NULL,
			encrypted BOOLEAN DEFAULT TRUE,
			description TEXT,
			year INT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (uploaded_by) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_user_category (user_id, category),
			INDEX idx_user_deleted (user_id, deleted_at)
		)`,
		// Document sharing permissions
		`CREATE TABLE IF NOT EXISTS document_shares (
			id INT PRIMARY KEY AUTO_INCREMENT,
			document_id INT NOT NULL,
			shared_with_id INT NOT NULL,
			shared_by_id INT NOT NULL,
			permission ENUM('view', 'download') NOT NULL DEFAULT 'view',
			expires_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE,
			FOREIGN KEY (shared_with_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (shared_by_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE KEY unique_share (document_id, shared_with_id)
		)`,
		// Client notes - advisor notes about clients for meeting prep
		`CREATE TABLE IF NOT EXISTS client_notes (
			id INT PRIMARY KEY AUTO_INCREMENT,
			advisor_id INT NOT NULL,
			client_id INT NOT NULL,
			note TEXT NOT NULL,
			category ENUM('general', 'meeting', 'goal', 'concern', 'action_item', 'personal') NOT NULL DEFAULT 'general',
			is_pinned BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (advisor_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (client_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_advisor_client (advisor_id, client_id),
			INDEX idx_client_category (client_id, category)
		)`,
		// Client goals - visible to both advisors and clients
		`CREATE TABLE IF NOT EXISTS client_goals (
			id INT PRIMARY KEY AUTO_INCREMENT,
			advisor_id INT NOT NULL,
			client_id INT NOT NULL,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			category ENUM('retirement', 'savings', 'debt', 'investment', 'education', 'emergency', 'major_purchase', 'other') NOT NULL DEFAULT 'other',
			status ENUM('pending', 'in_progress', 'completed', 'on_hold') NOT NULL DEFAULT 'pending',
			priority ENUM('low', 'medium', 'high') NOT NULL DEFAULT 'medium',
			target_amount DECIMAL(15, 2) NULL,
			current_amount DECIMAL(15, 2) NULL DEFAULT 0,
			target_date DATE NULL,
			completed_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (advisor_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (client_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_client_goals (client_id),
			INDEX idx_advisor_client_goals (advisor_id, client_id),
			INDEX idx_status (status)
		)`,
	}

	for _, migration := range migrations {
		_, err := DB.Exec(migration)
		if err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	// Add columns if missing (for existing databases)
	alterMigrations := []string{
		`ALTER TABLE assets ADD COLUMN IF NOT EXISTS plaid_account_id VARCHAR(255)`,
		`ALTER TABLE debts ADD COLUMN IF NOT EXISTS plaid_account_id VARCHAR(255)`,
		// Add role support to users table for existing databases
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS role ENUM('client', 'advisor') NOT NULL DEFAULT 'client'`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS created_by_advisor_id INT NULL`,
	}
	for _, m := range alterMigrations {
		DB.Exec(m) // Ignore errors - column may already exist
	}

	// Seed default asset types
	seedAssetTypes()

	log.Println("Database migrations completed")
	return nil
}

func seedAssetTypes() {
	defaults := []struct {
		name       string
		returnRate float64
		volatility float64
	}{
		{"Stocks (US)", 10.0, 15.0},
		{"Stocks (Intl)", 8.0, 18.0},
		{"Bonds", 4.0, 5.0},
		{"Real Estate", 7.0, 12.0},
		{"Cash/Savings", 2.0, 0.5},
		{"Crypto", 15.0, 60.0},
	}

	for _, d := range defaults {
		_, _ = DB.Exec(
			`INSERT IGNORE INTO asset_types (name, default_return, default_volatility) VALUES (?, ?, ?)`,
			d.name, d.returnRate, d.volatility,
		)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
