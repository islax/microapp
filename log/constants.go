package log

const (
	// EventTypeAuthenticationErr log event type for validation error
	EventTypeAuthenticationErr = "Key_AuthenticationError"
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

const (
	// MessageInvalidInputData error message for - invalid entity data
	MessageInvalidInputData = "Invalid input data."
	// MessageUnableToFindURLResource error message for - unable to find URL resource
	MessageUnableToFindURLResource = "Unable to find record."
	// MessageUnexpectedErrAddingNewEntityToDB error message for - unexpected error occured while adding new entity to database
	MessageUnexpectedErrAddingNewEntityToDB = "Unexpected error occured while adding to database."
	// MessageUnexpectedErrCreatingNewEntity error message for - unexpected error occured while creating new entity
	MessageUnexpectedErrCreatingNewEntity = "Unexpected error occured while creating / validating new entity."
	// MessageUnexpectedErrDeletingEntityFromDB error message for - unexpected error occured while deleting entity from database
	MessageUnexpectedErrDeletingEntityFromDB = "Unexpected error occured while deleting from database."
	// MessageUnexpectedErrGetEntitiesFromDB error message for - unexpected error occured while getting entities from database
	MessageUnexpectedErrGetEntitiesFromDB = "Unexpected error occured while getting data from database."
	// MessageUnexpectedErrGenericTemplate error message for - unexpected error occured while generic template
	MessageUnexpectedErrGenericTemplate = "Unexpected error occured while %v."
	// MessageUnexpectedErrRequetPayloadParsing error message for - unexpected error while request JSON parsing
	MessageUnexpectedErrRequetPayloadParsing = "Unexpected error occured while parsing JSON payload."
	// MessageUnexpectedErrUpdatingEntityToDB error message for - unexpected error occured while adding new entity to database
	MessageUnexpectedErrUpdatingEntityToDB = "Unexpected error occured while updating to database."
)
