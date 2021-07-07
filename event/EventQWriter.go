package event

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	fmt.Println("Write", "1")
	var evt map[string]interface{}
	d := json.NewDecoder(bytes.NewReader(p))
	d.UseNumber()
	err = d.Decode(&evt)
	fmt.Println("Write", "2")
	if err == nil {
		fmt.Println("Write", "3")
		writer.eventDispatcher.DispatchEvent("", "", "app_log", evt)
		return len(p), nil
	}
	fmt.Println("Write", "4")
	return 0, err
}
