.PHONY: help build run clean test lint lint-fast fmt deps docker-build docker-up docker-down docker-run docker-stop logs build-all cross-compile

.DEFAULT_GOAL := help

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Version information
VERSION := $(shell cat VERSION)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X csv2json/internal/version.GitCommit=$(GIT_COMMIT) -X csv2json/internal/version.BuildDate=$(BUILD_DATE)

# Build binary
build: ## Build the binary
	@echo "Building csv2json..."
	go build -ldflags="$(LDFLAGS)" -o csv2json ./cmd/csv2json
	@echo "✅ Build complete: csv2json"

# Build with optimization
build-release: ## Build optimized release binary
	@echo "Building optimized release binary..."
	CGO_ENABLED=0 go build -ldflags="-w -s $(LDFLAGS)" -o csv2json ./cmd/csv2json
	@echo "✅ Release build complete: csv2json"

# Run the application
run: ## Run the application
	@echo "Running csv2json..."
	go run ./cmd/csv2json

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -f csv2json csv2json-*
	go clean
	@echo "✅ Clean complete"

# Run tests
test: ## Run all tests with race detection and coverage
	@echo "Running tests..."
	go test -v -race -cover ./...
	@echo "✅ Tests complete"

# Run linter
lint: ## Run linter (auto-installs golangci-lint if missing)
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@echo "Running golangci-lint..."
	golangci-lint run ./...
	@echo "✅ Linting passed!"

# Quick lint (for pre-commit hook - skip slow checks)
lint-fast: ## Run fast lint checks (for pre-commit hook)
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@echo "Running fast lint checks..."
	golangci-lint run ./... --fast
	@echo "✅ Fast lint passed!"

# Format code
fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted"

# Download dependencies
deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	go mod download
	@echo "Tidying dependencies..."
	go mod tidy
	@echo "✅ Dependencies updated"

# Build Docker image
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t csv2json:latest .
	@echo "✅ Docker image built: csv2json:latest"

# Run with Docker Compose
docker-up: ## Start services with Docker Compose
	@echo "Starting services with Docker Compose..."
	docker-compose up -d
	@echo "✅ Services started"
	@echo "📊 RabbitMQ Management: http://localhost:15672 (admin/password)"

# Stop Docker containers
docker-down: ## Stop and remove Docker containers
	@echo "Stopping Docker containers..."
	docker-compose down
	@echo "✅ Containers stopped"
# Aliases for backward compatibility
docker-run: docker-up ## Alias for docker-up (backward compatibility)

docker-stop: docker-down ## Alias for docker-down (backward compatibility)
# View logs
logs: ## Tail application logs
	@echo "Tailing logs (Ctrl+C to exit)..."
	tail -f logs/csv2json.log

# Cross-compile for multiple platforms
build-all: ## Cross-compile for all platforms
	@echo "Cross-compiling for all platforms..."
	@echo "  Building linux/amd64..."
	GOOS=linux GOARCH=amd64 go build -ldflags="-w -s $(LDFLAGS)" -o csv2json-linux-amd64 ./cmd/csv2json
	@echo "  Building windows/amd64..."
	GOOS=windows GOARCH=amd64 go build -ldflags="-w -s $(LDFLAGS)" -o csv2json-windows-amd64.exe ./cmd/csv2json
	@echo "  Building darwin/amd64..."
	GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s $(LDFLAGS)" -o csv2json-darwin-amd64 ./cmd/csv2json
	@echo "  Building darwin/arm64..."
	GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s $(LDFLAGS)" -o csv2json-darwin-arm64 ./cmd/csv2json
	@echo "✅ Cross-compilation complete"

# Cross-compile target (alias)
cross-compile: build-all ## Alias for build-all
