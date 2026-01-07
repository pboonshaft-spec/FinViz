package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/finviz/backend/internal/db"
	"github.com/finviz/backend/internal/models"
)

// handleGetTransactions returns transactions for the authenticated user
func handleGetTransactions(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Use effective user ID for client context support
	userID := getEffectiveUserID(r)

	// Get query params for filtering
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	category := r.URL.Query().Get("category")

	// Default to last 30 days if no dates provided
	if startDate == "" {
		startDate = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	query := `
		SELECT id, user_id, plaid_transaction_id, plaid_account_id, account_name, amount, date,
		       name, merchant_name, category, subcategory, pending, transaction_type, iso_currency_code,
		       created_at, updated_at
		FROM transactions
		WHERE user_id = ? AND date >= ? AND date <= ?
	`
	args := []interface{}{userID, startDate, endDate}

	if category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}

	query += " ORDER BY date DESC, id DESC"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		var plaidTxnID, plaidAcctID, accountName, merchantName, category, subcategory, txnType, currency sql.NullString

		if err := rows.Scan(
			&t.ID, &t.UserID, &plaidTxnID, &plaidAcctID, &accountName, &t.Amount, &t.Date,
			&t.Name, &merchantName, &category, &subcategory, &t.Pending, &txnType, &currency,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if plaidTxnID.Valid {
			t.PlaidTransactionID = &plaidTxnID.String
		}
		if plaidAcctID.Valid {
			t.PlaidAccountID = &plaidAcctID.String
		}
		if accountName.Valid {
			t.AccountName = &accountName.String
		}
		if merchantName.Valid {
			t.MerchantName = &merchantName.String
		}
		if category.Valid {
			t.Category = &category.String
		}
		if subcategory.Valid {
			t.Subcategory = &subcategory.String
		}
		if txnType.Valid {
			t.TransactionType = &txnType.String
		}
		if currency.Valid {
			t.ISOCurrencyCode = &currency.String
		}

		transactions = append(transactions, t)
	}

	if transactions == nil {
		transactions = []models.Transaction{}
	}

	respondJSON(w, http.StatusOK, transactions)
}

