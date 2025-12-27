package api

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"

	"github.com/finviz/backend/internal/db"
)

func handleCSVImport(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "No file provided")
		return
	}
	defer file.Close()

	// Determine import type from form field or filename
	importType := r.FormValue("type")
	if importType == "" {
		if strings.Contains(strings.ToLower(header.Filename), "asset") {
			importType = "assets"
		} else if strings.Contains(strings.ToLower(header.Filename), "debt") {
			importType = "debts"
		} else {
			respondError(w, http.StatusBadRequest, "Please specify import type: 'assets' or 'debts'")
			return
		}
	}

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse CSV file")
		return
	}

	if len(records) < 2 {
		respondError(w, http.StatusBadRequest, "CSV file must have header row and at least one data row")
		return
	}

	var imported int
	var errors []string

	user := getUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	switch importType {
	case "assets":
		imported, errors = importAssets(records, user.ID)
	case "debts":
		imported, errors = importDebts(records, user.ID)
	case "transactions":
		imported, errors = importTransactions(records, user.ID)
	default:
		respondError(w, http.StatusBadRequest, "Invalid import type. Use 'assets', 'debts', or 'transactions'")
		return
	}

	response := map[string]interface{}{
		"imported": imported,
		"type":     importType,
	}
	if len(errors) > 0 {
		response["errors"] = errors
	}

	respondJSON(w, http.StatusOK, response)
}

// importAssets imports assets from CSV
// Expected columns: name, type_id, current_value, custom_return (optional), custom_volatility (optional)
func importAssets(records [][]string, userID int) (int, []string) {
	var imported int
	var errors []string

	// Find column indices from header
	header := records[0]
	cols := make(map[string]int)
	for i, col := range header {
		cols[strings.ToLower(strings.TrimSpace(col))] = i
	}

	// Required columns
	nameIdx, hasName := cols["name"]
	typeIdx, hasType := cols["type_id"]
	valueIdx, hasValue := cols["current_value"]

	if !hasName || !hasType || !hasValue {
		return 0, []string{"CSV must have columns: name, type_id, current_value"}
	}

	// Optional columns
	returnIdx, hasReturn := cols["custom_return"]
	volIdx, hasVol := cols["custom_volatility"]

	for i, row := range records[1:] {
		rowNum := i + 2 // 1-indexed, skip header

		if len(row) <= nameIdx || len(row) <= typeIdx || len(row) <= valueIdx {
			errors = append(errors, "Row "+strconv.Itoa(rowNum)+": missing required columns")
			continue
		}

		name := strings.TrimSpace(row[nameIdx])
		if name == "" {
			errors = append(errors, "Row "+strconv.Itoa(rowNum)+": name is required")
			continue
		}

		typeID, err := strconv.Atoi(strings.TrimSpace(row[typeIdx]))
		if err != nil {
			errors = append(errors, "Row "+strconv.Itoa(rowNum)+": invalid type_id")
			continue
		}

		value, err := strconv.ParseFloat(strings.TrimSpace(row[valueIdx]), 64)
		if err != nil {
			errors = append(errors, "Row "+strconv.Itoa(rowNum)+": invalid current_value")
			continue
		}

		var customReturn, customVol *float64
		if hasReturn && len(row) > returnIdx && row[returnIdx] != "" {
			r, err := strconv.ParseFloat(strings.TrimSpace(row[returnIdx]), 64)
			if err == nil {
				customReturn = &r
			}
		}
		if hasVol && len(row) > volIdx && row[volIdx] != "" {
			v, err := strconv.ParseFloat(strings.TrimSpace(row[volIdx]), 64)
			if err == nil {
				customVol = &v
			}
		}

		_, err = db.DB.Exec(
			`INSERT INTO assets (user_id, name, type_id, current_value, custom_return, custom_volatility) VALUES (?, ?, ?, ?, ?, ?)`,
			userID, name, typeID, value, customReturn, customVol,
		)
		if err != nil {
			errors = append(errors, "Row "+strconv.Itoa(rowNum)+": "+err.Error())
			continue
		}
		imported++
	}

	return imported, errors
}

