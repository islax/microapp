package events

// EventDispatcher interface must be implemented by Queue
type EventDispatcher interface {
	DispatchEvent(token string, topic string, payload interface{})
}
