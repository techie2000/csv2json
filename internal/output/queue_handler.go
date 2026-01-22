package output

import (
	"fmt"
	"log"
	"txt2json/internal/converter"
	"txt2json/internal/parser"

	"github.com/streadway/amqp"
)

type QueueHandler struct {
	queueType   string
	conn        *amqp.Connection
	channel     *amqp.Channel
	queueName   string
	converter   *converter.Converter
	logMessages bool
}

func NewQueueHandler(queueType, host string, port int, queueName, username, password string, logMessages bool) (*QueueHandler, error) {
	handler := &QueueHandler{
		queueType:   queueType,
		queueName:   queueName,
		converter:   converter.New(),
		logMessages: logMessages,
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

func (h *QueueHandler) Send(data []map[string]string, identifier string) error {
	message, err := marshalMessage(data, identifier)
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

	// For queue output, we wrap the data in a message envelope
	// but the data itself preserves CSV column order per ADR-003
	message := []byte(fmt.Sprintf(`{"identifier":"%s","data":%s}`, identifier, string(jsonBytes)))

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