// importDebts imports debts from CSV
// Expected columns: name, current_balance, interest_rate (optional), minimum_payment (optional)
func importDebts(records [][]string, userID int) (int, []string) {
	var imported int
	var errors []string

	// Find column indices from header
	header := records[0]
	cols := make(map[string]int)
	for i, col := range header {
		cols[strings.ToLower(strings.TrimSpace(col))] = i
	}

	// Required columns
	nameIdx, hasName := cols["name"]
	balanceIdx, hasBalance := cols["current_balance"]

	if !hasName || !hasBalance {
		return 0, []string{"CSV must have columns: name, current_balance"}
	}

	// Optional columns
	rateIdx, hasRate := cols["interest_rate"]
	paymentIdx, hasPayment := cols["minimum_payment"]

	for i, row := range records[1:] {
		rowNum := i + 2

		if len(row) <= nameIdx || len(row) <= balanceIdx {
			errors = append(errors, "Row "+strconv.Itoa(rowNum)+": missing required columns")
			continue
		}

		name := strings.TrimSpace(row[nameIdx])
		if name == "" {
			errors = append(errors, "Row "+strconv.Itoa(rowNum)+": name is required")
			continue
		}

		balance, err := strconv.ParseFloat(strings.TrimSpace(row[balanceIdx]), 64)
		if err != nil {
			errors = append(errors, "Row "+strconv.Itoa(rowNum)+": invalid current_balance")
			continue
		}

		var rate, payment *float64
		if hasRate && len(row) > rateIdx && row[rateIdx] != "" {
			r, err := strconv.ParseFloat(strings.TrimSpace(row[rateIdx]), 64)
			if err == nil {
				rate = &r
			}
		}
		if hasPayment && len(row) > paymentIdx && row[paymentIdx] != "" {
			p, err := strconv.ParseFloat(strings.TrimSpace(row[paymentIdx]), 64)
			if err == nil {
				payment = &p
			}
		}

		_, err = db.DB.Exec(
			`INSERT INTO debts (user_id, name, current_balance, interest_rate, minimum_payment) VALUES (?, ?, ?, ?, ?)`,
			userID, name, balance, rate, payment,
		)
		if err != nil {
			errors = append(errors, "Row "+strconv.Itoa(rowNum)+": "+err.Error())
			continue
		}
		imported++
	}

	return imported, errors
}

// importTransactions imports transactions from CSV
// Expected columns: date, amount, category (optional), description (optional)
func importTransactions(records [][]string, userID int) (int, []string) {
	var imported int
	var errors []string

	// Find column indices from header
	header := records[0]
	cols := make(map[string]int)
	for i, col := range header {
		cols[strings.ToLower(strings.TrimSpace(col))] = i
	}

	// Required columns
	dateIdx, hasDate := cols["date"]
	amountIdx, hasAmount := cols["amount"]

	if !hasDate || !hasAmount {
		return 0, []string{"CSV must have columns: date, amount"}
	}

	// Optional columns
	categoryIdx, hasCategory := cols["category"]
	descIdx, hasDesc := cols["description"]
	nameIdx, hasName := cols["name"]

	// Income keywords for classification
	incomeKeywords := []string{"income", "salary", "paycheck", "deposit", "dividend", "interest", "refund", "transfer in"}

	for i, row := range records[1:] {
		rowNum := i + 2

		if len(row) <= dateIdx || len(row) <= amountIdx {
			errors = append(errors, "Row "+strconv.Itoa(rowNum)+": missing required columns")
			continue
		}

		dateStr := strings.TrimSpace(row[dateIdx])
		if dateStr == "" {
			errors = append(errors, "Row "+strconv.Itoa(rowNum)+": date is required")
			continue
		}

		amountStr := strings.TrimSpace(row[amountIdx])
		amountStr = strings.ReplaceAll(amountStr, "$", "")
		amountStr = strings.ReplaceAll(amountStr, ",", "")
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			errors = append(errors, "Row "+strconv.Itoa(rowNum)+": invalid amount '"+row[amountIdx]+"'")
			continue
		}

		// Get optional fields
		var category, description, name string

		if hasCategory && len(row) > categoryIdx {
			category = strings.TrimSpace(row[categoryIdx])
		}
		if hasDesc && len(row) > descIdx {
			description = strings.TrimSpace(row[descIdx])
		}
		if hasName && len(row) > nameIdx {
			name = strings.TrimSpace(row[nameIdx])
		}

		// Use description as name if name not provided
		if name == "" && description != "" {
			name = description
		}
		if name == "" {
			name = "Imported Transaction"
		}

		// Determine if income based on category/description keywords
		combined := strings.ToLower(category + " " + description + " " + name)
		isIncome := false
		for _, kw := range incomeKeywords {
			if strings.Contains(combined, kw) {
				isIncome = true
				break
			}
		}

		// For Plaid convention: negative = income, positive = expense
		// If this looks like income but amount is positive, make it negative
		if isIncome && amount > 0 {
			amount = -amount
		}

		// Normalize income category to match Plaid convention (uppercase INCOME)
		if isIncome {
			category = "INCOME"
		}

		_, err = db.DB.Exec(
			`INSERT INTO transactions (user_id, amount, date, name, category, pending) VALUES (?, ?, ?, ?, ?, FALSE)`,
			userID, amount, dateStr, name, category,
		)
		if err != nil {
			errors = append(errors, "Row "+strconv.Itoa(rowNum)+": "+err.Error())
			continue
		}
		imported++
	}

	return imported, errors
}