// handleGetTransactionSummary returns aggregated transaction data
func handleGetTransactionSummary(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Use effective user ID for client context support
	userID := getEffectiveUserID(r)

	// Get query params for filtering
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	// Default to last 30 days
	if startDate == "" {
		startDate = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	var summary models.TransactionSummary

	// Get total income (negative amounts in Plaid = money coming in)
	// Also include INCOME and TRANSFER_IN categories regardless of amount sign (in case of data issues)
	err := db.DB.QueryRow(`
		SELECT COALESCE(SUM(ABS(amount)), 0) FROM transactions
		WHERE user_id = ? AND date >= ? AND date <= ? AND pending = FALSE
		AND (
			amount < 0
			OR category IN ('INCOME', 'INCOME_WAGES', 'INCOME_DIVIDENDS', 'INCOME_INTEREST', 'TRANSFER_IN')
			OR subcategory LIKE 'INCOME%'
			OR subcategory LIKE 'TRANSFER_IN%'
		)
	`, userID, startDate, endDate).Scan(&summary.TotalIncome)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get total expenses (positive amounts in Plaid = money going out)
	// Exclude income categories that might be miscategorized
	err = db.DB.QueryRow(`
		SELECT COALESCE(SUM(amount), 0) FROM transactions
		WHERE user_id = ? AND date >= ? AND date <= ? AND amount > 0 AND pending = FALSE
		AND category NOT IN ('INCOME', 'INCOME_WAGES', 'INCOME_DIVIDENDS', 'INCOME_INTEREST', 'TRANSFER_IN')
		AND (subcategory IS NULL OR (subcategory NOT LIKE 'INCOME%' AND subcategory NOT LIKE 'TRANSFER_IN%'))
	`, userID, startDate, endDate).Scan(&summary.TotalExpenses)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	summary.NetCashFlow = summary.TotalIncome - summary.TotalExpenses

	// Get spending by category (only expenses, excluding income categories)
	catRows, err := db.DB.Query(`
		SELECT COALESCE(category, 'Uncategorized') as cat, SUM(amount) as total, COUNT(*) as cnt
		FROM transactions
		WHERE user_id = ? AND date >= ? AND date <= ? AND amount > 0 AND pending = FALSE
		AND category NOT IN ('INCOME', 'INCOME_WAGES', 'INCOME_DIVIDENDS', 'INCOME_INTEREST', 'TRANSFER_IN')
		AND (subcategory IS NULL OR (subcategory NOT LIKE 'INCOME%' AND subcategory NOT LIKE 'TRANSFER_IN%'))
		GROUP BY category
		ORDER BY total DESC
	`, userID, startDate, endDate)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer catRows.Close()

	for catRows.Next() {
		var cs models.CategorySummary
		if err := catRows.Scan(&cs.Category, &cs.Amount, &cs.Count); err != nil {
			continue
		}
		summary.ByCategory = append(summary.ByCategory, cs)
	}

	if summary.ByCategory == nil {
		summary.ByCategory = []models.CategorySummary{}
	}

	// Get monthly breakdown with proper income/expense classification
	monthRows, err := db.DB.Query(`
		SELECT
			DATE_FORMAT(date, '%Y-%m') as month,
			COALESCE(SUM(CASE
				WHEN amount < 0 THEN ABS(amount)
				WHEN category IN ('INCOME', 'INCOME_WAGES', 'INCOME_DIVIDENDS', 'INCOME_INTEREST', 'TRANSFER_IN') THEN ABS(amount)
				WHEN subcategory LIKE 'INCOME%' OR subcategory LIKE 'TRANSFER_IN%' THEN ABS(amount)
				ELSE 0
			END), 0) as income,
			COALESCE(SUM(CASE
				WHEN amount > 0
				AND category NOT IN ('INCOME', 'INCOME_WAGES', 'INCOME_DIVIDENDS', 'INCOME_INTEREST', 'TRANSFER_IN')
				AND (subcategory IS NULL OR (subcategory NOT LIKE 'INCOME%' AND subcategory NOT LIKE 'TRANSFER_IN%'))
				THEN amount
				ELSE 0
			END), 0) as expenses
		FROM transactions
		WHERE user_id = ? AND date >= ? AND date <= ? AND pending = FALSE
		GROUP BY DATE_FORMAT(date, '%Y-%m')
		ORDER BY month
	`, userID, startDate, endDate)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer monthRows.Close()

	for monthRows.Next() {
		var ms models.MonthSummary
		if err := monthRows.Scan(&ms.Month, &ms.Income, &ms.Expenses); err != nil {
			continue
		}
		ms.Net = ms.Income - ms.Expenses
		summary.ByMonth = append(summary.ByMonth, ms)
	}

	if summary.ByMonth == nil {
		summary.ByMonth = []models.MonthSummary{}
	}

	respondJSON(w, http.StatusOK, summary)
}

// handleSyncTransactions syncs transactions from Plaid
func handleSyncTransactions(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	if !plaidClient.IsConfigured() {
		respondError(w, http.StatusServiceUnavailable, "Plaid is not configured")
		return
	}

	// Get date range from request or default to last 30 days
	var req struct {
		StartDate string `json:"startDate"`
		EndDate   string `json:"endDate"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	startDate := req.StartDate
	endDate := req.EndDate

	if startDate == "" {
		startDate = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	// Get all plaid items for user
	rows, err := db.DB.Query(`SELECT id, access_token FROM plaid_items WHERE user_id = ? AND status = 'active'`, user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var result models.SyncTransactionsResponse

	// Build account ID to name map
	accountMap := make(map[string]string)
	acctRows, _ := db.DB.Query(`SELECT account_id, name FROM plaid_accounts WHERE user_id = ?`, user.ID)
	if acctRows != nil {
		defer acctRows.Close()
		for acctRows.Next() {
			var accID, name string
			if acctRows.Scan(&accID, &name) == nil {
				accountMap[accID] = name
			}
		}
	}

	for rows.Next() {
		var itemID int
		var accessToken string
		if err := rows.Scan(&itemID, &accessToken); err != nil {
			continue
		}

		// Get transactions from Plaid
		txnResp, err := plaidClient.GetTransactions(accessToken, startDate, endDate)
		if err != nil {
			fmt.Printf("Error getting transactions for item %d: %v\n", itemID, err)
			continue
		}

		// Update account map with any new accounts
		for _, acc := range txnResp.Accounts {
			accountMap[acc.AccountID] = acc.Name
		}

		// Process transactions
		for _, txn := range txnResp.Transactions {
			// Determine category
			var category, subcategory string
			if txn.PersonalFinanceCat != nil {
				category = txn.PersonalFinanceCat.Primary
				subcategory = txn.PersonalFinanceCat.Detailed
			} else if len(txn.Category) > 0 {
				category = txn.Category[0]
				if len(txn.Category) > 1 {
					subcategory = txn.Category[1]
				}
			}

			accountName := accountMap[txn.AccountID]

			// Try to insert, update if exists
			res, err := db.DB.Exec(`
				INSERT INTO transactions (user_id, plaid_transaction_id, plaid_account_id, account_name, amount, date, name, merchant_name, category, subcategory, pending, transaction_type, iso_currency_code)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				ON DUPLICATE KEY UPDATE
					amount = VALUES(amount),
					name = VALUES(name),
					merchant_name = VALUES(merchant_name),
					category = VALUES(category),
					subcategory = VALUES(subcategory),
					pending = VALUES(pending),
					updated_at = NOW()
			`, user.ID, txn.TransactionID, txn.AccountID, accountName, txn.Amount, txn.Date, txn.Name,
				txn.MerchantName, category, subcategory, txn.Pending, txn.TransactionType, txn.ISOCurrencyCode)

			if err != nil {
				fmt.Printf("Error inserting transaction %s: %v\n", txn.TransactionID, err)
				continue
			}

			rowsAffected, _ := res.RowsAffected()
			if rowsAffected == 1 {
				result.NewTransactions++
			} else {
				result.UpdatedTransactions++
			}
		}
	}

	respondJSON(w, http.StatusOK, result)
}

// handleGetTransactionDebug returns transaction statistics for debugging
func handleGetTransactionDebug(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Use effective user ID for client context support
	userID := getEffectiveUserID(r)

	type DebugStats struct {
		TotalCount        int     `json:"totalCount"`
		IncomeCount       int     `json:"incomeCount"`
		ExpenseCount      int     `json:"expenseCount"`
		PendingCount      int     `json:"pendingCount"`
		IncomeTotal       float64 `json:"incomeTotal"`
		ExpenseTotal      float64 `json:"expenseTotal"`
		PendingIncomeTotal float64 `json:"pendingIncomeTotal"`
		Categories        map[string]int `json:"categories"`
		SampleIncome      []map[string]interface{} `json:"sampleIncome"`
	}

	var stats DebugStats
	stats.Categories = make(map[string]int)

	// Get all transactions (no date filter)
	rows, err := db.DB.Query(`
		SELECT amount, pending, COALESCE(category, 'NULL') as cat, name, date
		FROM transactions WHERE user_id = ?
		ORDER BY date DESC
	`, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var amount float64
		var pending bool
		var category, name, date string
		if rows.Scan(&amount, &pending, &category, &name, &date) != nil {
			continue
		}

		stats.TotalCount++
		stats.Categories[category]++

		if pending {
			stats.PendingCount++
			if amount < 0 {
				stats.PendingIncomeTotal += -amount
			}
		}

		if amount < 0 {
			stats.IncomeCount++
			stats.IncomeTotal += -amount
			// Sample first 5 income transactions
			if len(stats.SampleIncome) < 5 {
				stats.SampleIncome = append(stats.SampleIncome, map[string]interface{}{
					"name":     name,
					"amount":   amount,
					"category": category,
					"date":     date,
					"pending":  pending,
				})
			}
		} else {
			stats.ExpenseCount++
			stats.ExpenseTotal += amount
		}
	}

	respondJSON(w, http.StatusOK, stats)
}

// handleGetCategories returns distinct categories from user's transactions
func handleGetCategories(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Use effective user ID for client context support
	userID := getEffectiveUserID(r)

	rows, err := db.DB.Query(`
		SELECT DISTINCT COALESCE(category, 'Uncategorized') as cat
		FROM transactions
		WHERE user_id = ?
		ORDER BY cat
	`, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var cat string
		if rows.Scan(&cat) == nil {
			categories = append(categories, cat)
		}
	}

	if categories == nil {
		categories = []string{}
	}

	respondJSON(w, http.StatusOK, categories)
}
