package event

import (
	"io"
)

type eventQWriter struct {
	eventDispatcher Dispatcher
}

// NewEventQWriter creates new event queue writer
func NewEventQWriter(eventDispatcher Dispatcher) io.Writer {
	return &eventQWriter{eventDispatcher}
}

func (writer *eventQWriter) Write(p []byte) (n int, err error) {
	writer.eventDispatcher.DispatchEvent("", "", "app_log", p)
	return len(p), nil
}
