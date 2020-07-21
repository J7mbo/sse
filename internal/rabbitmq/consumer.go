package rabbitmq

import (
	"fmt"

	"github.com/streadway/amqp"

	"sse/internal"
	"sse/internal/rabbitmq/handler"
)

type amqpConnection interface {
	Channel() (*amqp.Channel, error)
	Close() error
	IsClosed() bool
}

type warningDebugLogger interface {
	Warning(message string, fields ...map[string]interface{})
	Debug(message string, fields ...map[string]interface{})
}

type Consumer struct {
	conn     amqpConnection
	conf     *internal.Config
	handlers *handler.Map
	logger   warningDebugLogger
}

func NewConsumer(
	conn amqpConnection, conf *internal.Config, handlers *handler.Map, lgr warningDebugLogger,
) (*Consumer, error) {
	if conn.IsClosed() {
		return nil, fmt.Errorf("invalid argument - rabbitmq consumer must be connected when passed to constructor")
	}

	return &Consumer{conn: conn, conf: conf, handlers: handlers, logger: lgr}, nil
}

//
// Consume is blocking. You should call this in a goroutine.
//
func (c *Consumer) Consume() error {
	amqpChannel, err := c.conn.Channel()
	if err != nil {
		return err
	}

	defer func() {
		_ = amqpChannel.Close()
		_ = c.conn.Close()
	}()

	queue, err := amqpChannel.QueueDeclare(c.conf.RabbitMQQueueName, true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = amqpChannel.Qos(1, 0, false)
	if err != nil {
		return err
	}

	messageChannel, err := amqpChannel.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// Attempt to consumer forever.
	for {
		select {
		case m := <-messageChannel:
			h, err := c.handlers.FindForMessage(&m)
			if err != nil {
				c.logger.Warning(fmt.Sprintf("could not find handler for message: %s", string(m.Body)))
				continue
			}

			c.logger.Debug(fmt.Sprintf("message received for: %s", h.HandlesMessageType()))
			h.Handle(&m)
		}
	}
}

func (c *Consumer) Close() {
	_ = c.conn.Close()
}
