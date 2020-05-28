package monitor

import (
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
func NewEventMonitorForQueue(logger *zerolog.Logger, queueName string, eventsToMonitor []string, eventSignal chan *EventInfo) (EventMonitor, error) {
	ctxLogger := logger.With().Str("module", "RabbitMQEventMonitor").Logger()
	monitor := &rabbitMQEventMonitor{logger: &ctxLogger, queueName: queueName, eventSignal: eventSignal}

	err := monitor.initialize(eventsToMonitor)
	if err != nil {
		return nil, err
	}

	return monitor, nil
}
