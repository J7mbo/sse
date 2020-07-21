package main

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/j7mbo/MethodCallRetrier/v2"
	"github.com/mroth/sseserver"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"

	"sse/internal"
	"sse/internal/logger"
	"sse/internal/rabbitmq"
	rmqHandler "sse/internal/rabbitmq/handler"
)

const (
	// SSE (webserver) certs.
	sseCertName = "cert.pem"
	sseKeyName  = "key.pem"
	sseCaName   = "ca.pem"

	// RabbitMQ certs.
	rabbitMQCertName = "server_certificate.pem"
	rabbitMQKeyName  = "server_key.pem"
	rabbitMQCaName   = "ca_certificate.pem"

	// Postgres ca.
	postgresCaName = "postgres_ca.crt"
)

func main() {
	// Configuration.
	cfg, err := internal.NewConfigFromEnvironmentVariables()
	if err != nil {
		panic(err)
	}

	// Logger.
	lgr, err := createLogger(cfg)
	if err != nil {
		panic(err)
	}

	// Database.
	dbConn, err := connectToPostgres(cfg, lgr)
	if err != nil {
		lgr.Error("unable to form initial connection to postgres, error: " + err.Error())
		panic("unable to form initial connection to postgres, error: " + err.Error())
	}

	// RabbitMQ connection.
	rmqConn, err := connectToRabbitMQ(cfg, lgr)
	if err != nil {
		lgr.Error("unable to form initial connection to rabbitmq, error: " + err.Error())
		panic("unable to form initial connection to rabbitmq, error: " + err.Error())
	}

	// Dependencies.
	sseServer := sseserver.NewServer()
	userPool := internal.NewUserPool()
	authenticator := internal.NewAuthenticator(internal.NewDB(dbConn), internal.NewRequestParser(), userPool, lgr)
	broadcaster := internal.NewBroadcaster(sseServer, userPool)

	// RabbitMQ consumed message handlers.
	msgHandlers := rmqHandler.NewMap(
		rmqHandler.NewFoundContract(broadcaster, lgr),
	)

	// RabbitMQ consumer - initial connection test.
	conn, err := rabbitmq.NewConsumer(rmqConn, cfg, msgHandlers, lgr)
	if err != nil {
		panic(err)
	}
	conn.Close()

	// RabbitMQ consumers, concurrently and in parallel. Horizontally scaled to maxConsumers.
	go consumeFromRabbitMQ(cfg, msgHandlers, lgr)

	// Routing.
	router := mux.NewRouter()
	router.Handle(
		"/subscribe/{topic}/{userID}/{nonce}",
		authenticator.AuthenticateMiddleware(sseserver.ProxyRemoteAddrHandler(sseServer)),
	)

	// Server.
	cert, err := tls.LoadX509KeyPair(cfg.CertsDir+sseCertName, cfg.CertsDir+sseKeyName)
	if err != nil {
		lgr.Error(err.Error())
		panic(err)
	}

	caCert, err := ioutil.ReadFile(cfg.CertsDir + sseCaName)
	if err != nil {
		lgr.Error(err.Error())
		panic(err)
	}

	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM(caCert) {
		lgr.Error("unable to AppendCertsFromPEM for ca certificate in main.go")
		panic("unable to AppendCertsFromPEM for ca certificate in main.go")
	}

	server := http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      rootCAs,
		},
	}

	lgr.Info("Starting HTTPS server...")
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		lgr.Error(err.Error())
		panic(err)
	}
}

func consumeFromRabbitMQ(cfg *internal.Config, msgHandlers *rmqHandler.Map, lgr logger.Logger) {
	const retryTime = 10 * time.Second
	consumerPool := make(chan int, cfg.RabbitMQConsumers)

	for {
		// No need to check a ctx.Done(). To stop consuming, we'll just kill the application itself.
		select {
		// When we can add one into the consumerPool bucket...
		case consumerPool <- 1:
			// Spin up a goroutine which consumes from RabbitMQ.
			go func() {
				for {
					rmqConn, err := connectToRabbitMQ(cfg, lgr)
					if err != nil {
						lgr.Error("Unable to connect to RabbitMQ, error: " + err.Error())

						// Give us a chance to get RabbitMQ back up before spamming the logs again...
						time.Sleep(retryTime)

						continue
					}

					// This should block. When it unblocks, there's been an error. Try and reconnect.
					consumer, err := rabbitmq.NewConsumer(rmqConn, cfg, msgHandlers, lgr)
					if err != nil {
						lgr.Error("Consumer lost connection to RabbitMQ, error: " + err.Error())

						// Give us a chance to get RabbitMQ back up before spamming the logs again...
						time.Sleep(10 * time.Second)

						continue
					}

					if err := consumer.Consume(); err != nil {
						lgr.Error(fmt.Sprintf("Error consuming from RabbitMQ: %s", err.Error()))
						continue
					}
				}
			}()
		}
	}
}

