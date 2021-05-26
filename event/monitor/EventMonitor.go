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

// NewEventMonitor creates a new eventMonitor that publishes received events to the specified channel
func NewEventMonitor(logger *zerolog.Logger, eventsToMonitor []string, eventSignal chan *EventInfo) (EventMonitor, error) {
	ctxLogger := logger.With().Str("module", "RabbitMQEventMonitor").Logger()
	monitor := &rabbitMQEventMonitor{logger: &ctxLogger, eventSignal: eventSignal}

	err := monitor.initialize(eventsToMonitor)
	if err != nil {
		return nil, err
	}

	return monitor, nil
}

// NewEventMonitorForQueue creates a new eventMonitor that publishes received events from a named queue to the specified channel
func NewEventMonitorForQueue(logger *zerolog.Logger, queueName string, eventsToMonitor []string, eventSignal chan *EventInfo, appConfig *config.Config) (EventMonitor, error) {
	ctxLogger := logger.With().Str("module", "ActiveMQEventMonitor").Logger()
	var monitor EventMonitor
	switch appConfig.GetString("MESSAGE_BROKER") {
	case constants.QUEUE_RABBITMQ:
		monitor = &rabbitMQEventMonitor{logger: &ctxLogger, queueName: queueName, eventSignal: eventSignal}
	case constants.QUEUE_ACTIVEMQ:
		monitor = &activeMQEventMonitor{logger: &ctxLogger, queueName: queueName, eventSignal: eventSignal}
	default:
		return nil, fmt.Errorf("Invalid MESSAGE_BROKER value %s. Possible values are %s, %s", appConfig.GetString("MESSAGE_BROKER"), constants.QUEUE_RABBITMQ, constants.QUEUE_ACTIVEMQ)
	}
	err := monitor.initialize(eventsToMonitor)
	if err != nil {
		return nil, err
	}

	return monitor, nil
}
