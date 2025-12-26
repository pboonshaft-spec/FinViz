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
		// Users table for multi-tenancy
		`CREATE TABLE IF NOT EXISTS users (
			id INT PRIMARY KEY AUTO_INCREMENT,
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
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
	}

	for _, migration := range migrations {
		_, err := DB.Exec(migration)
		if err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	// Add plaid_account_id columns if missing (for existing databases)
	alterMigrations := []string{
		`ALTER TABLE assets ADD COLUMN IF NOT EXISTS plaid_account_id VARCHAR(255)`,
		`ALTER TABLE debts ADD COLUMN IF NOT EXISTS plaid_account_id VARCHAR(255)`,
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
