# Testing Guide

## Test Status

✅ **52 Tests Passing** (as of latest run)

### Test Coverage

- **Archiver Module**: 77.1% coverage (9/9 tests passing + 2 benchmarks)
- **Config Module**: 82.1% coverage (13/13 tests passing)
- **Converter Module**: 75.0% coverage (10/10 tests passing + 2 benchmarks)
- **Monitor Module**: 84.9% coverage (9/9 tests passing + 2 benchmarks)
- **Output Module**: 13.6% coverage (7/7 tests passing + 2 benchmarks)
- **Parser Module**: 61.2% coverage (11/11 tests passing + 2 benchmarks)

### Modules Without Tests

- `cmd/csv2json` - 0% (main entry point, no tests yet)
- `internal/processor` - 0% (planned - integration tests)

## Test Infrastructure

This project has comprehensive test coverage validating
[ADR-003](docs/adrs/ADR-003-core-system-principles.md#5-data-type-contract-strings-not-nulls) contracts:

### Test Data (`testdata/`)

**Valid CSV Files:**

- `valid_basic.csv` - Basic 2-row CSV with all fields
- `valid_empty_fields.csv` - CSV with empty fields (tests empty string handling)
- `valid_quoted.csv` - CSV with quoted fields and embedded delimiters
- `valid_single_row.csv` - Single row CSV (tests array structure)

**Invalid CSV Files:**

- `invalid_empty.csv` - Empty file (should fail)
- `invalid_header_only.csv` - Header with no data rows
- `invalid_mismatched_columns.csv` - Rows with wrong column count
- `invalid_no_header.csv` - Data without header row

**Expected Output Files:**

- `*_expected.json` - Expected JSON output for valid CSV files

### Test Modules

**`internal/archiver/archiver_test.go`** (NEW)

- Archiver initialization and configuration
- Directory creation on archive
- Archiving with/without timestamps
- Duplicate filename handling (counter suffix)
- Error log generation for failed files
- Category-based archiving (processed/ignored/failed)
- Copy+delete fallback for cross-device links (Docker volumes)
- File content preservation
- Performance benchmarks (with/without timestamps)

**`internal/monitor/monitor_test.go`** (NEW)

- Monitor initialization and configuration
- Existing file detection (ignore on startup)
- MAX_FILES_PER_POLL enforcement (rate limiting)
- New file detection and processing
- Processed file tracking (skip duplicates)
- Directory filtering (ignore subdirectories)
- Stop mechanism and graceful shutdown
- Zero limit handling (unlimited processing)
- Performance benchmarks (small/large file sets)

**`internal/output/queue_handler_test.go`** (NEW)

- Message marshaling to JSON format
- Queue message structure validation
- Identifier field validation
- String value enforcement
  ([ADR-003](docs/adrs/ADR-003-core-system-principles.md#5-data-type-contract-strings-not-nulls))
- Empty data array handling
- Large dataset handling (1000+ records)
- Unsupported queue type error handling
- Not-yet-implemented queue types (Kafka, SQS, Azure Service Bus)

**`internal/parser/parser_test.go`**

- CSV parsing strictness (delimiter, quote handling)
- Empty field handling (empty string not null)
- Column count validation
- Row order preservation
- Quote escaping
- Performance benchmarks

**`internal/converter/converter_test.go`**

- JSON structure (array of objects)
- String value enforcement (no type coercion)
- Empty field handling (empty string not null)
- Single row array structure
- Special character escaping
- File output functionality
- Performance benchmarks

**`internal/config/config_test.go`**

- Environment variable loading
- Default value validation
- Queue configuration validation (QUEUE_TYPE enum, port range)
- MAX_FILES_PER_POLL configuration
- File filter configuration
- Fail-fast validation
- Configuration edge cases

## Running Tests

### All Tests

```basharchiver -v
go test ./internal/config -v
go test ./internal/converter -v
go test ./internal/monitor -v
go test ./internal/output -v
go test ./internal/parser

### Specific Module

```bash
go test ./internal/parser -v
go test ./internal/converter -v
go test ./internal/config -v
```

### With Coverage

```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Benchmarks

```bash
# All benchmarks
go test -bench=. ./...

# Specific modules
go test -bench=. ./internal/archiver
go test -bench=. ./internal/converter
go test -bench=. ./internal/monitor
go test -bench=. ./internal/output
go test -bench=. ./internal/parser
```

### Race Detection

```bash
go test -race ./...
```

## ADR-003 Contract Validation

Tests validate the following ADR-003 requirements:

### CSV Parsing Strictness (Section 3)

- ✅ Delimiter must match configuration
- ✅ Wrong column count must fail
- ✅ Quote mismatch must fail
- ✅ Empty files must fail
- ✅ Header required when HAS_HEADER=true

### File Output Mode JSON Contract (Section 5)

- ✅ Top-level array structure (even for single row)
- ✅ All values are strings (no type coercion: "30" not 30)
- ✅ Empty fields become `""` not `null`
- ✅ Headers are exact CSV header text (case-sensitive, whitespace preserved)
- ✅ Archiving Requirements (Section 8)

- ✅ Three categories: processed/ignored/failed
- ✅ Timestamp suffixes when enabled
- ✅ Duplicate handling with counter suffix
- ✅ Error log creation for failed files
- ✅ Cross-device link handling (Docker volumes)

### File Monitoring (Section 2)

- ✅ MAX_FILES_PER_POLL enforcement
- ✅ Existing file detection (skip on startup)
- ✅ New file detection and processing
- ✅ Directory filtering
- ✅ Processed file tracking

### Queue Output Mode (Section 5)

- ✅ Message structure validation (identifier + data fields)
- ✅ String value enforcement
- ✅ Empty data handling
- ✅ Large dataset support
- ⏳ RabbitMQ integration tests (requires live queue)
- ⏳ Kafka/SQS/Azure Service Bus (not yet implemented)
- ✅ Environment variables only (12-factor compliance)
- ✅ Explicit over implicit (QUEUE_TYPE enum validation)
- ✅ Fail fast with clear errors
- ✅ Port range validation (1-65535)
- ✅ MAX_FILES_PER_POLL rate limiting

### Streaming JSON Implementation (Section 9)

- ⏳ Not yet implemented (stubbed)
- Will validate: Row-by-row processing, streaming output, no in-memory CSV loading

## Expected Test Results

All tests should pass with these validations:

```text
=== RUN   TestParseValidBasicCSV
--- PASS: TestParseValidBasicCSV
=== RUN   TestParseEmptyFields
--- PASS: TestParseEmptyFields
=== RUN   TestParseInvalidMismatchedColumns
--- PASS: TestParseInvalidMismatchedColumns
=== RUN   TestToJSONStringValues
--- PASS: TestToJSONStringValues
=== RUN   TestToJSONEmptyFields
--- PASS: TestToJSONEmptyFields
=== RUN   TestValidateQueueConfig
--- PASS: TestValidateQueueConfig
```

## Test Coverage Goals

- **Parser**: >90% coverage (critical path validation)
- **Converter**: >90% coverage (JSON contract enforcement)
- **Config**: >85% coverage (all validation paths)
- **Overall**: >85% coverage

## Continuous Integration

Add to CI/CD pipeline:

```yaml
# .github/workflows/test.yml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      - run: go test -race -coverprofile=coverage.out ./...
      - run: go tool cover -func=coverage.out
```

## Adding New Tests

When adding features, ensure tests validate:

1. **Contract Compliance**: Does output match ADR-003 specification?
2. **Error Handling**: Do failures produce clear error messages?
3. **Edge Cases**: Empty data, single row, special characters
4. **Performance**: Benchmarks for performance-critical paths
5. **Integration**: End-to-end workflows

## Test Data Management

Test data in `testdata/` directory:

- ✅ Version controlled
- ✅ Small files (< 1KB each)
- ✅ Representative scenarios
- ✅ Both valid and invalid cases
- ✅ Expected output files for comparison

## Known Issues

- Parser tests require relative path `../../testdata/` (run from module directory)
- Go must be installed and in PATH
- Some tests create temporary files (cleaned up automatically)

## Troubleshooting

**Tests fail with "file not found":**

- Ensure you're running tests from project root or module directory
- Check testdata/ directory exists with CSV files

**Tests fail with "go not found":**

- Install Go 1.25 or later
- Add Go to system PATH

**Coverage report doesn't open:**

- Install Go tools: `go install golang.org/x/tools/cmd/cover@latest`
- Ensure browser is configured

## Next Steps

**Pending Test Implementation:**

- `internal/archiver/archiver_test.go` - File archiving validation
- `internal/monitor/monitor_test.go` - Directory polling and MAX_FILES_PER_POLL
- `internal/output/queue_handler_test.go` - Queue message structure validation
- Integration tests for end-to-end workflows
