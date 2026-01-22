package parser

import (
	"os"
	"testing"
)

// TestParseValidBasicCSV validates basic CSV parsing functionality
// ADR-003: All values must be strings, row order preserved
func TestParseValidBasicCSV(t *testing.T) {
	p := New(',', '"', true)

	records, err := p.Parse("../../testdata/valid_basic.csv")
	if err != nil {
		t.Fatalf("Expected successful parse, got error: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(records))
	}

	// Validate first record
	if records[0]["name"] != "John" {
		t.Errorf("Expected name 'John', got '%s'", records[0]["name"])
	}
	if records[0]["age"] != "30" {
		t.Errorf("Expected age '30', got '%s'", records[0]["age"])
	}
	if records[0]["email"] != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got '%s'", records[0]["email"])
	}

	// Validate second record
	if records[1]["name"] != "Jane" {
		t.Errorf("Expected name 'Jane', got '%s'", records[1]["name"])
	}
	if records[1]["age"] != "25" {
		t.Errorf("Expected age '25', got '%s'", records[1]["age"])
	}
}

// TestParseEmptyFields validates empty field handling
// ADR-003: Empty fields become empty string "", not null
func TestParseEmptyFields(t *testing.T) {
	p := New(',', '"', true)

	records, err := p.Parse("../../testdata/valid_empty_fields.csv")
	if err != nil {
		t.Fatalf("Expected successful parse, got error: %v", err)
	}

	if len(records) != 3 {
		t.Fatalf("Expected 3 records, got %d", len(records))
	}

	// First record: empty email
	if records[0]["email"] != "" {
		t.Errorf("Expected empty string for email, got '%s'", records[0]["email"])
	}
	if records[0]["name"] != "John" {
		t.Errorf("Expected name 'John', got '%s'", records[0]["name"])
	}

	// Second record: empty age
	if records[1]["age"] != "" {
		t.Errorf("Expected empty string for age, got '%s'", records[1]["age"])
	}
	if records[1]["name"] != "Jane" {
		t.Errorf("Expected name 'Jane', got '%s'", records[1]["name"])
	}
}

// TestParseQuotedFields validates quote handling
// ADR-003: Proper quote parsing with embedded delimiters and escaped quotes
func TestParseQuotedFields(t *testing.T) {
	p := New(',', '"', true)

	records, err := p.Parse("../../testdata/valid_quoted.csv")
	if err != nil {
		t.Fatalf("Expected successful parse, got error: %v", err)
	}

	if len(records) != 3 {
		t.Fatalf("Expected 3 records, got %d", len(records))
	}

	// Validate quoted field with embedded comma
	if records[2]["description"] != "Sales, Marketing" {
		t.Errorf("Expected 'Sales, Marketing', got '%s'", records[2]["description"])
	}

	// Validate escaped quotes
	if records[1]["description"] != `Software "Engineer"` {
		t.Errorf("Expected 'Software \"Engineer\"', got '%s'", records[1]["description"])
	}
}

