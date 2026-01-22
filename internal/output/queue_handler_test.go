package output

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMarshalMessage(t *testing.T) {
	data := []map[string]string{
		{"name": "Alice", "age": "30"},
		{"name": "Bob", "age": "25"},
	}
	identifier := "test.csv"

	message, err := marshalMessage(data, identifier)
	if err != nil {
		t.Fatalf("marshalMessage failed: %v", err)
	}

	if len(message) == 0 {
		t.Error("marshalMessage returned empty message")
	}

	// Parse back to verify structure
	var parsed map[string]interface{}
	if err := json.Unmarshal(message, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	// Check identifier
	if parsed["identifier"] != identifier {
		t.Errorf("Expected identifier '%s', got '%v'", identifier, parsed["identifier"])
	}

	// Check data array
	dataArray, ok := parsed["data"].([]interface{})
	if !ok {
		t.Fatal("data field is not an array")
	}

	if len(dataArray) != len(data) {
		t.Errorf("Expected %d records, got %d", len(data), len(dataArray))
	}

	// Verify first record
	firstRecord := dataArray[0].(map[string]interface{})
	if firstRecord["name"] != "Alice" {
		t.Errorf("Expected name 'Alice', got '%v'", firstRecord["name"])
	}
	if firstRecord["age"] != "30" {
		t.Errorf("Expected age '30', got '%v'", firstRecord["age"])
	}
}

func TestMarshalMessage_EmptyData(t *testing.T) {
	data := []map[string]string{}
	identifier := "empty.csv"

	message, err := marshalMessage(data, identifier)
	if err != nil {
		t.Fatalf("marshalMessage failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(message, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	dataArray, ok := parsed["data"].([]interface{})
	if !ok {
		t.Fatal("data field is not an array")
	}

	if len(dataArray) != 0 {
		t.Errorf("Expected empty data array, got %d records", len(dataArray))
	}
}

func TestMarshalMessage_StringValues(t *testing.T) {
	// ADR-003: All values must be strings
	data := []map[string]string{
		{"number": "123", "boolean": "true", "empty": ""},
	}
	identifier := "types.csv"

	message, err := marshalMessage(data, identifier)
	if err != nil {
		t.Fatalf("marshalMessage failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(message, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	dataArray := parsed["data"].([]interface{})
	record := dataArray[0].(map[string]interface{})

	// All values should be strings
	if _, ok := record["number"].(string); !ok {
		t.Error("number should be a string")
	}
	if _, ok := record["boolean"].(string); !ok {
		t.Error("boolean should be a string")
	}
	if _, ok := record["empty"].(string); !ok {
		t.Error("empty should be a string")
	}

	// Verify actual values
	if record["number"] != "123" {
		t.Errorf("Expected '123', got '%v'", record["number"])
	}
	if record["boolean"] != "true" {
		t.Errorf("Expected 'true', got '%v'", record["boolean"])
	}
	if record["empty"] != "" {
		t.Errorf("Expected empty string, got '%v'", record["empty"])
	}
}

func TestNewQueueHandler_UnsupportedType(t *testing.T) {
	handler, err := NewQueueHandler("invalid-type", "localhost", 5672, "test", "", "", false)

	if err == nil {
		t.Error("Expected error for unsupported queue type")
	}

	if handler != nil {
		t.Error("Handler should be nil for unsupported type")
	}

	if !strings.Contains(err.Error(), "unsupported queue type") {
		t.Errorf("Error message should mention 'unsupported queue type', got: %v", err)
	}
}

func TestNewQueueHandler_NotImplemented(t *testing.T) {
	notImplementedTypes := []string{"kafka", "sqs", "azure-servicebus"}

	for _, queueType := range notImplementedTypes {
		t.Run(queueType, func(t *testing.T) {
			handler, err := NewQueueHandler(queueType, "localhost", 5672, "test", "", "", false)

			if err == nil {
				t.Errorf("Expected error for %s (not yet implemented)", queueType)
			}

			if handler != nil {
				t.Error("Handler should be nil for not implemented type")
			}

			if !strings.Contains(err.Error(), "not yet implemented") {
				t.Errorf("Error message should mention 'not yet implemented', got: %v", err)
			}
		})
	}
}

func TestQueueHandler_MessageStructure(t *testing.T) {
	// Test the structure of marshaled messages matches expected format
	data := []map[string]string{
		{"col1": "value1", "col2": "value2"},
		{"col1": "value3", "col2": "value4"},
	}
	identifier := "test.csv"

	message, err := marshalMessage(data, identifier)
	if err != nil {
		t.Fatalf("marshalMessage failed: %v", err)
	}

	// Verify JSON structure
	var result map[string]interface{}
	if err := json.Unmarshal(message, &result); err != nil {
		t.Fatalf("Invalid JSON structure: %v", err)
	}

	// Check required fields
	if _, exists := result["identifier"]; !exists {
		t.Error("Message missing 'identifier' field")
	}

	if _, exists := result["data"]; !exists {
		t.Error("Message missing 'data' field")
	}

	// Verify data is an array
	dataArray, ok := result["data"].([]interface{})
	if !ok {
		t.Fatal("'data' field should be an array")
	}

	if len(dataArray) != 2 {
		t.Errorf("Expected 2 records in data array, got %d", len(dataArray))
	}

	// Verify each record is a map with string values
	for i, item := range dataArray {
		record, ok := item.(map[string]interface{})
		if !ok {
			t.Fatalf("Record %d is not a map", i)
		}

		for key, value := range record {
			if _, ok := value.(string); !ok {
				t.Errorf("Record %d field '%s' is not a string: %T", i, key, value)
			}
		}
	}
}

func TestQueueHandler_IdentifierValidation(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		valid      bool
	}{
		{"normal filename", "file.csv", true},
		{"with path", "path/to/file.csv", true},
		{"empty string", "", true}, // Empty is technically valid
		{"special chars", "file_@#$.csv", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []map[string]string{{"key": "value"}}

			message, err := marshalMessage(data, tt.identifier)
			if err != nil {
				if tt.valid {
					t.Errorf("Expected valid identifier, got error: %v", err)
				}
				return
			}

			if !tt.valid {
				t.Error("Expected error for invalid identifier")
			}

			var result map[string]interface{}
			json.Unmarshal(message, &result)

			if result["identifier"] != tt.identifier {
				t.Errorf("Identifier mismatch. Expected '%s', got '%v'", tt.identifier, result["identifier"])
			}
		})
	}
}

func TestQueueHandler_LargeDataset(t *testing.T) {
	// Test marshaling a large number of records
	data := make([]map[string]string, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = map[string]string{
			"id":   string(rune(i)),
			"name": "record_" + string(rune(i)),
		}
	}

	identifier := "large.csv"

	message, err := marshalMessage(data, identifier)
	if err != nil {
		t.Fatalf("marshalMessage failed for large dataset: %v", err)
	}

	if len(message) == 0 {
		t.Error("Message should not be empty for large dataset")
	}

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(message, &result); err != nil {
		t.Fatalf("Invalid JSON for large dataset: %v", err)
	}

	dataArray := result["data"].([]interface{})
	if len(dataArray) != 1000 {
		t.Errorf("Expected 1000 records, got %d", len(dataArray))
	}
}

// Benchmark tests
func BenchmarkMarshalMessage_Small(b *testing.B) {
	data := []map[string]string{
		{"col1": "value1", "col2": "value2"},
		{"col1": "value3", "col2": "value4"},
	}
	identifier := "test.csv"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		marshalMessage(data, identifier)
	}
}

func BenchmarkMarshalMessage_Large(b *testing.B) {
	data := make([]map[string]string, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = map[string]string{
			"id":   string(rune(i)),
			"name": "record_" + string(rune(i)),
			"data": "some data here",
		}
	}
	identifier := "large.csv"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		marshalMessage(data, identifier)
	}
}
