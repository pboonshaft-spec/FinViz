# Sample CSV Data Files

This directory contains example CSV files demonstrating the expected input formats for the Financial Analytics Dashboard.

## Supported Formats

The dashboard supports flexible CSV column detection. You can use any of the following formats:

### 1. Personal Expenses Format (`personal-expenses.csv`)

Single amount column with positive values for income and negative for expenses:

```csv
Date,Amount,Category,Description
2024-01-01,3500.00,Salary,Monthly Paycheck
2024-01-03,-1200.00,Rent,Monthly Rent Payment
```

### 2. Bank Statement Format (`bank-statement.csv`)

Separate debit and credit columns (common in bank exports):

```csv
Posted Date,Debit,Credit,Category,Description
01/02/2024,,4250.00,Salary,DIRECT DEPOSIT - ACME CORP
01/03/2024,1450.00,,Rent,ACH PAYMENT - PROPERTY MGMT
```

### 3. Minimal Format (`minimal-example.csv`)

Just the essentials - date and amount:

```csv
Date,Amount
2024-01-15,2500.00
2024-01-16,-150.00
```

## Column Reference

| Column | Required | Recognized Headers | Notes |
|--------|----------|-------------------|-------|
| Date | Yes | `Date`, `Posted Date`, `Transaction Date` | Supports ISO (YYYY-MM-DD) and MM/DD/YYYY formats |
| Amount | Yes* | `Amount` | Positive = income, Negative = expense |
| Debit | Yes* | `Debit`, `Withdrawal` | Alternative to Amount column |
| Credit | Yes* | `Credit`, `Deposit` | Alternative to Amount column |
| Category | No | `Category`, `Type` | Used for grouping transactions |
| Description | No | `Description`, `Merchant`, `Name` | Transaction details |

*Either a single Amount column OR separate Debit/Credit columns are required.

## Tips

- Column headers are case-insensitive
- Currency symbols (`$`) and commas are automatically removed
- Empty cells in Debit/Credit columns are treated as zero
- Multiple CSV files can be uploaded simultaneously
- If no category is provided, transactions are marked as "Uncategorized"
