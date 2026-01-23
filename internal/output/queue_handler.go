package output

import (
	"csv2json/internal/converter"
	"csv2json/internal/parser"
	"csv2json/internal/version"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

// MessageEnvelope represents the ADR-006 message envelope with full provenance
type MessageEnvelope struct {
	Meta MessageMeta         `json:"meta"`
	Data []map[string]string `json:"data"`
}

// MessageMeta contains provenance and ingestion metadata
type MessageMeta struct {
	IngestionContract string            `json:"ingestionContract"`
	Source            SourceMetadata    `json:"source"`
	Ingestion         IngestionMetadata `json:"ingestion"`
}

// SourceMetadata tracks message origin and routing
type SourceMetadata struct {
	Type   string `json:"type"`             // "file", "api", "stream"
	Name   string `json:"name"`             // Original source filename
	Path   string `json:"path"`             // Full source file path
	Queue  string `json:"queue,omitempty"`  // Queue name (for queue output)
	Broker string `json:"broker,omitempty"` // Broker URI
	Route  string `json:"route"`            // Route name from configuration
}

// IngestionMetadata tracks service and timing information
type IngestionMetadata struct {
	Service   string `json:"service"`   // Service name (csv2json)
	Version   string `json:"version"`   // Service semantic version
	Timestamp string `json:"timestamp"` // ISO8601 ingestion timestamp (UTC)
}

type QueueHandler struct {
	queueType         string
	conn              *amqp.Connection
	channel           *amqp.Channel
	queueName         string
	converter         *converter.Converter
	logMessages       bool
	routeName         string // Route name for context in messages
	ingestionContract string // Schema/contract identifier
	includeEnvelope   bool   // Whether to include full envelope (ADR-006)
	sourceFilePath    string // Full source file path
	brokerURI         string // Broker connection string
	serviceVersion    string // csv2json version
}

func NewQueueHandler(queueType, host string, port int, queueName, username, password string, logMessages bool) (*QueueHandler, error) {
	// Build broker URI
	var brokerURI string
	if username != "" && password != "" {
		brokerURI = fmt.Sprintf("%s://%s:%s@%s:%d/", queueType, username, "***", host, port) // Redacted password in URI
	} else {
		brokerURI = fmt.Sprintf("%s://%s:%d/", queueType, host, port)
	}

	handler := &QueueHandler{
		queueType:       queueType,
		queueName:       queueName,
		converter:       converter.New(),
		logMessages:     logMessages,
		includeEnvelope: true, // Default: include envelope with provenance (ADR-006)
		brokerURI:       brokerURI,
		serviceVersion:  version.GetVersion(), // Read from VERSION file (ADR-006)
	}

	// Route to appropriate queue implementation
	switch queueType {
	case "rabbitmq":
		return handler, handler.initRabbitMQ(host, port, username, password)
	case "kafka":
		return nil, fmt.Errorf("Kafka not yet implemented")
	case "sqs":
		return nil, fmt.Errorf("AWS SQS not yet implemented")
	case "azure-servicebus":
		return nil, fmt.Errorf("Azure Service Bus not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported queue type: %s", queueType)
	}
}

func (h *QueueHandler) initRabbitMQ(host string, port int, username, password string) error {
	// Build AMQP connection string
	var connStr string
	if username != "" && password != "" {
		// With authentication
		connStr = fmt.Sprintf("amqp://%s:%s@%s:%d/", username, password, host, port)
	} else {
		// Without authentication (guest:guest default)
		connStr = fmt.Sprintf("amqp://%s:%d/", host, port)
	}

	// Connect to RabbitMQ
	conn, err := amqp.Dial(connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	h.conn = conn

	// Create channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}
	h.channel = ch

	// Declare queue
	_, err = ch.QueueDeclare(
		h.queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	return nil
}

// SetEnvelopeContext configures message envelope metadata (ADR-006)
func (h *QueueHandler) SetEnvelopeContext(routeName, ingestionContract, sourceFilePath string, includeEnvelope bool) {
	h.routeName = routeName
	h.ingestionContract = ingestionContract
	h.sourceFilePath = sourceFilePath
	h.includeEnvelope = includeEnvelope
}

// buildMessageEnvelope creates ADR-006 compliant message envelope with full provenance
func (h *QueueHandler) buildMessageEnvelope(data []map[string]string, identifier string) ([]byte, error) {
	if !h.includeEnvelope {
		// Legacy format without envelope
		return marshalMessage(data, identifier)
	}

	// Build full message envelope with provenance metadata (ADR-006)
	envelope := MessageEnvelope{
		Meta: MessageMeta{
			IngestionContract: h.ingestionContract,
			Source: SourceMetadata{
				Type:   "file",
				Name:   identifier,
				Path:   h.sourceFilePath,
				Queue:  h.queueName,
				Broker: h.brokerURI,
				Route:  h.routeName,
			},
			Ingestion: IngestionMetadata{
				Service:   "csv2json",
				Version:   h.serviceVersion,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		},
		Data: data,
	}

	return json.Marshal(envelope)
}

func (h *QueueHandler) Send(data []map[string]string, identifier string) error {
	message, err := h.buildMessageEnvelope(data, identifier)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	switch h.queueType {
	case "rabbitmq":
		return h.sendToRabbitMQ(message)
	default:
		return fmt.Errorf("unsupported queue type: %s", h.queueType)
	}
}

func (h *QueueHandler) SendOrdered(result *parser.ParseResult, identifier string) error {
	// Convert to ordered JSON
	jsonBytes, err := h.converter.ToJSONOrdered(result)
	if err != nil {
		return fmt.Errorf("failed to marshal ordered JSON: %w", err)
	}

	// Parse JSON bytes back to []map[string]string for envelope
	var data []map[string]string
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return fmt.Errorf("failed to unmarshal JSON for envelope: %w", err)
	}

	// Build envelope with provenance metadata
	message, err := h.buildMessageEnvelope(data, identifier)
	if err != nil {
		return fmt.Errorf("failed to build message envelope: %w", err)
	}

	switch h.queueType {
	case "rabbitmq":
		return h.sendToRabbitMQ(message)
	default:
		return fmt.Errorf("unsupported queue type: %s", h.queueType)
	}
}

func (h *QueueHandler) sendToRabbitMQ(message []byte) error {
	if h.logMessages {
		log.Printf("Queuing message to %s: %s", h.queueName, string(message))
	}

	err := h.channel.Publish(
		"",          // exchange
		h.queueName, // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         message,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (h *QueueHandler) Close() error {
	if h.channel != nil {
		h.channel.Close()
	}
	if h.conn != nil {
		return h.conn.Close()
	}
	return nil
}
