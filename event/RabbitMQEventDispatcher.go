package event

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
)

type queueCommand struct {
	token        string
	topic        string
	corelationID string
	payload      interface{}
}

type retryCommand struct {
	retryCount int
	command    *queueCommand
}

// RabbitMQEventDispatcher is an event dispatcher that sends event to the RabbitMQ Exchange
type RabbitMQEventDispatcher struct {
	logger                 *zerolog.Logger
	exchangeName           string
	connection             *amqp.Connection
	channel                *amqp.Channel
	sendChannel            chan *queueCommand
	retryChannel           chan *retryCommand
	connectionCloseChannel chan *amqp.Error
	connectionMutex        sync.Mutex
}

// NewRabbitMQEventDispatcher create and returns a new RabbitMQEventDispatcher
func NewRabbitMQEventDispatcher(logger *zerolog.Logger) (*RabbitMQEventDispatcher, error) {
	sendChannel := make(chan *queueCommand, 200)
	retryChannel := make(chan *retryCommand, 200)
	connectionCloseChannel := make(chan *amqp.Error)

	ctxLogger := logger.With().Str("module", "RabbitMQEventDispatcher").Logger()

	dispatcher := &RabbitMQEventDispatcher{
		logger:                 &ctxLogger,
		exchangeName:           "isla_exchange",
		sendChannel:            sendChannel,
		retryChannel:           retryChannel,
		connectionCloseChannel: connectionCloseChannel,
	}

	go dispatcher.rabbitConnector()
	go dispatcher.start()

	connectionCloseChannel <- amqp.ErrClosed // Trigger the connection

	return dispatcher, nil
}

// DispatchEvent dispatches events to the message queue
func (eventDispatcher *RabbitMQEventDispatcher) DispatchEvent(token string, corelationID string, topic string, payload interface{}) {
	eventDispatcher.sendChannel <- &queueCommand{token: token, topic: topic, payload: payload}
}

func (eventDispatcher *RabbitMQEventDispatcher) start() {

	for {
		var command *queueCommand
		var retryCount int

		// Ensure that connection process is not going on
		eventDispatcher.connectionMutex.Lock()
		eventDispatcher.connectionMutex.Unlock()

		select {
		case commandFromSendChannel := <-eventDispatcher.sendChannel:
			command = commandFromSendChannel
		case commandFromRetryChannel := <-eventDispatcher.retryChannel:
			command = commandFromRetryChannel.command
			retryCount = commandFromRetryChannel.retryCount
		}

		routingKey := strings.ReplaceAll(command.topic, "_", ".")
		var body []byte
		var err error

		body, isByteMessage := command.payload.([]byte)
		if !isByteMessage {
			body, err = json.Marshal(command.payload)
			if err != nil {
				eventDispatcher.logger.Error().Msg("Failed to convert payload to JSON" + ": " + err.Error())
				//TODO: Can we log this message
			}
		}

		if err == nil {
			err = eventDispatcher.channel.Publish(
				eventDispatcher.exchangeName,
				routingKey,
				false,
				false,
				amqp.Publishing{
					ContentType: "application/json",
					Body:        body,
					Headers:     map[string]interface{}{"X-Authorization": command.token, "X-Correlation-ID": command.corelationID},
				})

			if err != nil {
				if retryCount < 3 {
					eventDispatcher.logger.Warn().Msg("Publish to queue failed. Trying again ... Error: " + err.Error())

					go func(command *queueCommand, retryCount int) {
						time.Sleep(time.Second)
						eventDispatcher.retryChannel <- &retryCommand{retryCount: retryCount, command: command}
					}(command, retryCount+1)
				} else {
					eventDispatcher.logger.Error().Msg("Failed to publish to an Exchange" + ": " + err.Error())
					//TODO: Can we log this message
				}
			} else {
				eventDispatcher.logger.Trace().Msg("Sent message to queue")
			}
		}
	}
}

