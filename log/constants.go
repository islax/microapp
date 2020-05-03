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
	// MessageInvalidData error message for - invalid entity data
	MessageInvalidData = "Invalid %v data."
	// MessageInvalidPayload error message for - invalid payload
	MessageInvalidPayload = "Invalid payload."
	// MessageUnableToFindURLResource error message for - unable to find URL resource
	MessageUnableToFindURLResource = "Unable to find %v."
	// MessageUnexpectedErrWhileAddingNewEntityToDB error message for - unexpected error occured while adding new entity to database
	MessageUnexpectedErrWhileAddingNewEntityToDB = "Unexpected error occured while adding %v to database."
	// MessageUnexpectedErrWhileCreatingNewEntity error message for - unexpected error occured while creating new entity
	MessageUnexpectedErrWhileCreatingNewEntity = "Unexpected error occured while creating new %v."
	// MessageUnexpectedErrWhileDeletingEntityFromDB error message for - unexpected error occured while deleting entity from database
	MessageUnexpectedErrWhileDeletingEntityFromDB = "Unexpected error occured while deleting %v from database."
	// MessageUnexpectedErrWhileGetEntitiesFromDB error message for - unexpected error occured while getting entities from database
	MessageUnexpectedErrWhileGetEntitiesFromDB = "Unexpected error occured while getting %v from database."
	// MessageUnexpectedErrWhileGenericTemplate error message for - unexpected error occured while generic template
	MessageUnexpectedErrWhileGenericTemplate = "Unexpected error occured while %v."
	// MessageUnexpectedErrWhileRequetPayloadParsing error message for - unexpected error while request JSON parsing
	MessageUnexpectedErrWhileRequetPayloadParsing = "Unexpected error occured while parsing JSON paload."
	// MessageUnexpectedErrWhileUpdatingEntityInDB error message for - unexpected error occured while adding new entity to database
	MessageUnexpectedErrWhileUpdatingEntityInDB = "Unexpected error occured while updating %v in database."
)
