# Go Linting Guide

This document describes the linting setup and common issues found in the csv2json codebase.

## Tools

### golangci-lint

We use [golangci-lint](https://golangci-lint.run/) to run multiple Go linters in parallel.

**Installation:**

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

**Running locally:**

```bash
# Run all linters
golangci-lint run ./...

# Run with verbose output
golangci-lint run ./... --verbose

# Auto-fix issues where possible
golangci-lint run ./... --fix
```

## Configuration

See [`.golangci.yml`](../.golangci.yml) for the full configuration.

### Enabled Linters

**Core Quality:**

- `errcheck` - Check for unchecked errors
- `gosimple` - Simplify code
- `govet` - Go vet checks (including fieldalignment)
- `ineffassign` - Detect ineffectual assignments
- `staticcheck` - Static analysis
- `unused` - Check for unused code

**Style & Formatting:**

- `gofmt` - Check formatting
- `goimports` - Check import organization
- `misspell` - Spell checker
- `whitespace` - Detect leading/trailing whitespace

**Code Quality:**

- `gocritic` - 100+ checks for code quality
- `revive` - Replacement for golint
- `cyclop` - Function complexity checks

**Security:**

- `gosec` - Security vulnerability scanner

**Performance:**

- `prealloc` - Find slice declarations that could be preallocated
- `unconvert` - Remove unnecessary type conversions

## Common Issues and Fixes

### 1. Octal Literal Style (gocritic)

**Issue:** Old-style octal literals `0755` should use new style `0o755`

```go
// ❌ Old style
os.MkdirAll(dir, 0755)
os.WriteFile(path, data, 0644)

// ✅ New style (Go 1.13+)
os.MkdirAll(dir, 0o755)
os.WriteFile(path, data, 0o644)
```

**Status:** Temporarily excluded in `.golangci.yml` - migrate in batch

### 2. Struct Field Alignment (govet)

**Issue:** Struct fields can be reordered to reduce memory usage

```go
// ❌ Suboptimal alignment (352 bytes)
type Config struct {
    ArchiveFailed       string // 8 bytes (pointer)
    ArchiveProcessed    string // 8 bytes (pointer)
    QueuePort           int    // 8 bytes
    QueueSSL            bool   // 1 byte
    // ... more fields
}

// ✅ Better alignment (336 bytes)
type Config struct {
    // Group 8-byte fields together
    ArchiveFailed       string
    ArchiveProcessed    string
    QueuePort           int
    
    // Group smaller fields at end
    QueueSSL            bool
}
```

**Fix:** Use `fieldalignment` tool or manually reorder fields

### 3. Import Formatting (goimports)

**Issue:** Imports not properly grouped or formatted

```go
// ❌ Not formatted
import (
    "csv2json/internal/parser"
    "encoding/json"
    "fmt"
)

// ✅ Properly formatted (groups: stdlib, external, local)
import (
    "encoding/json"
    "fmt"

    "csv2json/internal/parser"
)
```

**Fix:** Run `goimports -w .` or use IDE auto-format

### 4. Shadow Variables (govet)

**Issue:** Variable declaration shadows an outer variable

```go
// ❌ Shadows 'err'
err := somethingFirst()
if err != nil {
    return err
}

if err := somethingSecond(); err != nil {  // Shadows outer 'err'
    return err
}

// ✅ Use different variable name or reassign
err := somethingFirst()
if err != nil {
    return err
}

err = somethingSecond()  // Reuses outer 'err'
if err != nil {
    return err
}
```

### 5. Cyclomatic Complexity (cyclop)

**Issue:** Function has too many branches/conditions

```go
// ❌ Complex function (complexity 27)
func LoadRoutes(configPath string) (*RoutesConfig, error) {
    // Many if/else branches
    // Multiple nested conditions
    // Hard to test and maintain
}

// ✅ Refactor into smaller functions
func LoadRoutes(configPath string) (*RoutesConfig, error) {
    config, err := parseRoutesFile(configPath)
    if err != nil {
        return nil, err
    }
    
    if err := validateRoutes(config); err != nil {
        return nil, err
    }
    
    if err := applyDefaults(config); err != nil {
        return nil, err
    }
    
    return config, nil
}
```

**Current Status:** Max complexity set to 30 (temporarily) - refactor `LoadRoutes` to reduce

### 6. Error Comparison (errorlint)

**Issue:** Using `!=` for error comparison fails with wrapped errors

```go
// ❌ Won't work with wrapped errors
if err != nil && err != io.EOF {
    return err
}

// ✅ Use errors.Is for wrapped error checking
if err != nil && !errors.Is(err, io.EOF) {
    return err
}
```

### 7. Empty String Test (gocritic)

**Issue:** Using `len()` to check if string is empty

```go
// ❌ Less readable
if len(str) > 0 {
    // ...
}

// ✅ More idiomatic
if str != "" {
    // ...
}
```

### 8. Range Loop Copies (gocritic)

**Issue:** Range loop copies large values on each iteration

```go
// ❌ Copies 296 bytes per iteration
for i, route := range routesConfig.Routes {
    processRoute(route)
}

// ✅ Use index to avoid copy
for i := range routesConfig.Routes {
    processRoute(routesConfig.Routes[i])
}

// ✅ Or use pointers
for i := range routesConfig.Routes {
    route := &routesConfig.Routes[i]
    processRoute(route)
}
```

### 9. Unchecked Errors (errcheck)

**Issue:** Error return value not checked

```go
// ❌ Error ignored
_ = godotenv.Load()

// ✅ Check error (or document why it's safe to ignore)
if err := godotenv.Load(); err != nil {
    // .env file optional, use defaults
    log.Debug("No .env file found, using defaults")
}
```

## GitHub Actions Integration

Linting runs automatically on every push and pull request via [`.github/workflows/lint.yml`](../.github/workflows/lint.yml).

The workflow includes:

- **Go linting** (golangci-lint)
- **Markdown linting** (markdownlint)
- **Dockerfile linting** (hadolint)

## Continuous Improvement Plan

### Phase 1: Quick Fixes (Low Effort, High Impact)

- [ ] Fix import formatting with `goimports -w .`
- [ ] Fix octal literals: Replace `0755` → `0o755`, `0644` → `0o644`
- [ ] Fix empty string tests: Replace `len(str) > 0` → `str != ""`
- [ ] Fix error comparisons: Use `errors.Is()` for wrapped errors

### Phase 2: Refactoring (Medium Effort)

- [ ] Refactor `LoadRoutes()` to reduce cyclomatic complexity from 27 to <15
- [ ] Fix shadow variables by using unique names or reassignment
- [ ] Fix range loop copies by using indices or pointers

### Phase 3: Optimization (Low Priority)

- [ ] Optimize struct field alignment to reduce memory usage
- [ ] Investigate type assertion chains and convert to switch statements

## Resources

- [golangci-lint Documentation](https://golangci-lint.run/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
