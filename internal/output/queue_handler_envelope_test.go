package output

import (
	"encoding/json"
	"testing"
	"time"
)

// TestBuildMessageEnvelope_Structure validates the ADR-006 envelope structure
func TestBuildMessageEnvelope_Structure(t *testing.T) {
	handler := &QueueHandler{
		routeName:         "test-route",
		ingestionContract: "products.csv.v1",
		includeEnvelope:   true,
		sourceFilePath:    "/data/input/products.csv",
		queueName:         "products.inbound",
		brokerURI:         "amqp://rabbitmq:5672/",
		serviceVersion:    "test-version", // Use test version for predictable testing
	}

	data := []map[string]string{
		{"name": "Alice", "age": "30"},
		{"name": "Bob", "age": "25"},
	}

	message, err := handler.buildMessageEnvelope(data, "test-identifier")
	if err != nil {
		t.Fatalf("buildMessageEnvelope failed: %v", err)
	}

	// Unmarshal to validate structure
	var envelope MessageEnvelope
	if err := json.Unmarshal(message, &envelope); err != nil {
		t.Fatalf("Failed to unmarshal envelope: %v", err)
	}

	// Validate meta.ingestionContract
	if envelope.Meta.IngestionContract != "products.csv.v1" {
		t.Errorf("Expected ingestionContract 'products.csv.v1', got '%s'", envelope.Meta.IngestionContract)
	}

	// Validate meta.source
	if envelope.Meta.Source.Type != "file" {
		t.Errorf("Expected source.type 'file', got '%s'", envelope.Meta.Source.Type)
	}
	if envelope.Meta.Source.Name != "test-identifier" {
		t.Errorf("Expected source.name 'test-identifier', got '%s'", envelope.Meta.Source.Name)
	}
	if envelope.Meta.Source.Path != "/data/input/products.csv" {
		t.Errorf("Expected source.path '/data/input/products.csv', got '%s'", envelope.Meta.Source.Path)
	}
	if envelope.Meta.Source.Queue != "products.inbound" {
		t.Errorf("Expected source.queue 'products.inbound', got '%s'", envelope.Meta.Source.Queue)
	}
	if envelope.Meta.Source.Broker != "amqp://rabbitmq:5672/" {
		t.Errorf("Expected source.broker 'amqp://rabbitmq:5672/', got '%s'", envelope.Meta.Source.Broker)
	}
	if envelope.Meta.Source.Route != "test-route" {
		t.Errorf("Expected source.route 'test-route', got '%s'", envelope.Meta.Source.Route)
	}

	// Validate meta.ingestion
	if envelope.Meta.Ingestion.Service != "csv2json" {
		t.Errorf("Expected ingestion.service 'csv2json', got '%s'", envelope.Meta.Ingestion.Service)
	}
	if envelope.Meta.Ingestion.Version != "test-version" {
		t.Errorf("Expected ingestion.version 'test-version', got '%s'", envelope.Meta.Ingestion.Version)
	}

	// Validate timestamp is ISO8601 format
	_, err = time.Parse(time.RFC3339, envelope.Meta.Ingestion.Timestamp)
	if err != nil {
		t.Errorf("Timestamp should be RFC3339 format, got '%s': %v", envelope.Meta.Ingestion.Timestamp, err)
	}

	// Validate data payload
	if len(envelope.Data) != 2 {
		t.Errorf("Expected 2 data records, got %d", len(envelope.Data))
	}
	if envelope.Data[0]["name"] != "Alice" {
		t.Errorf("Expected first record name 'Alice', got '%s'", envelope.Data[0]["name"])
	}
}

