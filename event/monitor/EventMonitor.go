package monitor

import (
	"github.com/sirupsen/logrus"
)

// EventMonitor represents an interface used to monitor Events
type EventMonitor interface {
	initialize(eventsToMonitor []string) error
	Start()
	Stop()
}

// NewEventMonitor creates a new eventMonitor that publishes received events to the specified channel
func NewEventMonitor(logger *logrus.Logger, eventsToMonitor []string, eventSignal chan *EventInfo) (EventMonitor, error) {
	monitor := &rabbitMQEventMonitor{logger: logger, eventSignal: eventSignal}

	err := monitor.initialize(eventsToMonitor)
	if err != nil {
		return nil, err
	}

	return monitor, nil
}
