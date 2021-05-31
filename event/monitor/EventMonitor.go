package monitor

import (
	"fmt"

	"github.com/islax/microapp/config"
	"github.com/islax/microapp/event/constants"
	"github.com/rs/zerolog"
)

// EventMonitor represents an interface used to monitor Events
type EventMonitor interface {
	initialize(eventsToMonitor []string) error
	Start()
	Stop()
}

// NewEventMonitorForQueue creates a new eventMonitor that publishes received events from a named queue to the specified channel
func NewEventMonitorForQueue(logger *zerolog.Logger, queueName string, eventsToMonitor []string, eventSignal chan *EventInfo, appConfig *config.Config) (EventMonitor, error) {
	var monitor EventMonitor
	switch appConfig.GetString(config.EvMessageBroker) {
	case constants.QUEUE_RABBITMQ:
		ctxLogger := logger.With().Str("module", "RabbitMQEventMonitor").Logger()
		monitor = &rabbitMQEventMonitor{logger: &ctxLogger, queueName: queueName, eventSignal: eventSignal}
	case constants.QUEUE_ACTIVEMQ:
		ctxLogger := logger.With().Str("module", "ActiveMQEventMonitor").Logger()
		monitor = &activeMQEventMonitor{logger: &ctxLogger, queueName: queueName, eventSignal: eventSignal}
	case constants.QUEUE_AWSSQS:
		ctxLogger := logger.With().Str("module", "SQSEventMonitor").Logger()
		monitor = &sqsEventMonitor{logger: &ctxLogger, queueName: queueName, eventSignal: eventSignal}
	default:
		return nil, fmt.Errorf("invalid message broker value %s. possible values are %s, %s", appConfig.GetString("MESSAGE_BROKER"), constants.QUEUE_RABBITMQ, constants.QUEUE_ACTIVEMQ)
	}

	err := monitor.initialize(eventsToMonitor)
	if err != nil {
		return nil, err
	}

	return monitor, nil
}
