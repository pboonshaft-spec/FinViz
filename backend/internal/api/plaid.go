package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/finviz/backend/internal/db"
	"github.com/finviz/backend/internal/models"
	"github.com/finviz/backend/internal/plaid"
)

var plaidClient *plaid.Client

func init() {
	plaidClient = plaid.NewClient()
}

// handlePlaidStatus returns whether Plaid is configured
func handlePlaidStatus(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]bool{
		"configured": plaidClient.IsConfigured(),
	})
}

// handleCreateLinkToken creates a Plaid Link token
func handleCreateLinkToken(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	if !plaidClient.IsConfigured() {
		respondError(w, http.StatusServiceUnavailable, "Plaid is not configured")
		return
	}

	userID := strconv.Itoa(user.ID)
	resp, err := plaidClient.CreateLinkToken(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	expiration, _ := time.Parse(time.RFC3339, resp.Expiration)

	respondJSON(w, http.StatusOK, models.LinkTokenResponse{
		LinkToken:  resp.LinkToken,
		Expiration: expiration,
	})
}

// handleExchangeToken exchanges a public token for an access token
func handleExchangeToken(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	if !plaidClient.IsConfigured() {
		respondError(w, http.StatusServiceUnavailable, "Plaid is not configured")
		return
	}

	var req models.ExchangeTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.PublicToken == "" {
		respondError(w, http.StatusBadRequest, "Public token is required")
		return
	}

	// Exchange public token for access token
	exchangeResp, err := plaidClient.ExchangePublicToken(req.PublicToken)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get accounts
	accountsResp, err := plaidClient.GetAccounts(exchangeResp.AccessToken)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get institution name
	var institutionName string
	if accountsResp.Item.InstitutionID != "" {
		instResp, err := plaidClient.GetInstitution(accountsResp.Item.InstitutionID)
		if err == nil {
			institutionName = instResp.Institution.Name
		}
	}

	// Store the item
	result, err := db.DB.Exec(`
		INSERT INTO plaid_items (user_id, item_id, access_token, institution_id, institution_name)
		VALUES (?, ?, ?, ?, ?)
	`, user.ID, exchangeResp.ItemID, exchangeResp.AccessToken, accountsResp.Item.InstitutionID, institutionName)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	plaidItemID, _ := result.LastInsertId()

	// Store accounts
	var plaidAccounts []models.PlaidAccount
	now := time.Now()

	for _, acc := range accountsResp.Accounts {
		_, err := db.DB.Exec(`
			INSERT INTO plaid_accounts (plaid_item_id, user_id, account_id, name, official_name, type, subtype, current_balance, available_balance, credit_limit, iso_currency_code, last_synced_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, plaidItemID, user.ID, acc.AccountID, acc.Name, acc.OfficialName, acc.Type, acc.Subtype,
			acc.Balances.Current, acc.Balances.Available, acc.Balances.Limit, acc.Balances.ISOCurrencyCode, now)
		if err != nil {
			// Log but continue
			fmt.Printf("Error storing account %s: %v\n", acc.AccountID, err)
			continue
		}

		plaidAccounts = append(plaidAccounts, models.PlaidAccount{
			AccountID:        acc.AccountID,
			Name:             acc.Name,
			Type:             acc.Type,
			CurrentBalance:   acc.Balances.Current,
			AvailableBalance: acc.Balances.Available,
			LastSyncedAt:     &now,
		})
	}

	respondJSON(w, http.StatusOK, models.PlaidItemResponse{
		Item: models.PlaidItem{
			ID:              int(plaidItemID),
			ItemID:          exchangeResp.ItemID,
			InstitutionID:   accountsResp.Item.InstitutionID,
			InstitutionName: institutionName,
			Status:          "active",
		},
		Accounts: plaidAccounts,
	})
}

// handleGetPlaidItems returns linked Plaid items
func handleGetPlaidItems(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	rows, err := db.DB.Query(`
		SELECT id, user_id, item_id, institution_id, institution_name, status, created_at, updated_at
		FROM plaid_items
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var items []models.PlaidItem
	for rows.Next() {
		var item models.PlaidItem
		if err := rows.Scan(&item.ID, &item.UserID, &item.ItemID, &item.InstitutionID, &item.InstitutionName, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, item)
	}

	if items == nil {
		items = []models.PlaidItem{}
	}

	respondJSON(w, http.StatusOK, items)
}

// handleGetPlaidAccounts returns Plaid accounts for a user
func handleGetPlaidAccounts(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	rows, err := db.DB.Query(`
		SELECT id, plaid_item_id, user_id, account_id, name, official_name, type, subtype,
		       current_balance, available_balance, credit_limit, iso_currency_code, last_synced_at, created_at, updated_at
		FROM plaid_accounts
		WHERE user_id = ?
		ORDER BY name
	`, user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var accounts []models.PlaidAccount
	for rows.Next() {
		var acc models.PlaidAccount
		if err := rows.Scan(&acc.ID, &acc.PlaidItemID, &acc.UserID, &acc.AccountID, &acc.Name, &acc.OfficialName,
			&acc.Type, &acc.Subtype, &acc.CurrentBalance, &acc.AvailableBalance, &acc.CreditLimit,
			&acc.ISOCurrencyCode, &acc.LastSyncedAt, &acc.CreatedAt, &acc.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		accounts = append(accounts, acc)
	}

	if accounts == nil {
		accounts = []models.PlaidAccount{}
	}

	respondJSON(w, http.StatusOK, accounts)
}

// handleSyncAccounts syncs account balances and creates/updates assets and debts
func handleSyncAccounts(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	if !plaidClient.IsConfigured() {
		respondError(w, http.StatusServiceUnavailable, "Plaid is not configured")
		return
	}

	// Get all plaid items for user
	rows, err := db.DB.Query(`SELECT id, access_token FROM plaid_items WHERE user_id = ? AND status = 'active'`, user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var syncResult models.SyncResponse
	now := time.Now()

	for rows.Next() {
		var itemID int
		var accessToken string
		if err := rows.Scan(&itemID, &accessToken); err != nil {
			continue
		}

		// Get updated account balances
		accountsResp, err := plaidClient.GetAccounts(accessToken)
		if err != nil {
			fmt.Printf("Error getting accounts for item %d: %v\n", itemID, err)
			continue
		}

		for _, acc := range accountsResp.Accounts {
			syncResult.SyncedAccounts++

			// Update plaid_accounts
			_, err := db.DB.Exec(`
				UPDATE plaid_accounts
				SET current_balance = ?, available_balance = ?, credit_limit = ?, last_synced_at = ?
				WHERE account_id = ? AND user_id = ?
			`, acc.Balances.Current, acc.Balances.Available, acc.Balances.Limit, now, acc.AccountID, user.ID)
			if err != nil {
				fmt.Printf("Error updating account %s: %v\n", acc.AccountID, err)
			}

			// Determine if asset or debt based on account type
			isDebt := acc.Type == "credit" || acc.Type == "loan"

			if isDebt {
				// Check if debt exists with this plaid_account_id
				var existingID int
				err := db.DB.QueryRow(`SELECT id FROM debts WHERE plaid_account_id = ? AND user_id = ?`, acc.AccountID, user.ID).Scan(&existingID)

				balance := float64(0)
				if acc.Balances.Current != nil {
					balance = *acc.Balances.Current
				}

				if err == nil {
					// Update existing debt
					_, err = db.DB.Exec(`UPDATE debts SET current_balance = ?, updated_at = NOW() WHERE id = ?`, balance, existingID)
					if err == nil {
						syncResult.UpdatedDebts++
					}
				} else {
					// Create new debt
					_, err = db.DB.Exec(`
						INSERT INTO debts (user_id, name, current_balance, plaid_account_id)
						VALUES (?, ?, ?, ?)
					`, user.ID, acc.Name, balance, acc.AccountID)
					if err == nil {
						syncResult.NewDebts++
					}
				}
			} else {
				// Check if asset exists with this plaid_account_id
				var existingID int
				err := db.DB.QueryRow(`SELECT id FROM assets WHERE plaid_account_id = ? AND user_id = ?`, acc.AccountID, user.ID).Scan(&existingID)

				value := float64(0)
				if acc.Balances.Current != nil {
					value = *acc.Balances.Current
				}

				// Determine asset type based on Plaid account type
				typeID := getAssetTypeIDForPlaidType(acc.Type, acc.Subtype)

				if err == nil {
					// Update existing asset
					_, err = db.DB.Exec(`UPDATE assets SET current_value = ?, updated_at = NOW() WHERE id = ?`, value, existingID)
					if err == nil {
						syncResult.UpdatedAssets++
					}
				} else {
					// Create new asset
					_, err = db.DB.Exec(`
						INSERT INTO assets (user_id, name, type_id, current_value, plaid_account_id)
						VALUES (?, ?, ?, ?, ?)
					`, user.ID, acc.Name, typeID, value, acc.AccountID)
					if err == nil {
						syncResult.NewAssets++
					}
				}
			}
		}
	}

	respondJSON(w, http.StatusOK, syncResult)
}

// getAssetTypeIDForPlaidType maps Plaid account types to our asset types
func getAssetTypeIDForPlaidType(accType, subtype string) int {
	// Default to Cash/Savings (ID 5)
	switch accType {
	case "investment":
		return 1 // Stocks (US)
	case "brokerage":
		return 1 // Stocks (US)
	case "depository":
		return 5 // Cash/Savings
	default:
		return 5 // Cash/Savings
	}
}

// handleDeletePlaidItem removes a Plaid item and optionally associated data
func handleDeletePlaidItem(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	// Check if we should delete associated data
	deleteData := r.URL.Query().Get("delete_data") == "true"

	if deleteData {
		// Get account IDs for this item
		rows, err := db.DB.Query(`SELECT account_id FROM plaid_accounts WHERE plaid_item_id = ? AND user_id = ?`, id, user.ID)
		if err == nil {
			defer rows.Close()
			var accountIDs []string
			for rows.Next() {
				var accID string
				if rows.Scan(&accID) == nil {
					accountIDs = append(accountIDs, accID)
				}
			}

			// Delete associated assets, debts, and transactions
			for _, accID := range accountIDs {
				db.DB.Exec(`DELETE FROM assets WHERE plaid_account_id = ? AND user_id = ?`, accID, user.ID)
				db.DB.Exec(`DELETE FROM debts WHERE plaid_account_id = ? AND user_id = ?`, accID, user.ID)
				db.DB.Exec(`DELETE FROM transactions WHERE plaid_account_id = ? AND user_id = ?`, accID, user.ID)
			}
		}
	}

	// Delete will cascade to plaid_accounts
	result, err := db.DB.Exec(`DELETE FROM plaid_items WHERE id = ? AND user_id = ?`, id, user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		respondError(w, http.StatusNotFound, "Plaid item not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted", "data_deleted": strconv.FormatBool(deleteData)})
}
