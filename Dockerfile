# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Install ca-certificates and git for Go modules
RUN apk add --no-cache ca-certificates git

# Configure git to skip SSL verification (for corporate proxies)
RUN git config --global http.sslVerify false

# Set Go proxy to direct (bypass proxy.golang.org)
ENV GOPROXY=direct
ENV GOSUMDB=off

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o txt2json ./cmd/txt2json

# Runtime stage
FROM alpine:3.23

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/txt2json .

# Create necessary directories
RUN mkdir -p /app/input /app/output /app/archive/processed /app/archive/ignored /app/archive/failed /app/logs

# Set environment variables
ENV INPUT_FOLDER=/app/input
ENV OUTPUT_FOLDER=/app/output
ENV ARCHIVE_PROCESSED=/app/archive/processed
ENV ARCHIVE_IGNORED=/app/archive/ignored
ENV ARCHIVE_FAILED=/app/archive/failed
ENV LOG_FILE=/app/logs/txt2json.log

# Run the application
CMD ["./txt2json"]
