package log

const (
	// EventTypeValidationErr log event type for validation error
	EventTypeValidationErr = "Key_ValidationError"
	// EventTypeUnexpectedErr log event type for unexpected error
	EventTypeUnexpectedErr = "Key_UnexpectedError"
	// EventTypeSuccess log event type key success
	EventTypeSuccess = "Key_Success"
)

const (
	// EventCodeInvalidData log event code for invalid data
	EventCodeInvalidData = "Key_InvalidPayload"
	// EventCodeUnknown log event code for unknown errors
	EventCodeUnknown = "Key_Unknown"
	// EventCodeReadWriteFailure event code for read/write errors
	EventCodeReadWriteFailure = "Key_ReadWriteFailure"
	// EventCodeCryptoFaliure event code for crypto failure
	EventCodeCryptoFaliure = "Key_CryptoFailure"
	// EventCodeActionComplete log event code for completion of action
	EventCodeActionComplete = "Key_ActionComplete"
)
