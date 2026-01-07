package taxparser

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ledongthuc/pdf"
)

// TaxDocumentType identifies the type of tax document
type TaxDocumentType string

const (
	DocType1040    TaxDocumentType = "form_1040"
	DocTypeW2      TaxDocumentType = "form_w2"
	DocType1099    TaxDocumentType = "form_1099"
	DocTypeUnknown TaxDocumentType = "unknown"
)

// ExtractedTaxData contains parsed data from tax documents
type ExtractedTaxData struct {
	DocumentType TaxDocumentType `json:"document_type"`
	TaxYear      int             `json:"tax_year,omitempty"`
	FilingStatus string          `json:"filing_status,omitempty"`

	// Form 1040 fields
	TotalIncome        *float64 `json:"total_income,omitempty"`
	AGI                *float64 `json:"agi,omitempty"`
	TaxableIncome      *float64 `json:"taxable_income,omitempty"`
	TotalTax           *float64 `json:"total_tax,omitempty"`
	TotalPayments      *float64 `json:"total_payments,omitempty"`
	RefundAmount       *float64 `json:"refund_amount,omitempty"`
	AmountOwed         *float64 `json:"amount_owed,omitempty"`
	StandardDeduction  *float64 `json:"standard_deduction,omitempty"`
	ItemizedDeductions *float64 `json:"itemized_deductions,omitempty"`

	// W-2 fields
	Employer            string   `json:"employer,omitempty"`
	WagesTips           *float64 `json:"wages_tips,omitempty"`
	FederalWithheld     *float64 `json:"federal_withheld,omitempty"`
	SocialSecurityWages *float64 `json:"social_security_wages,omitempty"`
	SocialSecurityTax   *float64 `json:"social_security_tax,omitempty"`
	MedicareWages       *float64 `json:"medicare_wages,omitempty"`
	MedicareTax         *float64 `json:"medicare_tax,omitempty"`

	// 1099 fields
	Payer       string   `json:"payer,omitempty"`
	IncomeType  string   `json:"income_type,omitempty"` // DIV, INT, MISC, NEC
	GrossIncome *float64 `json:"gross_income,omitempty"`

	// Parsing metadata
	RawText     string   `json:"-"` // For debugging, not returned
	Confidence  float64  `json:"confidence"`
	ParseErrors []string `json:"parse_errors,omitempty"`
}

