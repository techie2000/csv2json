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
ENV GOINSECURE=golang.org,github.com

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY VERSION ./

# Build with version information
ARG VERSION
ARG GIT_COMMIT
ARG BUILD_DATE
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s \
    -X csv2json/internal/version.GitCommit=${GIT_COMMIT} \
    -X csv2json/internal/version.BuildDate=${BUILD_DATE}" \
    -o bin/csv2json ./cmd/csv2json

# Runtime stage
FROM alpine:3.23

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/bin/csv2json .

# Create necessary directories
RUN mkdir -p /app/input /app/output /app/archive/processed /app/archive/ignored /app/archive/failed /app/logs

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

# Set environment variables
ENV INPUT_FOLDER=/app/input
ENV OUTPUT_FOLDER=/app/output
ENV ARCHIVE_PROCESSED=/app/archive/processed
ENV ARCHIVE_IGNORED=/app/archive/ignored
ENV ARCHIVE_FAILED=/app/archive/failed
ENV LOG_FILE=/app/logs/csv2json.log

# Switch to non-root user
USER appuser

# Run the application
CMD ["./csv2json"]