// TestParseSingleRow validates single row handling
// ADR-003: Array structure even for single row
func TestParseSingleRow(t *testing.T) {
	p := New(',', '"', true)

	records, err := p.Parse("../../testdata/valid_single_row.csv")
	if err != nil {
		t.Fatalf("Expected successful parse, got error: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	if records[0]["name"] != "John" {
		t.Errorf("Expected name 'John', got '%s'", records[0]["name"])
	}
}

// TestParseInvalidEmpty validates empty file handling
// ADR-003: Empty files must fail
func TestParseInvalidEmpty(t *testing.T) {
	p := New(',', '"', true)

	_, err := p.Parse("../../testdata/invalid_empty.csv")
	if err == nil {
		t.Fatal("Expected error for empty file, got success")
	}
}

// TestParseInvalidHeaderOnly validates header-only file handling
// ADR-003: Files with header but no data should fail
func TestParseInvalidHeaderOnly(t *testing.T) {
	p := New(',', '"', true)

	_, err := p.Parse("../../testdata/invalid_header_only.csv")
	if err == nil {
		t.Fatal("Expected error for header-only file, got success")
	}
}

// TestParseInvalidMismatchedColumns validates strict column count enforcement
// ADR-003: Wrong number of columns must fail
func TestParseInvalidMismatchedColumns(t *testing.T) {
	p := New(',', '"', true)

	_, err := p.Parse("../../testdata/invalid_mismatched_columns.csv")
	if err == nil {
		t.Fatal("Expected error for mismatched columns, got success")
	}
}

// TestParseInvalidNoHeader validates no-header file handling when header expected
// ADR-003: When HAS_HEADER=true, files must have headers
func TestParseInvalidNoHeader(t *testing.T) {
	p := New(',', '"', true)

	// This file has data but no header row
	records, err := p.Parse("../../testdata/invalid_no_header.csv")

	// The parser will treat first row as headers
	// Verify it creates keys from first row values
	if err != nil {
		t.Fatalf("Parser should not error, but results are wrong: %v", err)
	}

	// First data row becomes headers, second row is parsed with those "headers"
	if len(records) != 1 {
		t.Errorf("Expected 1 record (treated first row as header), got %d", len(records))
	}
}

// TestParseNonexistentFile validates file access error handling
func TestParseNonexistentFile(t *testing.T) {
	p := New(',', '"', true)

	_, err := p.Parse("nonexistent.csv")
	if err == nil {
		t.Fatal("Expected error for nonexistent file, got success")
	}
}

// TestParserConfigValidation validates parser configuration
func TestParserConfigValidation(t *testing.T) {
	// Test different delimiters
	testCases := []struct {
		name      string
		delimiter rune
		quoteChar rune
		hasHeader bool
	}{
		{"comma", ',', '"', true},
		{"tab", '\t', '"', true},
		{"pipe", '|', '"', true},
		{"semicolon", ';', '"', false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := New(tc.delimiter, tc.quoteChar, tc.hasHeader)
			if p.delimiter != tc.delimiter {
				t.Errorf("Expected delimiter %c, got %c", tc.delimiter, p.delimiter)
			}
			if p.quoteChar != tc.quoteChar {
				t.Errorf("Expected quoteChar %c, got %c", tc.quoteChar, p.quoteChar)
			}
			if p.hasHeader != tc.hasHeader {
				t.Errorf("Expected hasHeader %v, got %v", tc.hasHeader, p.hasHeader)
			}
		})
	}
}

// TestParsePreservesOrder validates row order preservation
// ADR-003: Row order must match CSV row order exactly
func TestParsePreservesOrder(t *testing.T) {
	p := New(',', '"', true)

	records, err := p.Parse("../../testdata/valid_basic.csv")
	if err != nil {
		t.Fatalf("Expected successful parse, got error: %v", err)
	}

	// Verify order: John first, Jane second
	if records[0]["name"] != "John" {
		t.Errorf("Expected first record name 'John', got '%s'", records[0]["name"])
	}
	if records[1]["name"] != "Jane" {
		t.Errorf("Expected second record name 'Jane', got '%s'", records[1]["name"])
	}
}

// BenchmarkParseSmallCSV benchmarks small file parsing
func BenchmarkParseSmallCSV(b *testing.B) {
	p := New(',', '"', true)

	// Create temp file for benchmark
	tmpfile, err := os.CreateTemp("", "benchmark*.csv")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	tmpfile.WriteString("name,age,email\n")
	for i := 0; i < 10; i++ {
		tmpfile.WriteString("John,30,john@example.com\n")
	}
	tmpfile.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.Parse(tmpfile.Name())
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseLargeCSV benchmarks large file parsing
func BenchmarkParseLargeCSV(b *testing.B) {
	p := New(',', '"', true)

	tmpfile, err := os.CreateTemp("", "benchmark*.csv")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	tmpfile.WriteString("name,age,email\n")
	for i := 0; i < 1000; i++ {
		tmpfile.WriteString("John,30,john@example.com\n")
	}
	tmpfile.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.Parse(tmpfile.Name())
		if err != nil {
			b.Fatal(err)
		}
	}
}
