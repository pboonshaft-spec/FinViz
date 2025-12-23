# CLAUDE.md

## Project Overview

Financial Analytics Dashboard - A web-based tool for visualizing and analyzing personal financial data from CSV files.

## Important Guidelines

### Git Workflow

**CRITICAL: Always use feature branches. Never push directly to main.**

When making changes to this project:

1. Create a feature branch before making any changes (e.g., `feat/feature-name`, `fix/bug-name`)
2. Make your changes and commits on the feature branch
3. Push the feature branch to remote
4. Do NOT push or merge directly to main
5. Main branch merges will be handled separately

### Sample Data Maintenance

When modifying CSV parsing logic or adding support for new column types in the application:

1. Update the example files in `sample-data/` to demonstrate any new formats
2. Update `sample-data/README.md` to document new column types or format changes
3. Ensure all three example files remain valid and parseable after changes

### CSV Format Reference

The app uses flexible column detection via PapaParse. Recognized columns:
- **Date**: `date`, `posted` (required)
- **Amount**: `amount` (single column, positive=income, negative=expense)
- **Debit/Credit**: `debit`/`withdrawal` and `credit`/`deposit` (alternative to Amount)
- **Category**: `category`, `type` (optional)
- **Description**: `description`, `merchant`, `name` (optional)
