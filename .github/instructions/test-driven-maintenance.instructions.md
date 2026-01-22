---
applyTo: '**/*.go,**/*_test.go'
description: 'Mandatory test maintenance requirements for Go code changes in csv2json project'
---

# Test-Driven Maintenance Instructions

## Core Principle
**Every functional code change MUST be accompanied by corresponding test updates or new test cases.**

## Mandatory Test Update Rules

### When to Update Tests

**ALWAYS update or create tests when:**

1. **Adding New Functions/Methods**
   - Create new test function with table-driven test cases
   - Test both happy path and error conditions
   - Include edge cases and boundary conditions
   - Add benchmarks for performance-critical code

2. **Modifying Existing Functions**
   - Update existing test cases to match new behavior
   - Add new test cases for new functionality
   - Verify all existing tests still pass
   - Update test descriptions/comments if behavior changed

3. **Changing Function Signatures**
   - Update all test calls to match new signature
   - Add tests for new parameters
   - Verify backward compatibility if applicable

4. **Modifying Return Values**
   - Update all test assertions to expect new return values
   - Test new error conditions
   - Update expected JSON output files in `testdata/` if applicable

5. **Changing Validation Logic**
   - Add test cases for new validation rules
   - Update existing validation tests
   - Test both valid and invalid inputs

6. **Modifying Configuration**
   - Update `internal/config/config_test.go`
   - Test new environment variables
   - Test new default values
   - Update validation test cases

## Test File Organization

### Test Data Files (`testdata/`)
When adding test data:
- Use descriptive filenames: `valid_[scenario].csv`, `invalid_[reason].csv`
- Create matching `*_expected.json` for valid cases
- Document in TESTING.md what the test data validates

### Test Naming Convention
```go
func Test[FunctionName][Scenario](t *testing.T) {
    // Test implementation
}
```

Examples:
- `TestParseValidBasicCSV`
- `TestParseInvalidHeaderOnly`
- `TestToJSONEmptyFields`

## ADR-003 Contract Validation

**CRITICAL**: All tests must validate ADR-003 contracts:

1. **String Values Only** - No type coercion (test that `"30"` stays `"30"`, not `30`)
2. **Empty String Not Null** - Empty fields become `""`, never `null`
3. **Array Structure** - Single row produces array, not object
4. **Row Order Preservation** - Test that row order is maintained
5. **Strict Parsing** - Test that invalid files are rejected (not silently fixed)

## Test Execution Workflow

**Before committing code:**

1. Run tests for the modified module:
   ```bash
   go test ./internal/[module] -v
   ```

2. Run all tests:
   ```bash
   go test ./... -v
   ```

3. Check coverage:
   ```bash
   go test -cover ./...
   ```

4. Verify no coverage regressions (aim for >70% per module)

## Examples

### Example 1: Adding New Parser Feature

**Code Change:**
```go
// Added support for custom delimiter
func ParseWithDelimiter(filepath string, delimiter rune) ([]map[string]string, error) {
    // implementation
}
```

**Required Test:**
```go
func TestParseWithCustomDelimiter(t *testing.T) {
    tests := []struct {
        name      string
        delimiter rune
        wantErr   bool
    }{
        {"comma", ',', false},
        {"tab", '\t', false},
        {"pipe", '|', false},
        {"invalid", '\n', true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseWithDelimiter("testdata/valid_basic.csv", tt.delimiter)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseWithDelimiter() error = %v, wantErr %v", err, tt.wantErr)
            }
            // Additional assertions...
        })
    }
}
```

### Example 2: Modifying Validation Logic

**Code Change:**
```go
// Added port range validation
func ValidatePort(port int) error {
    if port < 1 || port > 65535 {
        return fmt.Errorf("port must be between 1 and 65535, got %d", port)
    }
    return nil
}
```

**Required Test Update:**
```go
func TestValidateQueuePortRange(t *testing.T) {
    tests := []struct {
        name    string
        port    int
        wantErr bool
    }{
        {"valid_min", 1, false},
        {"valid_mid", 5672, false},
        {"valid_max", 65535, false},
        {"invalid_zero", 0, true},
        {"invalid_negative", -1, true},
        {"invalid_high", 65536, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidatePort(tt.port)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidatePort(%d) error = %v, wantErr %v", tt.port, err, tt.wantErr)
            }
        })
    }
}
```

### Example 3: Updating Test Data

**Code Change:**
```go
// Changed behavior: now preserves leading/trailing spaces in fields
```

**Required Updates:**
1. Update `testdata/valid_quoted_expected.json` with spaces
2. Add new test case:
```go
func TestParsePreservesSpaces(t *testing.T) {
    // Test that " value " stays as " value " not "value"
}
```

## Test Coverage Goals

- **Config Module**: >80% (validates all configuration paths)
- **Parser Module**: >70% (covers all CSV parsing scenarios)
- **Converter Module**: >75% (covers all JSON conversion paths)
- **New Modules**: >60% minimum for first implementation

## Continuous Integration

Tests must pass before merging:
- All existing tests pass
- New tests added for new functionality
- Coverage maintained or improved
- No skipped tests without documented reason

## Common Mistakes to Avoid

❌ **DON'T:**
- Skip tests because "it's a small change"
- Only test happy path (always test error conditions)
- Use hardcoded values that might change (use testdata files)
- Commit code without running full test suite
- Ignore test failures in other modules

✅ **DO:**
- Write tests first (TDD) when possible
- Test error conditions thoroughly
- Use table-driven tests for multiple scenarios
- Update TESTING.md when adding new test categories
- Run tests locally before pushing

## Summary Checklist

Before every commit:
- [ ] New functions have new test cases
- [ ] Modified functions have updated test cases
- [ ] All tests pass: `go test ./... -v`
- [ ] Coverage maintained: `go test -cover ./...`
- [ ] Test data files updated if behavior changed
- [ ] TESTING.md updated if new test categories added
- [ ] ADR-003 contracts validated in tests

**Remember: Tests are documentation. They explain what the code does and prove it works correctly.**
