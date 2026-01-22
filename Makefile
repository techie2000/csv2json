.PHONY: build run clean test lint docker-build docker-run

# Build binary
build:
	go build -o txt2json ./cmd/txt2json

# Build with optimization
build-release:
	CGO_ENABLED=0 go build -ldflags="-w -s" -o txt2json ./cmd/txt2json

# Run the application
run:
	go run ./cmd/txt2json

# Clean build artifacts
clean:
	rm -f txt2json txt2json-*
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
	docker build -t txt2json:latest .

# Run with Docker Compose
docker-run:
	docker-compose up -d

# Stop Docker containers
docker-stop:
	docker-compose down

# View logs
logs:
	tail -f logs/txt2json.log

# Cross-compile for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o txt2json-linux-amd64 ./cmd/txt2json
	GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o txt2json-windows-amd64.exe ./cmd/txt2json
	GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o txt2json-darwin-amd64 ./cmd/txt2json
	GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o txt2json-darwin-arm64 ./cmd/txt2json
