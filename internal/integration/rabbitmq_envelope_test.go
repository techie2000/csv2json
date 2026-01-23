package integration

import (
	"csv2json/internal/output"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/streadway/amqp"
)

// TestRabbitMQIntegration_EnvelopeStructure validates full envelope with real RabbitMQ
// Run with: docker compose up -d && go test -v ./internal/integration -run TestRabbitMQIntegration
func TestRabbitMQIntegration_EnvelopeStructure(t *testing.T) {
	// Skip if RabbitMQ not available (CI environments)
	if os.Getenv("SKIP_INTEGRATION") == "true" {
		t.Skip("Skipping integration test (SKIP_INTEGRATION=true)")
	}

	// Connect to RabbitMQ (assumes docker compose is running)
	rabbitHost := os.Getenv("QUEUE_HOST")
	if rabbitHost == "" {
		rabbitHost = "localhost"
	}

	// Create queue handler
	handler, err := output.NewQueueHandler("rabbitmq", rabbitHost, 5672, "integration-test-queue", "", "", false)
	if err != nil {
		t.Skipf("Cannot connect to RabbitMQ (is docker compose running?): %v", err)
	}
	defer handler.Close()

	// Configure envelope context
	handler.SetEnvelopeContext(
		"integration-test-route",
		"integration.csv.v1",
		"/data/input/integration-test.csv",
		true,
	)

	// Test data
	data := []map[string]string{
		{"id": "1", "name": "Integration Test", "status": "active"},
		{"id": "2", "name": "Envelope Test", "status": "complete"},
	}

	// Send message
	err = handler.Send(data, "integration-test.csv")
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Connect consumer to verify message structure
	connStr := "amqp://guest:guest@" + rabbitHost + ":5672/"
	conn, err := amqp.Dial(connStr)
	if err != nil {
		t.Fatalf("Failed to connect consumer: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("Failed to open channel: %v", err)
	}
	defer ch.Close()

	// Consume message
	msgs, err := ch.Consume(
		"integration-test-queue",
		"",    // consumer tag
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		t.Fatalf("Failed to consume: %v", err)
	}

	// Wait for message (timeout after 5 seconds)
	select {
	case msg := <-msgs:
		// Unmarshal and validate envelope structure
		var envelope struct {
			Meta struct {
				IngestionContract string `json:"ingestionContract"`
				Source            struct {
					Type   string `json:"type"`
					Name   string `json:"name"`
					Path   string `json:"path"`
					Queue  string `json:"queue"`
					Broker string `json:"broker"`
					Route  string `json:"route"`
				} `json:"source"`
				Ingestion struct {
					Service   string `json:"service"`
					Version   string `json:"version"`
					Timestamp string `json:"timestamp"`
				} `json:"ingestion"`
			} `json:"meta"`
			Data []map[string]string `json:"data"`
		}

		if err := json.Unmarshal(msg.Body, &envelope); err != nil {
			t.Fatalf("Failed to unmarshal message: %v", err)
		}

		// Validate envelope fields (ADR-006 compliance)
		t.Run("IngestionContract", func(t *testing.T) {
			if envelope.Meta.IngestionContract != "integration.csv.v1" {
				t.Errorf("Expected contract 'integration.csv.v1', got '%s'", envelope.Meta.IngestionContract)
			}
		})

		t.Run("SourceType", func(t *testing.T) {
			if envelope.Meta.Source.Type != "file" {
				t.Errorf("Expected type 'file', got '%s'", envelope.Meta.Source.Type)
			}
		})

		t.Run("SourceName", func(t *testing.T) {
			if envelope.Meta.Source.Name != "integration-test.csv" {
				t.Errorf("Expected name 'integration-test.csv', got '%s'", envelope.Meta.Source.Name)
			}
		})

		t.Run("SourcePath", func(t *testing.T) {
			if envelope.Meta.Source.Path != "/data/input/integration-test.csv" {
				t.Errorf("Expected path '/data/input/integration-test.csv', got '%s'", envelope.Meta.Source.Path)
			}
		})

		t.Run("SourceQueue", func(t *testing.T) {
			if envelope.Meta.Source.Queue != "integration-test-queue" {
				t.Errorf("Expected queue 'integration-test-queue', got '%s'", envelope.Meta.Source.Queue)
			}
		})

		t.Run("SourceBroker", func(t *testing.T) {
			// Broker URI should contain protocol and host
			if envelope.Meta.Source.Broker == "" {
				t.Error("Broker URI should not be empty")
			}
			// Should be amqp:// protocol
			if len(envelope.Meta.Source.Broker) < 7 || envelope.Meta.Source.Broker[:7] != "amqp://" {
				t.Errorf("Broker should start with 'amqp://', got '%s'", envelope.Meta.Source.Broker)
			}
		})

		t.Run("SourceRoute", func(t *testing.T) {
			if envelope.Meta.Source.Route != "integration-test-route" {
				t.Errorf("Expected route 'integration-test-route', got '%s'", envelope.Meta.Source.Route)
			}
		})

		t.Run("IngestionService", func(t *testing.T) {
			if envelope.Meta.Ingestion.Service != "csv2json" {
				t.Errorf("Expected service 'csv2json', got '%s'", envelope.Meta.Ingestion.Service)
			}
		})

		t.Run("IngestionVersion", func(t *testing.T) {
			if envelope.Meta.Ingestion.Version == "" {
				t.Error("Version should not be empty")
			}
		})

		t.Run("IngestionTimestamp", func(t *testing.T) {
			// Validate RFC3339 format
			ts, err := time.Parse(time.RFC3339, envelope.Meta.Ingestion.Timestamp)
			if err != nil {
				t.Errorf("Timestamp should be RFC3339, got '%s': %v", envelope.Meta.Ingestion.Timestamp, err)
			}
			// Should be recent (within last minute)
			age := time.Since(ts)
			if age > time.Minute {
				t.Errorf("Timestamp too old: %v", age)
			}
		})

		t.Run("DataPayload", func(t *testing.T) {
			if len(envelope.Data) != 2 {
				t.Errorf("Expected 2 records, got %d", len(envelope.Data))
			}
			if len(envelope.Data) > 0 && envelope.Data[0]["name"] != "Integration Test" {
				t.Errorf("Expected first record name 'Integration Test', got '%s'", envelope.Data[0]["name"])
			}
		})

	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

// TestRabbitMQIntegration_MultipleMessages validates envelope uniqueness
func TestRabbitMQIntegration_MultipleMessages(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION") == "true" {
		t.Skip("Skipping integration test (SKIP_INTEGRATION=true)")
	}

	rabbitHost := os.Getenv("QUEUE_HOST")
	if rabbitHost == "" {
		rabbitHost = "localhost"
	}

	handler, err := output.NewQueueHandler("rabbitmq", rabbitHost, 5672, "multi-message-test-queue", "", "", false)
	if err != nil {
		t.Skipf("Cannot connect to RabbitMQ: %v", err)
	}
	defer handler.Close()

	// Send 3 messages with different contracts
	contracts := []string{"test1.csv.v1", "test2.csv.v1", "test3.csv.v1"}
	for i, contract := range contracts {
		handler.SetEnvelopeContext(
			"multi-test-route",
			contract,
			"/data/input/test"+string(rune(i))+".csv",
			true,
		)

		data := []map[string]string{
			{"message": string(rune(i)), "contract": contract},
		}

		err = handler.Send(data, "test"+string(rune(i))+".csv")
		if err != nil {
			t.Fatalf("Failed to send message %d: %v", i, err)
		}
	}

	// Verify all 3 messages have correct contracts
	connStr := "amqp://guest:guest@" + rabbitHost + ":5672/"
	conn, err := amqp.Dial(connStr)
	if err != nil {
		t.Fatalf("Failed to connect consumer: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("Failed to open channel: %v", err)
	}
	defer ch.Close()

	msgs, err := ch.Consume("multi-message-test-queue", "", true, false, false, false, nil)
	if err != nil {
		t.Fatalf("Failed to consume: %v", err)
	}

	receivedContracts := make(map[string]bool)
	timeout := time.After(5 * time.Second)
	messageCount := 0

	for messageCount < 3 {
		select {
		case msg := <-msgs:
			var envelope struct {
				Meta struct {
					IngestionContract string `json:"ingestionContract"`
				} `json:"meta"`
			}
			if err := json.Unmarshal(msg.Body, &envelope); err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}
			receivedContracts[envelope.Meta.IngestionContract] = true
			messageCount++

		case <-timeout:
			t.Fatalf("Timeout: received %d/3 messages", messageCount)
		}
	}

	// Verify all 3 contracts received
	for _, contract := range contracts {
		if !receivedContracts[contract] {
			t.Errorf("Contract '%s' not received", contract)
		}
	}
}