func (eventDispatcher *RabbitMQEventDispatcher) rabbitConnector() {
	var rabbitErr *amqp.Error

	for {
		rabbitErr = <-eventDispatcher.connectionCloseChannel
		if rabbitErr != nil {
			eventDispatcher.connectionMutex.Lock()

			connectionString, isTLS := getQueueConnectionString()
			connection, channel := connectToRabbitMQ(eventDispatcher.logger, connectionString, isTLS, eventDispatcher.exchangeName)

			eventDispatcher.connection = connection
			eventDispatcher.channel = channel
			eventDispatcher.connectionCloseChannel = make(chan *amqp.Error)

			eventDispatcher.connection.NotifyClose(eventDispatcher.connectionCloseChannel)

			eventDispatcher.connectionMutex.Unlock()
		}
	}
}

func connectToRabbitMQ(logger *zerolog.Logger, connectionString string, isTLS bool, exchangeName string) (*amqp.Connection, *amqp.Channel) {
	logger.Debug().Msg("Connecting to queue " + connectionString)
	for {

		conn, err := dialAMQP(connectionString, isTLS, logger)
		logger.Info().Msg(fmt.Sprintf("Connection String and TLS valus is %v     %v", connectionString, isTLS))

		if err == nil {
			logger.Info().Msg("RabittMQ connected")

			ch, err := conn.Channel()
			if err != nil {
				logger.Warn().Msg("Failed to open a Channel" + ": " + err.Error())
			} else {
				err = ch.ExchangeDeclare(exchangeName, "topic", true, false, false, false, nil)
				if err != nil {
					logger.Warn().Msg("Failed to declare an exchange" + ": " + err.Error())
				} else {
					return conn, ch
				}
			}
			conn.Close()
		}

		logger.Warn().Msgf("Cannot connect to RabbitMQ, Error [%v]. Trying again...", err)
		time.Sleep(5 * time.Second)
	}
}

func dialAMQP(connectionString string, isTLS bool, logger *zerolog.Logger) (*amqp.Connection, error) {
	var cfg *tls.Config = nil
	if isTLS {
		caCert, _ := os.LookupEnv("ISLA_QUEUE_RMQ_CA_CERT")
		clientCert, _ := os.LookupEnv("ISLA_QUEUE_RMQ_CERT")
		clientCertKey, _ := os.LookupEnv("ISLA_QUEUE_RMQ_CERT_KEY")

		if strings.TrimSpace(caCert) == "" || strings.TrimSpace(clientCert) == "" || strings.TrimSpace(clientCertKey) == "" {
			return nil, fmt.Errorf("One or more client certificates not found")
		}

		cfg = &tls.Config{}
		cfg.RootCAs = x509.NewCertPool()
		cfg.RootCAs.AppendCertsFromPEM([]byte(caCert))

		cert, err := tls.X509KeyPair([]byte(clientCert), []byte(clientCertKey))
		if err != nil {
			return nil, err
		}
		cfg.Certificates = append(cfg.Certificates, cert)
	}

	return amqp.DialTLS(connectionString, cfg)
}

//TODO: read from config
func getQueueConnectionString() (string, bool) {
	var queueHost, queuePort, queueUser, queuePassword string
	isTLS := false
	queueProtocol := "amqp"
	queueHost, ok := os.LookupEnv("ISLA_QUEUE_HOST")
	if !ok {
		queueHost = "localhost"
	}
	queuePassword, ok = os.LookupEnv("ISLA_QUEUE_PWD")
	if !ok {
		queuePassword = "guest"
	}
	queueUser, ok = os.LookupEnv("ISLA_QUEUE_USER")
	if !ok {
		queueUser = "guest"
	}
	queuePort, ok = os.LookupEnv("ISLA_QUEUE_PORT")
	if !ok {
		queuePort = "5672"
	}
	tls, ok := os.LookupEnv("ISLA_QUEUE_TLS_ENABLED")
	if ok {
		isTLS, _ = strconv.ParseBool(tls)
		if isTLS {
			queueProtocol = "amqps"
		}
	}

	return fmt.Sprintf("%v://%v:%v@%v:%v/", queueProtocol, queueUser, queuePassword, queueHost, queuePort), isTLS
}
