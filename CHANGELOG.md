# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.0] - 2026-01-23

### Added
- **Message Envelope with Provenance Metadata ([ADR-006](docs/adrs/ADR-006-message-envelope-and-provenance-metadata.md))**: Queue messages now include comprehensive metadata envelope preventing "context amnesia" in distributed systems
  - **Ingestion Contract**: Schema/version identifier (e.g., `products.csv.v1`) enables contract-based routing instead of payload shape inference
  - **Source Provenance**: Full audit trail including source file path, queue name, broker URI, and route name
  - **Ingestion Metadata**: Service name, semantic version (read from VERSION file), and ISO8601 timestamp
  - Downstream services can branch on `meta.ingestionContract` instead of guessing intent from payload structure
  - Eliminates fragile patterns like regex'ing filenames or "this queue usually means X" logic
  - Supports safe schema evolution and multiple contract versions coexisting
- **Automated Release System**: GitHub Actions workflow for automated releases on version tags
  - Automatic binary builds for 5 platforms (linux-amd64, linux-arm64, windows-amd64, darwin-amd64, darwin-arm64)
  - SHA256 checksum generation for all binaries
  - Automatic GitHub Release creation with CHANGELOG extraction
  - Multi-architecture Docker image builds (linux/amd64, linux/arm64)
  - Automatic push to GitHub Container Registry (ghcr.io)
  - Pre-release detection for alpha/beta/rc versions

### Changed
- Installation documentation now prioritizes pre-built binaries over building from source
- Added comprehensive Docker image usage instructions with ghcr.io registry
- Service version now read dynamically from VERSION file instead of hardcoded constant
  - Fallback version changed from hardcoded "0.2.0" to "unknown" to prevent maintenance trap
  - All version-related functions now use dynamic VERSION file reading
  - If VERSION file missing, service reports version as "unknown" instead of stale hardcoded value
- Queue message format changed to include full metadata envelope (see ADR-006)
  - **BREAKING CHANGE**: Queue message structure changed from simple `{"identifier": "...", "data": [...]}` to `{"meta": {...}, "data": [...]}`
  - Downstream consumers must update to parse new envelope structure
  - Route configuration now requires `ingestionContract` field
  - Configuration field renamed: `includeRouteContext` → `includeEnvelope`

## [0.2.0] - 2026-01-22

### Added
- **Multi-Ingress Routing ([ADR-004](docs/adrs/ADR-004-multi-ingress-routing-architecture.md))**: Service now supports monitoring multiple input folders, each with distinct output destinations and queue routing configurations via `routes.json`
- **Hybrid File Detection Strategy ([ADR-005](docs/adrs/ADR-005-hybrid-file-detection-strategy.md))**: Event-driven file monitoring using OS-level notifications (fsnotify) for near-instant file detection (<100ms latency)
  - Three watch modes: `event` (default, OS notifications), `poll` (legacy polling), `hybrid` (event + backup polling)
  - Automatic fallback to polling if event monitoring unavailable
  - Configurable via `WATCH_MODE` environment variable or `watchMode` in routes.json
- **Comprehensive CLI Documentation**: Added `--help` flag with detailed documentation covering:
  - All operational modes (legacy/routing)
  - Configuration options and environment variables
  - Watch mode strategies
  - Architecture Decision Records (ADR) references
  - Example configurations
- **Version Management System**: Added `--version` flag and internal version tracking
  - VERSION file in project root
  - `internal/version` package with version constants
  - Build-time version injection via ldflags

### Changed
- Project renamed from `txt2json` to `csv2json` to better reflect actual functionality
- Default `includeRouteContext` changed to `true` in routing mode (metadata now included by default)
- Refactored monitoring system to support multiple strategies via factory pattern
- Improved struct field alignment and code formatting consistency

### Fixed
- JSON unmarshaling of boolean config values now properly distinguishes between `false` and unset (using pointer types)

### Performance
- Event-driven monitoring reduces file detection latency from 5+ seconds to <100ms
- Eliminates CPU waste from continuous polling of empty directories
- Hybrid mode provides resilience with 60-second backup polling

### Documentation
- Added ADR-004: Multi-Ingress Routing Architecture
- Added ADR-005: Hybrid File Detection Strategy
- Updated README with watch mode configuration guidance
- Added source filename datetime convention recommendations to ADR-003

## [0.1.0] - 2026-01-21

### Added
- Initial release of csv2json file processing service
- **Core Processing Pipeline**:
  - CSV to JSON conversion with proper type handling
  - Automatic file monitoring and processing
  - Configurable archive management (processed/failed/ignored)
  - RabbitMQ queue integration for downstream processing
- **Configuration System**:
  - Environment variable based configuration
  - Support for `.env` files
  - Input/output folder configuration
  - Archive folder categorization
- **File Handling**:
  - UTF-8 encoding support with BOM detection
  - Configurable file extensions (.csv, .txt)
  - Ignore patterns for excluding specific files
  - Timestamp-based archive naming
- **Error Handling**:
  - Comprehensive error logging
  - Failed file archiving with .error files
  - Validation for empty files and missing headers
- **Architecture**:
  - Modular package structure (parser, converter, monitor, archiver, output)
  - Concurrent processing with goroutines
  - Clean separation of concerns following SOLID principles
- **Testing**:
  - Comprehensive unit tests for all modules
  - Test coverage >70% per module
  - Integration test data in testdata/
- **Deployment**:
  - Docker support with multi-stage builds
  - Docker Compose for local development
  - Makefile for common operations
- **Documentation**:
  - Comprehensive README with configuration guide
  - TESTING.md with test execution instructions
  - ADR-001: Go language selection rationale
  - ADR-002: RabbitMQ queue selection rationale
  - ADR-003: Core system behavioral principles
  - Security documentation (SECURITY.md, SECURITY-IMPLEMENTATION.md)

### Technical Details
- Language: Go 1.x
- Dependencies: Standard library + RabbitMQ client
- File Detection: Time-based polling (5-second intervals)
- Architecture: Single-ingress, single-output design

---

## Version History

- **v0.3.0** (2026-01-23): Message envelope with provenance metadata (ADR-006), Dynamic VERSION reading
- **v0.2.0** (2026-01-22): Major features: Multi-ingress routing, Event-driven monitoring
- **v0.1.0** (2026-01-21): Initial release

[Unreleased]: https://github.com/techie2000/csv2json/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/techie2000/csv2json/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/techie2000/csv2json/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/techie2000/csv2json/releases/tag/v0.1.0
