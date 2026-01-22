package output

import (
	"encoding/json"
	"fmt"
	"txt2json/internal/parser"
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
	default:
		return nil, fmt.Errorf("invalid output type: %s", outputType)
	}
}

func marshalMessage(data []map[string]string, identifier string) ([]byte, error) {
	msg := Message{
		Identifier: identifier,
		Data:       data,
	}
	return json.Marshal(msg)
}
