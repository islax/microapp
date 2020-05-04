package event

// Dispatcher interface must be implemented by Queue
type Dispatcher interface {
	DispatchEvent(token string, corelationID string, topic string, payload interface{})
}
