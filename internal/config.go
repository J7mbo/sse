package internal

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

const (
	// Env - public so the server can run https if on prod.
	ProdEnv = "prod"
)

type Config struct {
	AppEnv            string `envconfig:"SSE_APP_ENV" default:"dev"`
	LogIndex          string `envconfig:"SSE_LOG_INDEX" default:"sse"`
	LogsDir           string `envconfig:"SSE_LOGS_DIR" default:"infrastructure/logs/"`
	CertsDir          string `envconfig:"SSE_CERTS_DIR" default:"infrastructure/certs/"`
	RabbitMQConsumers int    `envconfig:"SSE_RABBITMQ_CONSUMERS" default:"50"`

	// Server port.
	Port string `envconfig:"SSE_PORT" default:"8001"`

	// Database configuration.
	PostgresHost     string `envconfig:"SSE_POSTGRES_HOST" default:"localhost"`
	PostgresPort     int    `envconfig:"SSE_POSTGRES_PORT" default:"5432"`
	PostgresUsername string `envconfig:"SSE_POSTGRES_USERNAME" default:"postgres"`
	PostgresPassword string `envconfig:"SSE_POSTGRES_PASSWORD" default:"postgres"`
	PostgresDBName   string `envconfig:"SSE_POSTGRES_DBNAME" default:"sse"`

	// RabbitMQ configuration.
	RabbitMQHost      string `envconfig:"SSE_RABBITMQ_HOST" default:"localhost"`
	RabbitMQPort      int    `envconfig:"SSE_RABBITMQ_PORT" default:"5671"`
	RabbitMQUsername  string `envconfig:"SSE_RABBITMQ_USERNAME" default:"guest"`
	RabbitMQPassword  string `envconfig:"SSE_RABBITMQ_PASSWORD" default:"guest"`
	RabbitMQQueueName string `envconfig:"SSE_RABBITMQ_QUEUE_NAME" default:"messages"`
}

func NewConfigFromEnvironmentVariables() (*Config, error) {
	var c Config
	if err := envconfig.Process("app", &c); err != nil {
		return nil, fmt.Errorf("unable to process environment variables for config, err: %w", err)
	}

	return &c, nil
}
