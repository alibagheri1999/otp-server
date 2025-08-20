# OTP Server Makefile
# Common development commands for the OTP server

.PHONY: help build run test clean docker-build docker-run docker-stop migrate-up migrate-down deps fmt lint swagger setup build-prod dev

# Default target
help:
	@echo "OTP Server - Available Commands:"
	@echo ""
	@echo "Development:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application locally"
	@echo "  dev          - Run in development mode with hot reload"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Install/update dependencies"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"
	@echo "  docker-stop  - Stop Docker services"
	@echo "  docker-logs  - View Docker logs"
	@echo ""
	@echo "Database:"
	@echo "  migrate-up        - Run database migrations (local)"
	@echo "  migrate-up-docker- Run database migrations (Docker container)"
	@echo "  migrate-up-docker-ci - Run database migrations (Docker, non-interactive)"
	@echo "  db-status        - Check database tables"
	@echo "  db-schema        - Show table structure"
	@echo "  migrate-down     - Rollback database migrations"
	@echo ""
	@echo "Documentation:"
	@echo "  swagger      - Generate Swagger documentation"
	@echo ""
	@echo "Setup:"
	@echo "  setup        - Complete development setup"
	@echo "  build-prod   - Build production binary"
	@echo ""
	@echo "Utilities:"
	@echo "  help         - Show this help message"

# Build the application
build:
	@echo "Building OTP Server..."
	go build -o bin/otp-server cmd/main.go
	@echo "Build complete: bin/otp-server"

# Run the application locally
run: build
	@echo "Starting OTP Server..."
	./bin/otp-server

# Run in development mode (if you have air installed)
dev:
	@echo "Starting OTP Server in development mode..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "Air not found. Installing air for hot reload..."; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "Clean complete"

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t otp-server:latest .

# Run with Docker Compose
docker-run:
	@echo "Starting OTP Server with Docker Compose..."
	docker-compose up -d

# Stop Docker services
docker-stop:
	@echo "Stopping Docker services..."
	docker-compose down

# View Docker logs
docker-logs:
	@echo "Showing Docker logs..."
	docker-compose logs -f

# Run database migrations
migrate-up:
	@echo "Running database migrations..."
	@if [ -f .env ]; then \
		source .env; \
		psql -h $$POSTGRES_HOST -U $$POSTGRES_USER -d $$POSTGRES_DB -f migrations/otp_schema.sql; \
	else \
		echo "Error: .env file not found. Please copy config.env.example to .env and configure it."; \
		exit 1; \
	fi

migrate-up-docker:
	@echo "Running database migrations using Docker container..."
	@docker exec -it otp-server-postgres psql -U otp_server_user -d otp_server_db -f /docker-entrypoint-initdb.d/001_create_users_table.sql

migrate-up-docker-ci:
	@echo "Running database migrations using Docker container (non-interactive)..."
	@docker exec otp-server-postgres psql -U otp_server_user -d otp_server_db -f /docker-entrypoint-initdb.d/001_create_users_table.sql

db-status:
	@echo "Checking database status..."
	@docker exec otp-server-postgres psql -U otp_server_user -d otp_server_db -c "\dt"

db-schema:
	@echo "Showing users table structure..."
	@docker exec otp-server-postgres psql -U otp_server_user -d otp_server_db -c "\d users"

# Rollback database migrations (placeholder)
migrate-down:
	@echo "Warning: Rollback functionality not implemented yet"
	@echo "To reset database, manually drop and recreate the database"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/main.go

# Development setup
setup: deps migrate-up
	@echo "Development setup complete!"

# Production build
build-prod:
	@echo "Building production binary..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o bin/otp-server cmd/main.go
	@echo "Production build complete: bin/otp-server"

# =============================================================================
# Utility Commands
# =============================================================================

# Check if all tools are installed
check-tools:
	@echo "Checking required tools..."
	@command -v go >/dev/null 2>&1 || { echo "Go is required but not installed. Aborting." >&2; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "Docker is required but not installed. Aborting." >&2; exit 1; }
	@command -v docker-compose >/dev/null 2>&1 || { echo "Docker Compose is required but not installed. Aborting." >&2; exit 1; }
	@echo "All required tools are installed"

# Show application info
info:
	@echo "OTP Server Information:"
	@echo "Version: 1.0.0"
	@echo "Go Version: $(shell go version)"
	@echo "Docker Version: $(shell docker --version)"
	@echo "Docker Compose Version: $(shell docker-compose --version)" 