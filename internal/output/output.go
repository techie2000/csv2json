package output

import (
	"csv2json/internal/parser"
	"encoding/json"
	"fmt"
)

type Handler interface {
	Send(data []map[string]string, identifier string) error
	SendOrdered(result *parser.ParseResult, identifier string) error
	Close() error
}

type Message struct {
	Identifier string              `json:"identifier"`
	Data       []map[string]string `json:"data"`
}

func CreateHandler(outputType, outputFolder, queueType, queueHost string, queuePort int, queueName, queueUsername, queuePassword string, logMessages bool) (Handler, error) {
	switch outputType {
	case "file":
		return NewFileHandler(outputFolder), nil
	case "queue":
		return NewQueueHandler(queueType, queueHost, queuePort, queueName, queueUsername, queuePassword, logMessages)
	case "both":
		fileHandler := NewFileHandler(outputFolder)
		queueHandler, err := NewQueueHandler(queueType, queueHost, queuePort, queueName, queueUsername, queuePassword, logMessages)
		if err != nil {
			return nil, fmt.Errorf("failed to create queue handler: %w", err)
		}
		return NewBothHandler(fileHandler, queueHandler), nil
	default:
		return nil, fmt.Errorf("invalid output type: %s (valid: file, queue, both)", outputType)
	}
}

// BothHandler sends output to both file and queue
type BothHandler struct {
	fileHandler  Handler
	queueHandler Handler
}

func NewBothHandler(fileHandler, queueHandler Handler) *BothHandler {
	return &BothHandler{
		fileHandler:  fileHandler,
		queueHandler: queueHandler,
	}
}

func (h *BothHandler) Send(data []map[string]string, identifier string) error {
	// Write to file first (creates archive/audit trail)
	if err := h.fileHandler.Send(data, identifier); err != nil {
		return fmt.Errorf("file output failed: %w", err)
	}

	// Then send to queue (if file succeeds, we have archive)
	if err := h.queueHandler.Send(data, identifier); err != nil {
		return fmt.Errorf("queue output failed: %w", err)
	}

	return nil
}

func (h *BothHandler) SendOrdered(result *parser.ParseResult, identifier string) error {
	// Write to file first
	if err := h.fileHandler.SendOrdered(result, identifier); err != nil {
		return fmt.Errorf("file output failed: %w", err)
	}

	// Then send to queue
	if err := h.queueHandler.SendOrdered(result, identifier); err != nil {
		return fmt.Errorf("queue output failed: %w", err)
	}

	return nil
}

func (h *BothHandler) Close() error {
	// Close both handlers (ignore file handler close errors as it's a no-op)
	h.fileHandler.Close()
	return h.queueHandler.Close()
}

// SetEnvelopeContext configures envelope metadata for the queue handler (ADR-006)
func (h *BothHandler) SetEnvelopeContext(routeName, ingestionContract, sourceFilePath string, includeEnvelope bool) {
	if qh, ok := h.queueHandler.(*QueueHandler); ok {
		qh.SetEnvelopeContext(routeName, ingestionContract, sourceFilePath, includeEnvelope)
	}
}

func marshalMessage(data []map[string]string, identifier string) ([]byte, error) {
	msg := Message{
		Identifier: identifier,
		Data:       data,
	}
	return json.Marshal(msg)
}
