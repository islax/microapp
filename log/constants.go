package log

const (
	// EventTypeAuthenticationErr log event type for validation error
	EventTypeAuthenticationErr = "Key_AuthenticationError"
	// EventTypeServiceDataReplication log event type for
	EventTypeServiceDataReplication = "Key_ServiceDataReplication"
	// EventTypeSuccess log event type key success
	EventTypeSuccess = "Key_Success"
	// EventTypeUnexpectedErr log event type for unexpected error
	EventTypeUnexpectedErr = "Key_UnexpectedError"
	// EventTypeValidationErr log event type for validation error
	EventTypeValidationErr = "Key_ValidationError"
)

const (
	// EventCodeActionComplete log event code for completion of action
	EventCodeActionComplete = "Key_ActionComplete"
	// EventCodeCryptoFaliure event code for crypto failure
	EventCodeCryptoFaliure = "Key_CryptoFailure"
	// EventCodeInvalidData log event code for invalid data
	EventCodeInvalidData = "Key_InvalidPayload"
	// EventCodeReadWriteFailure event code for read/write errors
	EventCodeReadWriteFailure = "Key_ReadWriteFailure"
	// EventCodeUnknown log event code for unknown errors
	EventCodeUnknown = "Key_Unknown"
)

const (
	// MessageInvalidInputData error message for - invalid entity data
	MessageInvalidInputData = "Invalid input data."
	// MessageUnableToFindURLResource error message for - unable to find URL resource
	MessageUnableToFindURLResource = "Unable to find record."
	// MessageAddEntityError error message for - unexpected error occured while adding new entity to database
	MessageAddEntityError = "Unexpected error occured while adding to database."
	// MessageNewEntityError error message for - unexpected error occured while creating new entity
	MessageNewEntityError = "Unexpected error occured while creating / validating new entity."
	// MessageDeleteEntityError error message for - unexpected error occured while deleting entity from database
	MessageDeleteEntityError = "Unexpected error occured while deleting from database."
	// MessageGetEntityError error message for - unexpected error occured while getting entities from database
	MessageGetEntityError = "Unexpected error occured while getting data from database."
	// MessageGenericErrorTemplate error message for - unexpected error occured while generic template
	MessageGenericErrorTemplate = "Unexpected error occured while %v."
	// MessageParseError error message for - unexpected error while parsing payload
	MessageParseError = "Unexpected error occured while parsing payload."
	// MessageUpdateEntityError error message for - unexpected error occured while adding new entity to database
	MessageUpdateEntityError = "Unexpected error occured while updating to database."
)
