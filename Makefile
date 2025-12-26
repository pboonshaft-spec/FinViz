.PHONY: dev build docker-build docker-up docker-down docker-logs clean backend-dev backend-build docker-dev docker-dev-build docker-dev-up docker-dev-down

# Frontend Development
dev:
	npm run dev

build:
	npm run build

preview:
	npm run preview

# Backend Development
backend-dev:
	cd backend && go run ./cmd/server

backend-build:
	cd backend && go build -o server ./cmd/server

backend-test:
	cd backend && go test ./...

backend-tidy:
	cd backend && go mod tidy

# Docker - Full Stack
docker-build:
	docker-compose build --no-cache

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-logs-backend:
	docker-compose logs -f backend

docker-logs-frontend:
	docker-compose logs -f frontend

docker-restart: docker-down docker-up

# Full docker workflow (production)
docker: docker-build docker-up
	@echo "FinViz Dashboard running at http://localhost:8090"
	@echo "Backend API running at http://localhost:8085"
	@echo "MySQL available at localhost:3307"

# Docker - Development (with hot reload)
docker-dev-build:
	docker-compose -f docker-compose.dev.yml build

docker-dev-up:
	docker-compose -f docker-compose.dev.yml up -d

docker-dev-down:
	docker-compose -f docker-compose.dev.yml down

docker-dev-logs:
	docker-compose -f docker-compose.dev.yml logs -f

docker-dev: docker-dev-build docker-dev-up
	@echo "FinViz Dev Server running at http://localhost:8090 (hot reload enabled)"
	@echo "Backend API running at http://localhost:8085"
	@echo "MySQL available at localhost:3307"

# Database
db-shell:
	docker-compose exec mysql mysql -u finviz -pfinviz finviz

db-reset:
	docker-compose down -v
	docker-compose up -d mysql
	@echo "Database reset. Run 'make docker-up' to start all services."

# Clean
clean:
	rm -rf node_modules dist
	rm -f backend/server

clean-docker:
	docker-compose down -v --rmi local

# Install dependencies
install:
	npm install
	cd backend && go mod download

# Help
help:
	@echo "Available commands:"
	@echo ""
	@echo "Frontend:"
	@echo "  make dev            - Start frontend development server"
	@echo "  make build          - Build frontend for production"
	@echo "  make preview        - Preview production build"
	@echo ""
	@echo "Backend:"
	@echo "  make backend-dev    - Run backend in development mode"
	@echo "  make backend-build  - Build backend binary"
	@echo "  make backend-test   - Run backend tests"
	@echo "  make backend-tidy   - Run go mod tidy"
	@echo ""
	@echo "Docker (Production):"
	@echo "  make docker         - Build and start all services"
	@echo "  make docker-build   - Build all Docker images"
	@echo "  make docker-up      - Start all containers"
	@echo "  make docker-down    - Stop all containers"
	@echo "  make docker-logs    - View all logs"
	@echo "  make docker-restart - Restart all services"
	@echo ""
	@echo "Docker (Development - Hot Reload):"
	@echo "  make docker-dev     - Build and start dev environment"
	@echo "  make docker-dev-down - Stop dev environment"
	@echo "  make docker-dev-logs - View dev logs"
	@echo ""
	@echo "Database:"
	@echo "  make db-shell       - Open MySQL shell"
	@echo "  make db-reset       - Reset database (WARNING: deletes all data)"
	@echo ""
	@echo "Other:"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make clean-docker   - Remove Docker volumes and images"
	@echo "  make install        - Install all dependencies"
