package event

import (
	"net"
	"os"
	"strings"

	"github.com/go-stomp/stomp/v3"
	"github.com/rs/zerolog"
)

// ActiveMQEventDispatcher is an event dispatcher that sends event to the RabbitMQ Exchange
type ActiveMQPublisher struct {
	logger     *zerolog.Logger
	connection *stomp.Conn
}

// NewActiveMQEventDispatcher create and returns a new ActiveMQEventDispatcher
func NewActiveMQPublisher(logger *zerolog.Logger) (*ActiveMQPublisher, error) {
	ctxLogger := logger.With().Str("module", "ActiveMQEventDispatcher").Logger()

	publisher := &ActiveMQPublisher{
		logger: &ctxLogger,
	}

	return publisher, nil
}

func (eventPublisher *ActiveMQPublisher) publish(topic string, headers map[string]interface{}, body []byte) error {
	routingKey := strings.ReplaceAll(topic, "_", ".")
	return eventPublisher.connection.Send(routingKey, "application/json", body, stomp.SendOpt.Header("X-Authorization", headers["X-Authorization"].(string)), stomp.SendOpt.Header("X-Correlation-ID", headers["X-Correlation-ID"].(string)))
}

func (eventPublisher *ActiveMQPublisher) connect() {
	eventPublisher.logger.Debug().Msg("Connecting to queue " + getQueueHostPort())
	stompConn, _ := stomp.Dial("tcp", getQueueHostPort())
	eventPublisher.connection = stompConn
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