// ParsePDFContent extracts and parses tax data from PDF bytes
func ParsePDFContent(pdfBytes []byte) (*ExtractedTaxData, error) {
	// Extract text from PDF
	reader := bytes.NewReader(pdfBytes)
	pdfReader, err := pdf.NewReader(reader, int64(len(pdfBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF: %w", err)
	}

	var textBuilder strings.Builder
	for pageNum := 1; pageNum <= pdfReader.NumPage(); pageNum++ {
		page := pdfReader.Page(pageNum)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		textBuilder.WriteString(text)
		textBuilder.WriteString("\n---PAGE BREAK---\n")
	}

	rawText := textBuilder.String()

	// Detect document type
	docType := detectDocumentType(rawText)

	// Parse based on document type
	var data *ExtractedTaxData
	switch docType {
	case DocType1040:
		data = parse1040(rawText)
	case DocTypeW2:
		data = parseW2(rawText)
	case DocType1099:
		data = parse1099(rawText)
	default:
		data = &ExtractedTaxData{
			DocumentType: DocTypeUnknown,
			ParseErrors:  []string{"Could not identify document type"},
			Confidence:   0.0,
		}
	}

	data.RawText = rawText
	return data, nil
}

func detectDocumentType(text string) TaxDocumentType {
	textUpper := strings.ToUpper(text)

	if strings.Contains(textUpper, "FORM 1040") ||
		strings.Contains(textUpper, "U.S. INDIVIDUAL INCOME TAX RETURN") {
		return DocType1040
	}
	if strings.Contains(textUpper, "FORM W-2") ||
		strings.Contains(textUpper, "WAGE AND TAX STATEMENT") {
		return DocTypeW2
	}
	if strings.Contains(textUpper, "FORM 1099") {
		return DocType1099
	}
	return DocTypeUnknown
}

// parse1040 extracts data from Form 1040
func parse1040(text string) *ExtractedTaxData {
	data := &ExtractedTaxData{
		DocumentType: DocType1040,
		Confidence:   0.5, // Start with 50%, increase as we find fields
	}

	// Extract tax year (look for 4-digit year near "Form 1040")
	yearRegex := regexp.MustCompile(`(?i)(?:form\s*1040|tax\s*year)[^\d]*(\d{4})`)
	if match := yearRegex.FindStringSubmatch(text); len(match) > 1 {
		if year, err := strconv.Atoi(match[1]); err == nil && year >= 2018 && year <= 2030 {
			data.TaxYear = year
			data.Confidence += 0.1
		}
	}

	// Extract filing status
	filingStatusPatterns := map[string]string{
		"single":                      `(?i)(?:\[x\]|☑|✓)\s*single`,
		"married_filing_jointly":     `(?i)(?:\[x\]|☑|✓)\s*married\s*filing\s*jointly`,
		"married_filing_separately":  `(?i)(?:\[x\]|☑|✓)\s*married\s*filing\s*separately`,
		"head_of_household":          `(?i)(?:\[x\]|☑|✓)\s*head\s*of\s*household`,
		"qualifying_widow":           `(?i)(?:\[x\]|☑|✓)\s*qualifying\s*(?:widow|surviving\s*spouse)`,
	}
	for status, pattern := range filingStatusPatterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			data.FilingStatus = status
			data.Confidence += 0.1
			break
		}
	}

	// Try to infer from text if no checkbox found
	if data.FilingStatus == "" {
		textLower := strings.ToLower(text)
		if strings.Contains(textLower, "married filing jointly") {
			data.FilingStatus = "married_filing_jointly"
		} else if strings.Contains(textLower, "single") && !strings.Contains(textLower, "married") {
			data.FilingStatus = "single"
		}
	}

	// Extract key monetary values using line number patterns
	// Line 9: Total income
	data.TotalIncome = extractLineValue(text, []string{
		`(?i)line\s*9[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
		`(?i)total\s*income[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
	})
	if data.TotalIncome != nil {
		data.Confidence += 0.1
	}

	// Line 11: Adjusted Gross Income
	data.AGI = extractLineValue(text, []string{
		`(?i)line\s*11[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
		`(?i)adjusted\s*gross\s*income[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
		`(?i)\bagi\b[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
	})
	if data.AGI != nil {
		data.Confidence += 0.15
	}

	// Line 12: Standard/Itemized deduction
	deduction := extractLineValue(text, []string{
		`(?i)line\s*12[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
		`(?i)standard\s*deduction[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
	})
	if deduction != nil {
		// Determine if standard or itemized based on common amounts
		standardAmounts := map[float64]bool{
			13850: true, 14600: true, // Single 2023, 2024
			27700: true, 29200: true, // MFJ 2023, 2024
			20800: true, 21900: true, // HOH 2023, 2024
		}
		if standardAmounts[*deduction] {
			data.StandardDeduction = deduction
		} else {
			// Assume itemized if not a standard amount
			data.ItemizedDeductions = deduction
		}
		data.Confidence += 0.05
	}

	// Line 15: Taxable income
	data.TaxableIncome = extractLineValue(text, []string{
		`(?i)line\s*15[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
		`(?i)taxable\s*income[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
	})
	if data.TaxableIncome != nil {
		data.Confidence += 0.1
	}

	// Line 24: Total tax
	data.TotalTax = extractLineValue(text, []string{
		`(?i)line\s*24[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
		`(?i)total\s*tax[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
	})
	if data.TotalTax != nil {
		data.Confidence += 0.1
	}

	// Line 33: Total payments
	data.TotalPayments = extractLineValue(text, []string{
		`(?i)line\s*33[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
		`(?i)total\s*payments[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
	})

	// Line 34: Refund
	data.RefundAmount = extractLineValue(text, []string{
		`(?i)line\s*34[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
		`(?i)refund[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
		`(?i)overpaid[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
	})

	// Line 37: Amount owed
	data.AmountOwed = extractLineValue(text, []string{
		`(?i)line\s*37[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
		`(?i)amount\s*(?:you\s*)?owe[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
	})

	// Cap confidence at 1.0
	if data.Confidence > 1.0 {
		data.Confidence = 1.0
	}

	return data
}

// parseW2 extracts data from Form W-2
func parseW2(text string) *ExtractedTaxData {
	data := &ExtractedTaxData{
		DocumentType: DocTypeW2,
		Confidence:   0.5,
	}

	// Extract tax year
	yearRegex := regexp.MustCompile(`(?i)(?:w-2|tax\s*year|wages)[^\d]*(\d{4})`)
	if match := yearRegex.FindStringSubmatch(text); len(match) > 1 {
		if year, err := strconv.Atoi(match[1]); err == nil && year >= 2018 && year <= 2030 {
			data.TaxYear = year
			data.Confidence += 0.1
		}
	}

	// Box c: Employer name/address (look for company-like text)
	employerRegex := regexp.MustCompile(`(?i)(?:employer|box\s*c)[^\n]*\n([A-Za-z][\w\s&.,'-]+(?:inc|llc|corp|company|ltd)?\.?)`)
	if match := employerRegex.FindStringSubmatch(text); len(match) > 1 {
		data.Employer = strings.TrimSpace(match[1])
		data.Confidence += 0.1
	}

	// Box 1: Wages, tips, other compensation
	data.WagesTips = extractLineValue(text, []string{
		`(?i)box\s*1[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
		`(?i)wages,?\s*tips[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
	})
	if data.WagesTips != nil {
		data.Confidence += 0.15
	}

	// Box 2: Federal income tax withheld
	data.FederalWithheld = extractLineValue(text, []string{
		`(?i)box\s*2[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
		`(?i)federal\s*(?:income\s*)?tax\s*withheld[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
	})
	if data.FederalWithheld != nil {
		data.Confidence += 0.1
	}

	// Box 3: Social security wages
	data.SocialSecurityWages = extractLineValue(text, []string{
		`(?i)box\s*3[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
		`(?i)social\s*security\s*wages[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
	})

	// Box 4: Social security tax withheld
	data.SocialSecurityTax = extractLineValue(text, []string{
		`(?i)box\s*4[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
		`(?i)social\s*security\s*tax[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
	})

	// Box 5: Medicare wages
	data.MedicareWages = extractLineValue(text, []string{
		`(?i)box\s*5[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
		`(?i)medicare\s*wages[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
	})

	// Box 6: Medicare tax withheld
	data.MedicareTax = extractLineValue(text, []string{
		`(?i)box\s*6[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
		`(?i)medicare\s*tax[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
	})

	if data.Confidence > 1.0 {
		data.Confidence = 1.0
	}

	return data
}

// parse1099 extracts data from Form 1099 variants
func parse1099(text string) *ExtractedTaxData {
	data := &ExtractedTaxData{
		DocumentType: DocType1099,
		Confidence:   0.5,
	}

	textUpper := strings.ToUpper(text)

	// Determine 1099 type
	if strings.Contains(textUpper, "1099-DIV") {
		data.IncomeType = "DIV"
	} else if strings.Contains(textUpper, "1099-INT") {
		data.IncomeType = "INT"
	} else if strings.Contains(textUpper, "1099-MISC") {
		data.IncomeType = "MISC"
	} else if strings.Contains(textUpper, "1099-NEC") {
		data.IncomeType = "NEC"
	} else if strings.Contains(textUpper, "1099-B") {
		data.IncomeType = "B"
	} else if strings.Contains(textUpper, "1099-R") {
		data.IncomeType = "R"
	}

	if data.IncomeType != "" {
		data.Confidence += 0.1
	}

	// Extract tax year
	yearRegex := regexp.MustCompile(`(?i)(?:1099|tax\s*year)[^\d]*(\d{4})`)
	if match := yearRegex.FindStringSubmatch(text); len(match) > 1 {
		if year, err := strconv.Atoi(match[1]); err == nil && year >= 2018 && year <= 2030 {
			data.TaxYear = year
			data.Confidence += 0.1
		}
	}

	// Extract payer name
	payerRegex := regexp.MustCompile(`(?i)(?:payer|payor)[^\n]*\n([A-Za-z][\w\s&.,'-]+)`)
	if match := payerRegex.FindStringSubmatch(text); len(match) > 1 {
		data.Payer = strings.TrimSpace(match[1])
		data.Confidence += 0.1
	}

	// Extract gross income based on form type
	switch data.IncomeType {
	case "DIV":
		// Box 1a: Total ordinary dividends
		data.GrossIncome = extractLineValue(text, []string{
			`(?i)(?:box\s*)?1a[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
			`(?i)total\s*ordinary\s*dividends[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
		})
	case "INT":
		// Box 1: Interest income
		data.GrossIncome = extractLineValue(text, []string{
			`(?i)box\s*1[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
			`(?i)interest\s*income[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
		})
	case "NEC", "MISC":
		// Box 1: Nonemployee compensation
		data.GrossIncome = extractLineValue(text, []string{
			`(?i)box\s*1[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
			`(?i)nonemployee\s*compensation[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
		})
	default:
		// Generic gross income extraction
		data.GrossIncome = extractLineValue(text, []string{
			`(?i)(?:gross|total)[^\d]*\$?([\d,]+(?:\.\d{2})?)`,
		})
	}

	if data.GrossIncome != nil {
		data.Confidence += 0.15
	}

	if data.Confidence > 1.0 {
		data.Confidence = 1.0
	}

	return data
}

// extractLineValue tries multiple regex patterns and returns the first matching value
func extractLineValue(text string, patterns []string) *float64 {
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			// Remove commas and parse as float
			valueStr := strings.ReplaceAll(match[1], ",", "")
			if value, err := strconv.ParseFloat(valueStr, 64); err == nil && value > 0 {
				return &value
			}
		}
	}
	return nil
}

// GetRawText returns the raw text for debugging (not included in JSON by default)
func (d *ExtractedTaxData) GetRawText() string {
	return d.RawText
}
