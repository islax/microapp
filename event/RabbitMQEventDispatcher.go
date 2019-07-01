package event

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type queueCommand struct {
	token   string
	topic   string
	payload interface{}
}

type retryCommand struct {
	retryCount int
	command    *queueCommand
}

// RabbitMQEventDispatcher is an event dispatcher that sends event to the RabbitMQ Exchange
type RabbitMQEventDispatcher struct {
	logger                 *log.Logger
	exchangeName           string
	connection             *amqp.Connection
	channel                *amqp.Channel
	sendChannel            chan *queueCommand
	retryChannel           chan *retryCommand
	connectionCloseChannel chan *amqp.Error
	connectionMutex        sync.Mutex
}

// NewRabbitMQEventDispatcher create and returns a new RabbitMQEventDispatcher
func NewRabbitMQEventDispatcher(logger *log.Logger) (*RabbitMQEventDispatcher, error) {
	sendChannel := make(chan *queueCommand, 200)
	retryChannel := make(chan *retryCommand, 200)
	connectionCloseChannel := make(chan *amqp.Error)

	dispatcher := &RabbitMQEventDispatcher{
		logger:                 logger,
		exchangeName:           "isla_Exchange",
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
func (eventDispatcher *RabbitMQEventDispatcher) DispatchEvent(token string, topic string, payload interface{}) {
	eventDispatcher.sendChannel <- &queueCommand{token: token, topic: topic, payload: payload}
}

func (eventDispatcher *RabbitMQEventDispatcher) start() {
	contextLogger := eventDispatcher.logger.WithFields(log.Fields{
		"module": "RabbitMQEventDispatcher",
	})

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
		body, err := json.Marshal(command.payload)
		if err != nil {
			contextLogger.Error("Failed to convert payload to JSON" + ": " + err.Error())
			//TODO: Can we log this message
		}

		if err == nil {
			err = eventDispatcher.channel.Publish(
				eventDispatcher.exchangeName,
				routingKey,
				false,
				false,
				amqp.Publishing{
					ContentType: "application/json",
					Body:        []byte(body),
				})

			if err != nil {
				if retryCount < 3 {
					contextLogger.Warn("Publish to queue failed. Trying again ... Error: " + err.Error())

					go func(command *queueCommand, retryCount int) {
						time.Sleep(time.Second)
						eventDispatcher.retryChannel <- &retryCommand{retryCount: retryCount, command: command}
					}(command, retryCount+1)
				} else {
					contextLogger.Error("Failed to publish to an Exchange" + ": " + err.Error())
					//TODO: Can we log this message
				}
			} else {
				contextLogger.Info("Sent message to queue")
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

			connectionString := getQueueConnectionString()
			connection, channel := connectToRabbitMQ(eventDispatcher.logger, connectionString, eventDispatcher.exchangeName)

			eventDispatcher.connection = connection
			eventDispatcher.channel = channel
			eventDispatcher.connectionCloseChannel = make(chan *amqp.Error)

			eventDispatcher.connection.NotifyClose(eventDispatcher.connectionCloseChannel)

			eventDispatcher.connectionMutex.Unlock()
		}
	}
}

func connectToRabbitMQ(logger *log.Logger, connectionString string, exchangeName string) (*amqp.Connection, *amqp.Channel) {
	logger.Infof("Connecting to queue %s\n", connectionString)
	for {
		conn, err := amqp.Dial(connectionString)

		if err == nil {
			logger.Info("RabittMQ connected")

			ch, err := conn.Channel()
			if err != nil {
				logger.Warn("Failed to open a Channel" + ": " + err.Error())
			} else {
				err = ch.ExchangeDeclare(exchangeName, "topic", true, false, false, false, nil)
				if err != nil {
					logger.Warn("Failed to declare an exchange" + ": " + err.Error())
				} else {
					return conn, ch
				}
			}
			conn.Close()
		}

		logger.Warnf("Cannot connect to RabbitMQ. Trying again ... Error %s", err.Error())
		time.Sleep(5 * time.Second)
	}
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