func connectToRabbitMQ(c *internal.Config, lgr logger.Logger) (*amqp.Connection, error) {
	// Load cert + key.
	cert, err := tls.LoadX509KeyPair(c.CertsDir+rabbitMQCertName, c.CertsDir+rabbitMQKeyName)
	if err != nil {
		return nil, err
	}

	// Load CA.
	caCert, err := ioutil.ReadFile(c.CertsDir + rabbitMQCaName)
	if err != nil {
		return nil, err
	}

	rootCAs := x509.NewCertPool()
	rootCAs.AppendCertsFromPEM(caCert)

	retrier := MethodCallRetrier.New(5*time.Second, 12, 1)

	var rmqConn *amqp.Connection
	errs, wasSuccessful := retrier.ExecuteFuncWithRetry(func() error {
		rc := fmt.Sprintf("amqps://%s:%s@%s:%d", c.RabbitMQUsername, c.RabbitMQPassword, c.RabbitMQHost, c.RabbitMQPort)
		rmqConn, err = amqp.DialTLS(rc, &tls.Config{Certificates: []tls.Certificate{cert}, RootCAs: rootCAs})
		if err != nil {
			lgr.Warning(fmt.Sprintf("retrying connection to rabbitmq at: %s during initial startup...", rc))
			return fmt.Errorf("unable to connect to rabbitmq at %s, err: %w", rc, err)
		}

		return nil
	})
	if !wasSuccessful {
		return nil, errs[0]
	}

	return rmqConn, nil
}

func connectToPostgres(conf *internal.Config, lgr logger.Logger) (*sql.DB, error) {
	retrier := MethodCallRetrier.New(3*time.Second, 10, 1)

	var dbConn *sql.DB
	errs, wasSuccessful := retrier.ExecuteFuncWithRetry(func() error {
		var err error
		dbConn, err = sql.Open("postgres", fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslrootcert=%s sslmode=verify-full",
			conf.PostgresHost,
			conf.PostgresPort,
			conf.PostgresUsername,
			conf.PostgresPassword,
			conf.PostgresDBName,
			conf.CertsDir+postgresCaName,
		))
		if err != nil {
			lgr.Warning("retrying connection to postgres during initial startup...")
			return fmt.Errorf("unable to connect to postgres db, err: %w", err)
		}

		return nil
	})
	if !wasSuccessful {
		return nil, errs[0]
	}

	return dbConn, nil
}

func createLogger(cfg *internal.Config) (logger.Logger, error) {
	const (
		fileMode   = os.O_APPEND | os.O_CREATE | os.O_RDWR
		filePerms  = 0644
		dateFormat = "02-Jan-2006"
	)

	dateStr := time.Now().Format(dateFormat)
	logFilePath := fmt.Sprintf("%s%s.log", cfg.LogsDir, dateStr)

	fh, err := os.OpenFile(logFilePath, fileMode, filePerms)
	if err != nil {
		return nil, fmt.Errorf("unable to create / open log file: %s, err: %w", logFilePath, err)
	}

	logrusLogger := logrus.Logger{
		Out:       os.Stderr,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level: func() logrus.Level {
			// If Prod, we don't want debug logs.
			if cfg.AppEnv == internal.ProdEnv {
				return logrus.InfoLevel
			}
			return logrus.DebugLevel
		}(),
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}

	return logger.New(fh, &logrusLogger, &logger.RequiredLogFields{Env: cfg.AppEnv, Index: cfg.LogIndex}), nil
}
