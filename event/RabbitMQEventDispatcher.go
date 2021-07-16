package event

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
)

// RabbitMQPublisher is an event publisher that sends event to the RabbitMQ Exchange
type RabbitMQPublisher struct {
	logger                 *zerolog.Logger
	exchangeName           string
	connection             *amqp.Connection
	channel                *amqp.Channel
	connectionCloseChannel chan *amqp.Error
}

// NewRabbitMQEventDispatcher create and returns a new RabbitMQEventDispatcher
func NewRabbitMQPublisher(logger *zerolog.Logger) *RabbitMQPublisher {
	connectionCloseChannel := make(chan *amqp.Error)

	ctxLogger := logger.With().Str("module", "RabbitMQEventDispatcher").Logger()

	publisher := &RabbitMQPublisher{
		logger:                 &ctxLogger,
		exchangeName:           "isla_exchange",
		connectionCloseChannel: connectionCloseChannel,
	}

	return publisher
}

func (eventPublisher *RabbitMQPublisher) publish(topic string, headers map[string]interface{}, body []byte) error {
	routingKey := strings.ReplaceAll(topic, "_", ".")
	return eventPublisher.channel.Publish(
		eventPublisher.exchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Headers:     headers,
		})
}

func (eventPublisher *RabbitMQPublisher) connect() {
	connectionString, isTLS, connectionStringForLog := eventPublisher.getQueueConnectionString()
	connection, channel := connectToRabbitMQ(eventPublisher.logger, connectionString, isTLS, eventPublisher.exchangeName, connectionStringForLog)
	eventPublisher.connection = connection
	eventPublisher.channel = channel
	eventPublisher.connectionCloseChannel = make(chan *amqp.Error)
	eventPublisher.connection.NotifyClose(eventPublisher.connectionCloseChannel)
}

func connectToRabbitMQ(logger *zerolog.Logger, connectionString string, isTLS bool, exchangeName string, connectionStringForLog string) (*amqp.Connection, *amqp.Channel) {
	logger.Debug().Msg("Connecting to queue " + connectionStringForLog)
	for {

		conn, err := dialAMQP(connectionString, isTLS)
		logger.Info().Msg(fmt.Sprintf("Connection String and TLS valus is %v     %v", connectionStringForLog, isTLS))

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

func dialAMQP(connectionString string, isTLS bool) (*amqp.Connection, error) {
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
func (eventPublisher *RabbitMQPublisher) getQueueConnectionString() (string, bool, string) {
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

	return fmt.Sprintf("%v://%v:%v@%v:%v/", queueProtocol, queueUser, queuePassword, queueHost, queuePort), isTLS, fmt.Sprintf("%v://%v:%v@%v:%v/", queueProtocol, "######", "######", queueHost, queuePort)
}