// TestBuildMessageEnvelope_EmptyData validates envelope with empty data array
func TestBuildMessageEnvelope_EmptyData(t *testing.T) {
	handler := &QueueHandler{
		routeName:         "empty-route",
		ingestionContract: "test.csv.v1",
		includeEnvelope:   true,
		sourceFilePath:    "/data/input/empty.csv",
		queueName:         "test.queue",
		brokerURI:         "amqp://localhost:5672/",
		serviceVersion:    "test-version",
	}

	data := []map[string]string{}

	message, err := handler.buildMessageEnvelope(data, "empty-test")
	if err != nil {
		t.Fatalf("buildMessageEnvelope failed: %v", err)
	}

	var envelope MessageEnvelope
	if err := json.Unmarshal(message, &envelope); err != nil {
		t.Fatalf("Failed to unmarshal envelope: %v", err)
	}

	// Data should be empty array, not null
	if envelope.Data == nil {
		t.Error("Data should be empty array, not null")
	}
	if len(envelope.Data) != 0 {
		t.Errorf("Expected empty data array, got %d records", len(envelope.Data))
	}

	// Metadata should still be complete
	if envelope.Meta.IngestionContract == "" {
		t.Error("IngestionContract should not be empty")
	}
	if envelope.Meta.Source.Path == "" {
		t.Error("Source path should not be empty")
	}
}

// TestBuildMessageEnvelope_StringValues validates ADR-003 contract (string values only)
func TestBuildMessageEnvelope_StringValues(t *testing.T) {
	handler := &QueueHandler{
		routeName:         "test-route",
		ingestionContract: "types.csv.v1",
		includeEnvelope:   true,
		sourceFilePath:    "/data/input/types.csv",
		queueName:         "test.queue",
		brokerURI:         "amqp://localhost:5672/",
		serviceVersion:    "test-version",
	}

	data := []map[string]string{
		{"number": "123", "boolean": "true", "empty": ""},
	}

	message, err := handler.buildMessageEnvelope(data, "types-test")
	if err != nil {
		t.Fatalf("buildMessageEnvelope failed: %v", err)
	}

	var envelope MessageEnvelope
	if err := json.Unmarshal(message, &envelope); err != nil {
		t.Fatalf("Failed to unmarshal envelope: %v", err)
	}

	// Verify all values remain strings (no type coercion)
	record := envelope.Data[0]
	if record["number"] != "123" {
		t.Errorf("Number should be string '123', got '%s'", record["number"])
	}
	if record["boolean"] != "true" {
		t.Errorf("Boolean should be string 'true', got '%s'", record["boolean"])
	}
	if record["empty"] != "" {
		t.Errorf("Empty should be empty string, got '%s'", record["empty"])
	}
}

// TestBuildMessageEnvelope_SourceFilename extracts correct filename
func TestBuildMessageEnvelope_SourceFilename(t *testing.T) {
	testCases := []struct {
		name         string
		fullPath     string
		expectedName string
	}{
		{"unix absolute path", "/data/input/products.csv", "products.csv"},
		{"windows path", "C:\\data\\input\\orders.csv", "orders.csv"},
		{"relative path", "input/accounts.csv", "accounts.csv"},
		{"filename only", "test.csv", "test.csv"},
		{"nested path", "/a/b/c/d/file.csv", "file.csv"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := &QueueHandler{
				routeName:         "test-route",
				ingestionContract: "test.csv.v1",
				includeEnvelope:   true,
				sourceFilePath:    tc.fullPath,
				queueName:         "test.queue",
				brokerURI:         "amqp://localhost:5672/",
				serviceVersion:    "test-version",
			}

			message, err := handler.buildMessageEnvelope([]map[string]string{}, tc.expectedName)
			if err != nil {
				t.Fatalf("buildMessageEnvelope failed: %v", err)
			}

			var envelope MessageEnvelope
			if err := json.Unmarshal(message, &envelope); err != nil {
				t.Fatalf("Failed to unmarshal envelope: %v", err)
			}

			if envelope.Meta.Source.Name != tc.expectedName {
				t.Errorf("Expected filename '%s', got '%s'", tc.expectedName, envelope.Meta.Source.Name)
			}
		})
	}
}

