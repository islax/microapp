package event

import (
	"encoding/json"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-stomp/stomp/v3"
	"github.com/rs/zerolog"
)

// ActiveMQEventDispatcher is an event dispatcher that sends event to the RabbitMQ Exchange
type ActiveMQEventDispatcher struct {
	logger          *zerolog.Logger
	sendChannel     chan *queueCommand
	retryChannel    chan *retryCommand
	connection      *stomp.Conn
	connectionMutex sync.Mutex
}

// NewActiveMQEventDispatcher create and returns a new ActiveMQEventDispatcher
func NewActiveMQEventDispatcher(logger *zerolog.Logger) (*ActiveMQEventDispatcher, error) {
	sendChannel := make(chan *queueCommand, 200)
	retryChannel := make(chan *retryCommand, 200)

	stompConn, err := stomp.Dial("tcp", getQueueHostPort())
	if err != nil {
		return nil, err
	}

	ctxLogger := logger.With().Str("module", "ActiveMQEventDispatcher").Logger()

	dispatcher := &ActiveMQEventDispatcher{
		logger:       &ctxLogger,
		connection:   stompConn,
		sendChannel:  sendChannel,
		retryChannel: retryChannel,
	}

	go dispatcher.start()

	return dispatcher, nil
}

func (eventDispatcher *ActiveMQEventDispatcher) start() {

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

			err = eventDispatcher.connection.Send(routingKey, "application/json", body, stomp.SendOpt.Header("X-Authorization", command.token), stomp.SendOpt.Header("X-Correlation-ID", command.correlationID))

			if err != nil {
				if retryCount < 3 {
					eventDispatcher.logger.Warn().Msg("Publish to queue failed. Trying again ... Error: " + err.Error())

					go func(command *queueCommand, retryCount int) {
						time.Sleep(time.Second)
						eventDispatcher.retryChannel <- &retryCommand{retryCount: retryCount, command: command}
					}(command, retryCount+1)
				} else {
					eventDispatcher.logger.Error().Msg("Failed to publish to an Exchange" + ": " + err.Error())
				}
			} else {
				eventDispatcher.logger.Trace().Msg("Sent message to queue")
			}
		}
	}
}

// DispatchEvent dispatches events to the message queue
func (eventDispatcher *ActiveMQEventDispatcher) DispatchEvent(token string, correlationID string, topic string, payload interface{}) {
	eventDispatcher.sendChannel <- &queueCommand{token: token, topic: topic, payload: payload}
}

func getQueueHostPort() string {
	var queueHost string
	queueHost = os.Getenv("ISLA_QUEUE_HOST")
	if len(queueHost) == 0 {
		queueHost = "localhost"
	}

	var queuePort string
	queuePort = os.Getenv("ISLA_QUEUE_PORT")
	if len(queueHost) == 0 {
		queuePort = "61616"
	}

	return net.JoinHostPort(queueHost, queuePort)
}
