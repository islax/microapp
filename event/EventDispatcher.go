package event

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/islax/microapp/config"
	"github.com/islax/microapp/event/constants"
	"github.com/rs/zerolog"
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
type EventDispatcher struct {
	logger                 *zerolog.Logger
	sendChannel            chan *queueCommand
	retryChannel           chan *retryCommand
	connectionCloseChannel chan error
	connectionMutex        sync.Mutex
	Queue
}

type Queue interface {
	connect()
	publish(topic string, headers map[string]interface{}, body []byte) error
}

// NewRabbitMQEventDispatcher create and returns a new RabbitMQEventDispatcher
func NewEventDispatcher(logger *zerolog.Logger, queue Queue) (*EventDispatcher, error) {
	sendChannel := make(chan *queueCommand, 200)
	retryChannel := make(chan *retryCommand, 200)

	ctxLogger := logger.With().Str("module", "EventDispatcher").Logger()

	dispatcher := &EventDispatcher{
		logger:       &ctxLogger,
		sendChannel:  sendChannel,
		retryChannel: retryChannel,
		Queue:        queue,
	}

	go dispatcher.connector()
	go dispatcher.start()
	dispatcher.connectionCloseChannel <- errors.New("trigger conn")

	return dispatcher, nil
}

// DispatchEvent dispatches events to the message queue
func (eventDispatcher *EventDispatcher) DispatchEvent(token string, corelationID string, topic string, payload interface{}) {
	eventDispatcher.sendChannel <- &queueCommand{token: token, topic: topic, payload: payload}
}

func (eventDispatcher *EventDispatcher) start() {

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
			//eventDispatcher.Queue.Publish
			err = eventDispatcher.Queue.publish(command.topic, map[string]interface{}{"X-Authorization": command.token, "X-Correlation-ID": command.corelationID}, body)

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

func (eventDispatcher *EventDispatcher) connector() {
	for {
		err := <-eventDispatcher.connectionCloseChannel
		if err != nil {
			eventDispatcher.connectionMutex.Lock()
			eventDispatcher.Queue.connect()
			eventDispatcher.connectionMutex.Unlock()
		}
	}
}

func NewEventPublisher(appConfig *config.Config, logger *zerolog.Logger) (Queue, error) {
	switch appConfig.GetString(config.EvMessageBroker) {
	case constants.QUEUE_RABBITMQ:
		return NewRabbitMQPublisher(logger), nil
	case constants.QUEUE_ACTIVEMQ:
		return NewActiveMQPublisher(logger)
	//case constants.QUEUE_AWSSQS:
	//return NewSQSEventDispatcher(logger)
	default:
		return nil, fmt.Errorf("invalid message broker value %s. possible values are %s, %s, %s", appConfig.GetString("MESSAGE_BROKER"), constants.QUEUE_RABBITMQ, constants.QUEUE_ACTIVEMQ, constants.QUEUE_AWSSQS)
	}
}
