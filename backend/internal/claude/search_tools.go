package claude

import (
	"encoding/json"
	"fmt"
)

// getCurrentRates returns current financial rates and limits
func (e *ToolExecutor) getCurrentRates(input map[string]interface{}) (string, error) {
	rateType, ok := input["rate_type"].(string)
	if !ok || rateType == "" {
		return "", fmt.Errorf("rate_type is required")
	}

	// 2024/2025 tax year data (update annually)
	data := map[string]interface{}{}

	switch rateType {
	case "federal_funds":
		data = map[string]interface{}{
			"rate_type":    "Federal Funds Rate",
			"current_rate": "5.25% - 5.50%",
			"as_of":        "December 2024",
			"notes":        "Target range set by the Federal Reserve. Check federalreserve.gov for the most current rate.",
		}

	case "tax_brackets":
		data = map[string]interface{}{
			"rate_type": "2024 Federal Income Tax Brackets (Single)",
			"brackets": []map[string]interface{}{
				{"rate": "10%", "income_range": "$0 - $11,600"},
				{"rate": "12%", "income_range": "$11,601 - $47,150"},
				{"rate": "22%", "income_range": "$47,151 - $100,525"},
				{"rate": "24%", "income_range": "$100,526 - $191,950"},
				{"rate": "32%", "income_range": "$191,951 - $243,725"},
				{"rate": "35%", "income_range": "$243,726 - $609,350"},
				{"rate": "37%", "income_range": "Over $609,350"},
			},
			"notes": "Married filing jointly brackets are approximately double. See IRS.gov for complete tables.",
		}

	case "401k_limits":
		data = map[string]interface{}{
			"rate_type": "2024 401(k) Contribution Limits",
			"limits": map[string]interface{}{
				"employee_contribution":   "$23,000",
				"catch_up_50_plus":        "$7,500",
				"total_with_catch_up":     "$30,500",
				"total_annual_limit":      "$69,000 (including employer contributions)",
				"total_with_catch_up_all": "$76,500",
			},
			"notes": "Catch-up contributions available for those 50 and older.",
		}

	case "ira_limits":
		data = map[string]interface{}{
			"rate_type": "2024 IRA Contribution Limits",
			"limits": map[string]interface{}{
				"contribution_limit":   "$7,000",
				"catch_up_50_plus":     "$1,000",
				"total_with_catch_up":  "$8,000",
				"roth_income_phaseout": "Single: $146,000 - $161,000 MAGI",
			},
			"notes": "Traditional IRA deductibility depends on income and workplace retirement plan coverage.",
		}

	case "hsa_limits":
		data = map[string]interface{}{
			"rate_type": "2024 HSA Contribution Limits",
			"limits": map[string]interface{}{
				"individual_coverage": "$4,150",
				"family_coverage":     "$8,300",
				"catch_up_55_plus":    "$1,000",
			},
			"requirements": "Must be enrolled in a High Deductible Health Plan (HDHP).",
		}

	case "social_security":
		data = map[string]interface{}{
			"rate_type": "2024 Social Security",
			"data": map[string]interface{}{
				"wage_base_limit":      "$168,600",
				"tax_rate_employee":    "6.2%",
				"tax_rate_self_employ": "12.4%",
				"full_retirement_age":  "67 (for those born 1960 or later)",
				"early_retirement":     "62 (with reduced benefits)",
				"delayed_retirement":   "Up to 8% increase per year until age 70",
			},
		}

	default:
		return "", fmt.Errorf("unknown rate_type: %s", rateType)
	}

	jsonBytes, _ := json.MarshalIndent(data, "", "  ")
	return string(jsonBytes), nil
}

// createChart returns chart data in a format the frontend can render
func (e *ToolExecutor) createChart(input map[string]interface{}) (string, error) {
	// Validate required fields
	chartType, _ := input["chart_type"].(string)
	title, _ := input["title"].(string)
	data, _ := input["data"].([]interface{})

	if chartType == "" || title == "" || len(data) == 0 {
		return "", fmt.Errorf("chart_type, title, and data are required")
	}

	// Return the chart specification as-is for frontend rendering
	output := map[string]interface{}{
		"type":       "chart",
		"chart_type": chartType,
		"title":      title,
		"data":       data,
	}

	if colors, ok := input["colors"].([]interface{}); ok {
		output["colors"] = colors
	}

	jsonBytes, _ := json.MarshalIndent(output, "", "  ")
	return string(jsonBytes), nil
}

// createTable returns table data in a format the frontend can render
func (e *ToolExecutor) createTable(input map[string]interface{}) (string, error) {
	title, _ := input["title"].(string)
	headers, _ := input["headers"].([]interface{})
	rows, _ := input["rows"].([]interface{})

	if title == "" || len(headers) == 0 || len(rows) == 0 {
		return "", fmt.Errorf("title, headers, and rows are required")
	}

	output := map[string]interface{}{
		"type":    "table",
		"title":   title,
		"headers": headers,
		"rows":    rows,
	}

	jsonBytes, _ := json.MarshalIndent(output, "", "  ")
	return string(jsonBytes), nil
}

// createMetricCard returns metric card data for frontend rendering
func (e *ToolExecutor) createMetricCard(input map[string]interface{}) (string, error) {
	label, _ := input["label"].(string)
	value, _ := input["value"].(string)

	if label == "" || value == "" {
		return "", fmt.Errorf("label and value are required")
	}

	output := map[string]interface{}{
		"type":  "metric_card",
		"label": label,
		"value": value,
	}

	if change, ok := input["change"].(string); ok {
		output["change"] = change
	}
	if trend, ok := input["trend"].(string); ok {
		output["trend"] = trend
	}

	jsonBytes, _ := json.MarshalIndent(output, "", "  ")
	return string(jsonBytes), nil
}
