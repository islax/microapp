package event

import (
	"fmt"

	"github.com/islax/microapp/config"
	"github.com/islax/microapp/event/constants"
	"github.com/rs/zerolog"
)

// Dispatcher interface must be implemented by Queue
type Dispatcher interface {
	DispatchEvent(token string, correlationID string, topic string, payload interface{})
}

func NewEventDispatcher(appConfig *config.Config, logger *zerolog.Logger) (Dispatcher, error) {
	switch appConfig.GetString(config.EvMessageBroker) {
	case constants.QUEUE_RABBITMQ:
		return NewRabbitMQEventDispatcher(logger)
	case constants.QUEUE_ACTIVEMQ:
		return NewActiveMQEventDispatcher(logger)
	case constants.QUEUE_AWSSQS:
		return NewSQSEventDispatcher(logger)
	default:
		return nil, fmt.Errorf("invalid message broker value %s. possible values are %s, %s, %s", appConfig.GetString("MESSAGE_BROKER"), constants.QUEUE_RABBITMQ, constants.QUEUE_ACTIVEMQ, constants.QUEUE_AWSSQS)
	}
}
