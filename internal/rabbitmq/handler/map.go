package handler

import (
	"encoding/json"
	"fmt"

	"github.com/streadway/amqp"

	"sse/internal/rabbitmq/message"
)

type Handler interface {
	Handle(delivery *amqp.Delivery)
	HandlesMessageType() string
}

type Map struct {
	handlers []Handler
}

func NewMap(handlers ...Handler) *Map {
	return &Map{handlers: handlers}
}

func (m *Map) FindForMessage(msg *amqp.Delivery) (Handler, error) {
	msgType, err := m.getMessageType(msg)
	if err != nil {
		return nil, err
	}

	for _, handler := range m.handlers {
		if handler.HandlesMessageType() == msgType {
			return handler, nil
		}
	}

	return nil, fmt.Errorf("got message of unknown type: %s, check if this is mapped in map.go", msgType)
}

func (m *Map) getMessageType(msg *amqp.Delivery) (string, error) {
	messageType := message.Type{}

	err := json.Unmarshal(msg.Body, &messageType)
	if err != nil {
		return "", fmt.Errorf("unable to get message type from message, message: %s, error: %w", msg.Body, err)
	}

	if messageType.MessageType == "" {
		return "", fmt.Errorf("unable to get message type from message, empty type. Message: %s", msg.Body)
	}

	return messageType.MessageType, nil
}
