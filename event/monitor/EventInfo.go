package monitor

// EventInfo represents the message received from queu
type EventInfo struct {
	RawToken     string
	CorelationID string
	Name         string
	Payload      string
}
