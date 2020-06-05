package event

import (
	"bytes"
	"encoding/json"
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
	var evt map[string]interface{}
	d := json.NewDecoder(bytes.NewReader(p))
	d.UseNumber()
	err = d.Decode(&evt)

	if err == nil {
		writer.eventDispatcher.DispatchEvent("", "", "app_log", evt)
		return len(p), nil
	}

	return 0, err
}
