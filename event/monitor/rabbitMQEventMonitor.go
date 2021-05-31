package monitor

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

type rabbitMQEventMonitor struct {
	logger          *zerolog.Logger
	queueName       string
	eventSignal     chan *EventInfo
	eventsToMonitor []string

	messageChanel   <-chan amqp.Delivery
	queueConnection *amqp.Connection
	queueChannel    *amqp.Channel

	connectionCloseChannel chan *amqp.Error
}

func (monitor *rabbitMQEventMonitor) initialize(eventsToMonitor []string) error {
	monitor.connectionCloseChannel = make(chan *amqp.Error)
	monitor.eventsToMonitor = eventsToMonitor
	go monitor.rabbitConnector()

	return nil
}

func (monitor *rabbitMQEventMonitor) rabbitConnector() {
	var rabbitErr *amqp.Error

	for {
		rabbitErr = <-monitor.connectionCloseChannel
		if rabbitErr != nil {
			connectionString, isTLS := getQueueConnectionString()
			connection, queueChannel, messageChanel := monitor.connectToRabbitMQ(monitor.queueName, connectionString, isTLS, monitor.eventsToMonitor)

			monitor.queueConnection = connection
			monitor.queueChannel = queueChannel
			monitor.messageChanel = messageChanel

			monitor.connectionCloseChannel = make(chan *amqp.Error)
			monitor.queueConnection.NotifyClose(monitor.connectionCloseChannel)

			go monitor.monitorQueueAndProcessMessages()
		}
	}
}

func (monitor *rabbitMQEventMonitor) connectToRabbitMQ(queueName string, connectionString string, isTLS bool, eventsToMonitor []string) (*amqp.Connection, *amqp.Channel, <-chan amqp.Delivery) {
	for {
		queueConnection, err := dialAMQP(connectionString, isTLS)

		if err != nil {
			monitor.logger.Error().Err(err).Msg("Unable to connect to rabbitMQ.")
		} else {
			queueChannel, err := queueConnection.Channel()
			if err != nil {
				monitor.logger.Error().Err(err).Msg("Failed to open a channel.")
			} else {
				err = queueChannel.ExchangeDeclare(
					"isla_exchange", // name
					"topic",         // type
					true,            // durable
					false,           // auto-deleted
					false,           // internal
					false,           // no-wait
					nil,             // arguments
				)
				if err != nil {
					monitor.logger.Error().Err(err).Msg("Failed to declare an exchange.")
				} else {
					q, err := queueChannel.QueueDeclare(
						monitor.queueName, // name
						false,             // durable
						false,             // delete when unused
						false,             // exclusive
						false,             // no-wait
						nil,               // arguments
					)
					if err != nil {
						monitor.logger.Error().Err(err).Msg("Failed to declare a queue.")
					} else {
						for _, event := range eventsToMonitor {
							normalizedEvent := strings.ReplaceAll(event, "_", ".")
							err = queueChannel.QueueBind(
								q.Name,          // queue name
								normalizedEvent, // routing key
								"isla_exchange", // exchange
								false,
								nil)
							if err != nil {
								monitor.logger.Error().Err(err).Msgf("Failed to bind a event - %v", event)
							}
						}

						messageChanel, err := queueChannel.Consume(
							q.Name, // queue
							"",     // consumer
							true,   // auto ack
							false,  // exclusive
							false,  // no local
							false,  // no wait
							nil,    // args
						)
						if err != nil {
							monitor.logger.Error().Err(err).Msg("Failed to register a consumer.")
						} else {
							return queueConnection, queueChannel, messageChanel
						}
					}
				}
			}
		}
		monitor.logger.Warn().Msgf("Cannot connect to RabbitMQ. Trying again ... Error %s", err.Error())
		time.Sleep(5 * time.Second)
	}
}

func (monitor *rabbitMQEventMonitor) monitorQueueAndProcessMessages() {
	for message := range monitor.messageChanel {
		payload := string(message.Body)
		token := ""
		correlationID := ""
		authorizationHeader, ok := message.Headers["X-Authorization"]
		if ok {
			token = authorizationHeader.(string)
		}
		correlationIDHeader, ok := message.Headers["X-Correlation-ID"]
		if ok {
			correlationID = correlationIDHeader.(string)
		}

		command := &EventInfo{
			CorelationID: correlationID,
			Payload:      payload,
			RawToken:     token,

			Name: message.RoutingKey,
		}

		monitor.eventSignal <- command
	}
}

func (monitor *rabbitMQEventMonitor) Start() {
	monitor.connectionCloseChannel <- amqp.ErrClosed // Trigger the connection
}

func (monitor *rabbitMQEventMonitor) Stop() {
	monitor.queueChannel.Close()
	monitor.queueConnection.Close()
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