// TestBuildMessageEnvelope_TimestampRecent validates timestamp is recent (within 1 second)
func TestBuildMessageEnvelope_TimestampRecent(t *testing.T) {
	handler := &QueueHandler{
		routeName:         "test-route",
		ingestionContract: "test.csv.v1",
		includeEnvelope:   true,
		sourceFilePath:    "/data/input/test.csv",
		queueName:         "test.queue",
		brokerURI:         "amqp://localhost:5672/",
		serviceVersion:    "test-version",
	}

	before := time.Now().UTC()
	message, err := handler.buildMessageEnvelope([]map[string]string{}, "timestamp-test")
	after := time.Now().UTC()

	if err != nil {
		t.Fatalf("buildMessageEnvelope failed: %v", err)
	}

	var envelope MessageEnvelope
	if err := json.Unmarshal(message, &envelope); err != nil {
		t.Fatalf("Failed to unmarshal envelope: %v", err)
	}

	timestamp, err := time.Parse(time.RFC3339, envelope.Meta.Ingestion.Timestamp)
	if err != nil {
		t.Fatalf("Failed to parse timestamp: %v", err)
	}

	// Timestamp should be recent (within test execution time + 1 second buffer)
	if timestamp.Before(before.Add(-1*time.Second)) || timestamp.After(after.Add(1*time.Second)) {
		t.Errorf("Timestamp %s outside reasonable range %s to %s", timestamp, before, after)
	}
}

// TestSetEnvelopeContext validates envelope context configuration
func TestSetEnvelopeContext(t *testing.T) {
	handler := &QueueHandler{}

	handler.SetEnvelopeContext(
		"test-route",
		"products.csv.v2",
		"/data/input/products.csv",
		true,
	)

	if handler.routeName != "test-route" {
		t.Errorf("Expected routeName 'test-route', got '%s'", handler.routeName)
	}
	if handler.ingestionContract != "products.csv.v2" {
		t.Errorf("Expected ingestionContract 'products.csv.v2', got '%s'", handler.ingestionContract)
	}
	if handler.sourceFilePath != "/data/input/products.csv" {
		t.Errorf("Expected sourceFilePath '/data/input/products.csv', got '%s'", handler.sourceFilePath)
	}
	if !handler.includeEnvelope {
		t.Error("Expected includeEnvelope true")
	}
}

// BenchmarkBuildMessageEnvelope measures envelope marshaling overhead
func BenchmarkBuildMessageEnvelope(b *testing.B) {
	handler := &QueueHandler{
		routeName:         "benchmark-route",
		ingestionContract: "products.csv.v1",
		includeEnvelope:   true,
		sourceFilePath:    "/data/input/products.csv",
		queueName:         "products.inbound",
		brokerURI:         "amqp://rabbitmq:5672/",
		serviceVersion:    "test-version",
	}

	data := []map[string]string{
		{"id": "1", "name": "Product A", "price": "10.99", "category": "Electronics"},
		{"id": "2", "name": "Product B", "price": "25.50", "category": "Furniture"},
		{"id": "3", "name": "Product C", "price": "5.00", "category": "Office"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.buildMessageEnvelope(data, "benchmark-test")
		if err != nil {
			b.Fatalf("buildMessageEnvelope failed: %v", err)
		}
	}
}

// BenchmarkBuildMessageEnvelope_LargePayload measures overhead with 100 records
func BenchmarkBuildMessageEnvelope_LargePayload(b *testing.B) {
	handler := &QueueHandler{
		routeName:         "benchmark-route",
		ingestionContract: "products.csv.v1",
		includeEnvelope:   true,
		sourceFilePath:    "/data/input/products.csv",
		queueName:         "products.inbound",
		brokerURI:         "amqp://rabbitmq:5672/",
		serviceVersion:    "test-version",
	}

	// Generate 100 records
	data := make([]map[string]string, 100)
	for i := 0; i < 100; i++ {
		data[i] = map[string]string{
			"id":       string(rune(i)),
			"name":     "Product " + string(rune(i)),
			"price":    "10.99",
			"category": "Test",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.buildMessageEnvelope(data, "large-payload-test")
		if err != nil {
			b.Fatalf("buildMessageEnvelope failed: %v", err)
		}
	}
}
