package converter

import (
	"csv2json/internal/parser"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestToJSON validates JSON conversion functionality
// ADR-003: Array structure, string values only, empty string not null
func TestToJSON(t *testing.T) {
	c := New()

	data := []map[string]string{
		{"name": "John", "age": "30", "email": "john@example.com"},
		{"name": "Jane", "age": "25", "email": "jane@example.com"},
	}

	jsonBytes, err := c.ToJSON(data)
	if err != nil {
		t.Fatalf("Expected successful conversion, got error: %v", err)
	}

	// Validate it's valid JSON
	var result []map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Generated invalid JSON: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 records, got %d", len(result))
	}
}

// TestToJSONStringValues validates all values are strings
// ADR-003: No type coercion - "30" not 30, "true" not true
func TestToJSONStringValues(t *testing.T) {
	c := New()

	data := []map[string]string{
		{"name": "John", "age": "30", "active": "true", "balance": "100.50"},
	}

	jsonBytes, err := c.ToJSON(data)
	if err != nil {
		t.Fatalf("Expected successful conversion, got error: %v", err)
	}

	// Parse and verify types
	var result []map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Generated invalid JSON: %v", err)
	}

	// Verify all values are strings, not numbers or booleans
	record := result[0]

	if age, ok := record["age"].(string); !ok || age != "30" {
		t.Errorf("Expected age to be string '30', got %v (%T)", record["age"], record["age"])
	}

	if active, ok := record["active"].(string); !ok || active != "true" {
		t.Errorf("Expected active to be string 'true', got %v (%T)", record["active"], record["active"])
	}

	if balance, ok := record["balance"].(string); !ok || balance != "100.50" {
		t.Errorf("Expected balance to be string '100.50', got %v (%T)", record["balance"], record["balance"])
	}
}

// TestToJSONEmptyFields validates empty field handling
// ADR-003: Empty fields become empty string "", not null
func TestToJSONEmptyFields(t *testing.T) {
	c := New()

	data := []map[string]string{
		{"name": "John", "age": "30", "email": ""},
		{"name": "Jane", "age": "", "email": "jane@example.com"},
	}

	jsonBytes, err := c.ToJSON(data)
	if err != nil {
		t.Fatalf("Expected successful conversion, got error: %v", err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Generated invalid JSON: %v", err)
	}

	// First record: empty email should be "", not nil
	if email := result[0]["email"]; email != "" {
		t.Errorf("Expected empty string for email, got %v", email)
	}

	// Second record: empty age should be "", not nil
	if age := result[1]["age"]; age != "" {
		t.Errorf("Expected empty string for age, got %v", age)
	}
}

// TestToJSONSingleRow validates single row handling
// ADR-003: Array structure even for single row
func TestToJSONSingleRow(t *testing.T) {
	c := New()

	data := []map[string]string{
		{"name": "John", "age": "30"},
	}

	jsonBytes, err := c.ToJSON(data)
	if err != nil {
		t.Fatalf("Expected successful conversion, got error: %v", err)
	}

	// Verify it's an array, not a single object
	if jsonBytes[0] != '[' {
		t.Error("Expected JSON to start with '[', indicating array structure")
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Generated invalid JSON array: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected array with 1 element, got %d", len(result))
	}
}

// TestToJSONEmptyArray validates empty data handling
func TestToJSONEmptyArray(t *testing.T) {
	c := New()

	data := []map[string]string{}

	jsonBytes, err := c.ToJSON(data)
	if err != nil {
		t.Fatalf("Expected successful conversion, got error: %v", err)
	}

	// Should produce empty array []
	var result []map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Generated invalid JSON: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected empty array, got %d elements", len(result))
	}
}

