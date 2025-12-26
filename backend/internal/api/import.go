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

	switch importType {
	case "assets":
		imported, errors = importAssets(records)
	case "debts":
		imported, errors = importDebts(records)
	default:
		respondError(w, http.StatusBadRequest, "Invalid import type. Use 'assets' or 'debts'")
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
func importAssets(records [][]string) (int, []string) {
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
			`INSERT INTO assets (name, type_id, current_value, custom_return, custom_volatility) VALUES (?, ?, ?, ?, ?)`,
			name, typeID, value, customReturn, customVol,
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
func importDebts(records [][]string) (int, []string) {
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
			`INSERT INTO debts (name, current_balance, interest_rate, minimum_payment) VALUES (?, ?, ?, ?)`,
			name, balance, rate, payment,
		)
		if err != nil {
			errors = append(errors, "Row "+strconv.Itoa(rowNum)+": "+err.Error())
			continue
		}
		imported++
	}

	return imported, errors
}
