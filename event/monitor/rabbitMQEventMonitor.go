package monitor

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
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
		queueConnection, err := dialAMQP(connectionString, isTLS, monitor.logger)

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
		corelationID := ""
		authorizationHeader, ok := message.Headers["X-Authorization"]
		if ok {
			token = authorizationHeader.(string)
		}
		corelationIDHeader, ok := message.Headers["X-Correlation-ID"]
		if ok {
			corelationID = corelationIDHeader.(string)
		}

		command := &EventInfo{
			CorelationID: corelationID,
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

func dialAMQP(connectionString string, isTLS bool, logger *zerolog.Logger) (*amqp.Connection, error) {
	var cfg *tls.Config = nil
	if isTLS {
		caCertPath, _ := os.LookupEnv("ISLA_QUEUE_CA_CERT_PATH")
		clientCertPath, _ := os.LookupEnv("ISLA_QUEUE_CLIENT_CERT_PATH")
		clientCertKeyPath, _ := os.LookupEnv("ISLA_QUEUE_CLIENT_CERT_KEY_PATH")

		cfg = &tls.Config{}
		cfg.RootCAs = x509.NewCertPool()

		if ca, err := ioutil.ReadFile(caCertPath); err == nil {
			cfg.RootCAs.AppendCertsFromPEM(ca)
		} else {
			return nil, err
		}

		if cert, err := tls.LoadX509KeyPair(clientCertPath, clientCertKeyPath); err == nil {
			cfg.Certificates = append(cfg.Certificates, cert)
		} else {
			return nil, err
		}
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
