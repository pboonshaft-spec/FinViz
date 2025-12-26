# CLAUDE.md

## Project Overview

Financial Analytics Dashboard - A full-stack application for personal financial management featuring:
- **Budget Tab**: CSV-based transaction analysis with ApexCharts visualizations
- **Net Worth Tab**: Assets/debts tracking with Monte Carlo investment projections

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     React Frontend (Vite)                   │
│  ┌──────────────────┐  ┌──────────────────────────────────┐│
│  │  Budget Tab      │  │  Net Worth Tab                   ││
│  │  - CSV upload    │  │  - Assets/Debts CRUD             ││
│  │  - Charts        │  │  - Monte Carlo projections       ││
│  └──────────────────┘  └──────────────────────────────────┘│
│                     Auth: Login/Register                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Go API Server (:8081)                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ /api/auth   │  │ /api/assets │  │ /api/monte-carlo    │ │
│  │ /api/debts  │  │ /api/import │  │ /api/asset-types    │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
│               JWT Auth + Multi-tenant isolation             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                       MySQL Database (:3307)                │
│        users | assets | debts | asset_types                 │
└─────────────────────────────────────────────────────────────┘
```

## Tech Stack

### Frontend
- **Framework**: React 18 + Vite
- **Charts**: ApexCharts (react-apexcharts) + ApexSankey
- **CSV Parsing**: PapaParse
- **Styling**: CSS with CSS Variables (Dark Mode)

### Backend
- **Language**: Go 1.22+
- **Database**: MySQL 8.0
- **Auth**: JWT-based with bcrypt password hashing

## Project Structure

```
/
├── src/                         # React Frontend
│   ├── components/
│   │   ├── auth/                # Login/Register forms
│   │   ├── charts/              # ApexCharts components
│   │   ├── networth/            # Asset/Debt management
│   │   └── tabs/                # Budget/NetWorth tabs
│   ├── contexts/
│   │   └── AuthContext.jsx      # Auth state management
│   ├── hooks/
│   │   ├── useApi.js            # API client with auth
│   │   ├── useDataProcessor.js
│   │   └── useFileUpload.js
│   └── styles/
│
├── backend/                     # Go API
│   ├── cmd/server/main.go       # Entry point
│   ├── internal/
│   │   ├── api/                 # HTTP handlers
│   │   ├── auth/                # JWT + bcrypt
│   │   ├── db/                  # MySQL connection
│   │   ├── ingestion/           # Data import (CSV, future Plaid)
│   │   ├── models/              # Data models
│   │   └── simulation/          # Monte Carlo engine
│   ├── Dockerfile
│   └── go.mod
│
├── docker-compose.yml           # Full stack deployment
└── Makefile
```

## Important Guidelines

### Git Workflow

**CRITICAL: Always use feature branches. Never push directly to main.**

1. Create a feature branch before making any changes
2. Make your changes and commits on the feature branch
3. Push the feature branch to remote
4. Do NOT push or merge directly to main

### Multi-Tenant Data Isolation

All user data is isolated by `user_id`. When adding new tables or queries:
- Always include `user_id` column in tables that store user data
- Always filter by `user_id` in queries
- Use `getUserFromContext(r)` in handlers to get the authenticated user

### Sample Data Maintenance

When modifying CSV parsing logic:
1. Update example files in `sample-data/`
2. Update `sample-data/README.md` to document changes
3. Ensure all example files remain valid

## API Endpoints

### Public (No Auth Required)
- `POST /api/auth/register` - Create account
- `POST /api/auth/login` - Login
- `GET /api/asset-types` - Get asset types
- `GET /api/health` - Health check

### Protected (Requires Auth)
- `GET /api/auth/me` - Get current user
- `GET/POST /api/assets` - List/Create assets
- `PUT/DELETE /api/assets/{id}` - Update/Delete asset
- `GET/POST /api/debts` - List/Create debts
- `PUT/DELETE /api/debts/{id}` - Update/Delete debt
- `POST /api/monte-carlo` - Run simulation
- `POST /api/import/csv` - Import CSV data

## Development

### Frontend Only
```bash
npm install
npm run dev     # Start on port 3000
```

### Backend Only
```bash
cd backend
go run ./cmd/server  # Requires MySQL running
```

### Full Stack (Docker)
```bash
make docker       # Start all services
# Frontend: http://localhost:8090
# Backend:  http://localhost:8081
# MySQL:    localhost:3307
```

## Makefile Commands

```bash
# Frontend
make dev            # Start frontend dev server
make build          # Build frontend for production

# Backend
make backend-dev    # Run backend locally
make backend-build  # Build backend binary
make backend-tidy   # Run go mod tidy

# Docker (Full Stack)
make docker         # Build and start all services
make docker-down    # Stop all services
make docker-logs    # View all logs
make db-shell       # Open MySQL shell
make db-reset       # Reset database (WARNING: deletes all data)

# Other
make clean          # Remove build artifacts
make install        # Install all dependencies
make help           # Show all commands
```

## Environment Variables

### Backend
- `DB_HOST` - MySQL host (default: localhost)
- `DB_PORT` - MySQL port (default: 3306)
- `DB_USER` - MySQL user (default: finviz)
- `DB_PASSWORD` - MySQL password (default: finviz)
- `DB_NAME` - Database name (default: finviz)
- `JWT_SECRET` - JWT signing secret (auto-generated if not set)
- `PORT` - Server port (default: 8080)

### Frontend
- `VITE_API_URL` - Backend API URL (default: http://localhost:8081)
