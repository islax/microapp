package event

// Dispatcher interface must be implemented by Queue
type Dispatcher interface {
	DispatchEvent(token string, topic string, payload interface{})
}
