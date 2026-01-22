.PHONY: build run clean test lint docker-build docker-run

# Version information
VERSION := $(shell cat VERSION)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X csv2json/internal/version.GitCommit=$(GIT_COMMIT) -X csv2json/internal/version.BuildDate=$(BUILD_DATE)

# Build binary
build:
	go build -ldflags="$(LDFLAGS)" -o csv2json ./cmd/csv2json

# Build with optimization
build-release:
	CGO_ENABLED=0 go build -ldflags="-w -s $(LDFLAGS)" -o csv2json ./cmd/csv2json

# Run the application
run:
	go run ./cmd/csv2json

# Clean build artifacts
clean:
	rm -f csv2json csv2json-*
	go clean

# Run tests
test:
	go test -v -race -cover ./...

# Run linter
lint:
	golangci-lint run
	go vet ./...

# Format code
fmt:
	go fmt ./...

# Download dependencies
deps:
	go mod download
	go mod tidy

# Build Docker image
docker-build:
	docker build -t csv2json:latest .

# Run with Docker Compose
docker-run:
	docker-compose up -d

# Stop Docker containers
docker-stop:
	docker-compose down

# View logs
logs:
	tail -f logs/csv2json.log

# Cross-compile for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -ldflags="-w -s $(LDFLAGS)" -o csv2json-linux-amd64 ./cmd/csv2json
	GOOS=windows GOARCH=amd64 go build -ldflags="-w -s $(LDFLAGS)" -o csv2json-windows-amd64.exe ./cmd/csv2json
	GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s $(LDFLAGS)" -o csv2json-darwin-amd64 ./cmd/csv2json
	GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s $(LDFLAGS)" -o csv2json-darwin-arm64 ./cmd/csv2json

# Cross-compile target (alias)
cross-compile: build-all
