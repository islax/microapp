package event

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/go-stomp/stomp"
	"github.com/rs/zerolog"
)

// ActiveMQEventDispatcher is an event dispatcher that sends event to the RabbitMQ Exchange
type ActiveMQEventDispatcher struct {
	logger     *zerolog.Logger
	connection *stomp.Conn
}

// NewActiveMQEventDispatcher create and returns a new ActiveMQEventDispatcher
func NewActiveMQEventDispatcher(logger *zerolog.Logger) (*ActiveMQEventDispatcher, error) {
	netConn, err := net.DialTimeout("tcp", "172.20.80.1:61613", 10*time.Second)
	if err != nil {
		return nil, err
	}

	stompConn, err := stomp.Connect(netConn, stomp.Options{HeartBeat: "1000,0"})
	if err != nil {
		return nil, err
	}

	ctxLogger := logger.With().Str("module", "ActiveMQEventDispatcher").Logger()

	dispatcher := &ActiveMQEventDispatcher{
		logger:     &ctxLogger,
		connection: stompConn,
	}

	return dispatcher, nil
}

// DispatchEvent dispatches events to the message queue
func (eventDispatcher *ActiveMQEventDispatcher) DispatchEvent(token string, corelationID string, topic string, payload interface{}) {
	body, isByteMessage := payload.([]byte)
	var err error
	if !isByteMessage {
		body, err = json.Marshal(payload)
		if err != nil {
			eventDispatcher.logger.Error().Msg("Failed to convert payload to JSON" + ": " + err.Error())
			//TODO: Can we log this message
		}
	}

	//headers
	h := stomp.NewHeader(
		"X-Authorization", token,
		"X-Correlation-ID", corelationID)
	fmt.Println(fmt.Sprintf("%+v", eventDispatcher.connection))
	if err := eventDispatcher.connection.Send(topic, "application/json", body, h); err != nil {
		fmt.Println(err)
		eventDispatcher.logger.Error().Msg("Failed to publish to queue" + ": " + err.Error())
		return
	}

	eventDispatcher.logger.Trace().Msg("Sent message to queue")
}
