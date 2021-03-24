package event

import (
	"fmt"

	"github.com/islax/microapp/config"
	"github.com/islax/microapp/event/constants"
	"github.com/rs/zerolog"
)

// Dispatcher interface must be implemented by Queue
type Dispatcher interface {
	DispatchEvent(token string, corelationID string, topic string, payload interface{})
}

func NewEventDispatcher(appConfig *config.Config, logger *zerolog.Logger) (Dispatcher, error) {
	switch appConfig.GetString("MESSAGE_BROKER") {
	case constants.QUEUE_RABBITMQ:
		return NewRabbitMQEventDispatcher(logger)
	case constants.QUEUE_ACTIVEMQ:
		return NewActiveMQEventDispatcher(logger)
	default:
		return nil, fmt.Errorf("Invalid MESSAGE_BROKER value %s. Possible values are %s, %s", appConfig.GetString("MESSAGE_BROKER"), constants.QUEUE_RABBITMQ, constants.QUEUE_ACTIVEMQ)
	}
}
