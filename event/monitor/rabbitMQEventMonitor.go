package monitor

import (
	"fmt"
	"os"
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
			connection, queueChannel, messageChanel := monitor.connectToRabbitMQ(monitor.queueName, monitor.eventsToMonitor)

			monitor.queueConnection = connection
			monitor.queueChannel = queueChannel
			monitor.messageChanel = messageChanel

			monitor.connectionCloseChannel = make(chan *amqp.Error)
			monitor.queueConnection.NotifyClose(monitor.connectionCloseChannel)

			go monitor.monitorQueueAndProcessMessages()
		}
	}
}

func (monitor *rabbitMQEventMonitor) connectToRabbitMQ(queueName string, eventsToMonitor []string) (*amqp.Connection, *amqp.Channel, <-chan amqp.Delivery) {
	for {
		queueConnection, err := amqp.Dial(getQueueConnectionString())
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

func getQueueConnectionString() string {
	var queueHost, queuePort, queueUser, queuePassword string
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

	return fmt.Sprintf("amqp://%v:%v@%v:%v/", queueUser, queuePassword, queueHost, queuePort)
}