// TestToJSONFile validates file output functionality
func TestToJSONFile(t *testing.T) {
	c := New()

	data := []map[string]string{
		{"name": "John", "age": "30"},
	}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "converter_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	outputPath := filepath.Join(tmpDir, "output.json")

	if err := c.ToJSONFile(data, outputPath); err != nil {
		t.Fatalf("Expected successful file write, got error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}

	// Verify file content
	fileBytes, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(fileBytes, &result); err != nil {
		t.Fatalf("File contains invalid JSON: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 record in file, got %d", len(result))
	}
}

// TestToJSONFileCreatesDirectory validates directory creation
func TestToJSONFileCreatesDirectory(t *testing.T) {
	c := New()

	data := []map[string]string{
		{"name": "John"},
	}

	tmpDir, err := os.MkdirTemp("", "converter_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Path with nested directory that doesn't exist
	outputPath := filepath.Join(tmpDir, "nested", "deep", "output.json")

	if err := c.ToJSONFile(data, outputPath); err != nil {
		t.Fatalf("Expected successful file write with dir creation, got error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}
}

// TestGetOutputFilename validates filename transformation
func TestGetOutputFilename(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"data.csv", "data.json"},
		{"file.txt", "file.json"},
		{"report.dat", "report.json"},
		{"path/to/data.csv", "path/to/data.json"},
		{"noext", "noext.json"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := GetOutputFilename(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// TestToJSONOrderPreservation validates row order is preserved
// ADR-003: Output order must match input order
func TestToJSONOrderPreservation(t *testing.T) {
	c := New()

	data := []map[string]string{
		{"id": "1", "name": "First"},
		{"id": "2", "name": "Second"},
		{"id": "3", "name": "Third"},
	}

	jsonBytes, err := c.ToJSON(data)
	if err != nil {
		t.Fatal(err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatal(err)
	}

	// Verify order
	if result[0]["id"] != "1" || result[0]["name"] != "First" {
		t.Error("First record order not preserved")
	}
	if result[1]["id"] != "2" || result[1]["name"] != "Second" {
		t.Error("Second record order not preserved")
	}
	if result[2]["id"] != "3" || result[2]["name"] != "Third" {
		t.Error("Third record order not preserved")
	}
}

// TestToJSONSpecialCharacters validates special character handling
func TestToJSONSpecialCharacters(t *testing.T) {
	c := New()

	data := []map[string]string{
		{"name": "John \"Doe\"", "note": "Line1\nLine2", "path": "C:\\Users\\test"},
	}

	jsonBytes, err := c.ToJSON(data)
	if err != nil {
		t.Fatalf("Expected successful conversion, got error: %v", err)
	}

	// Verify it's valid JSON (special chars properly escaped)
	var result []map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Generated invalid JSON with special chars: %v", err)
	}

	// Verify special characters preserved
	if result[0]["name"] != "John \"Doe\"" {
		t.Errorf("Quotes not preserved correctly")
	}
	if result[0]["note"] != "Line1\nLine2" {
		t.Errorf("Newlines not preserved correctly")
	}
}

// BenchmarkToJSON benchmarks JSON conversion
func BenchmarkToJSON(b *testing.B) {
	c := New()

	data := make([]map[string]string, 100)
	for i := 0; i < 100; i++ {
		data[i] = map[string]string{
			"name":  "John Doe",
			"age":   "30",
			"email": "john@example.com",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := c.ToJSON(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkToJSONFile benchmarks file writing
func BenchmarkToJSONFile(b *testing.B) {
	c := New()

	data := []map[string]string{
		{"name": "John", "age": "30"},
	}

	tmpDir, err := os.MkdirTemp("", "benchmark")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join(tmpDir, "output.json")
		if err := c.ToJSONFile(data, outputPath); err != nil {
			b.Fatal(err)
		}
		os.Remove(outputPath)
	}
}

// TestToJSONOrdered_FieldOrderPreservation verifies that JSON field order matches CSV header order per ADR-003
func TestToJSONOrdered_FieldOrderPreservation(t *testing.T) {
	// Create test data with specific column order
	result := &parser.ParseResult{
		Headers: []string{"demoID", "DemoField1", "DemoField2"},
		Rows: []parser.OrderedMap{
			{
				Keys:   []string{"demoID", "DemoField1", "DemoField2"},
				Values: map[string]string{"demoID": "5", "DemoField1": "foo", "DemoField2": "bar"},
			},
			{
				Keys:   []string{"demoID", "DemoField1", "DemoField2"},
				Values: map[string]string{"demoID": "6", "DemoField1": "fubar", "DemoField2": "fubar"},
			},
		},
	}

	c := New()
	jsonBytes, err := c.ToJSONOrdered(result)
	if err != nil {
		t.Fatalf("ToJSONOrdered failed: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Verify JSON is valid
	var decoded []map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("Generated JSON is invalid: %v", err)
	}

	// Verify field order by checking string positions
	// "demoID" should appear before "DemoField1"
	demoIDPos := strings.Index(jsonStr, `"demoID"`)
	field1Pos := strings.Index(jsonStr, `"DemoField1"`)
	field2Pos := strings.Index(jsonStr, `"DemoField2"`)

	if demoIDPos == -1 || field1Pos == -1 || field2Pos == -1 {
		t.Fatalf("Expected fields not found in JSON output")
	}

	// First object field order
	if demoIDPos > field1Pos {
		t.Errorf("Field order violation: demoID appears at position %d, should be before DemoField1 at %d", demoIDPos, field1Pos)
	}
	if field1Pos > field2Pos {
		t.Errorf("Field order violation: DemoField1 appears at position %d, should be before DemoField2 at %d", field1Pos, field2Pos)
	}

	// Verify values
	if len(decoded) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(decoded))
	}

	if decoded[0]["demoID"] != "5" || decoded[0]["DemoField1"] != "foo" || decoded[0]["DemoField2"] != "bar" {
		t.Errorf("Row 1 values incorrect: %v", decoded[0])
	}
}
