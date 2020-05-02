package error

//NOTE: Please maintain in ascending order
const (
	// ErrorKeyInvalidFields error key for invalid fields
	ErrorKeyInvalidFields = "Key_InvalidFields"
	// ErrorKeyInvalidRequestPayload error key for invalid request payload
	ErrorKeyInvalidRequestPayload = "Key_InvalidRequestPayload"
)

//NOTE: Please maintain in ascending order
const (
	// ErrorCodeCryptoFailure error code for encrypt / decrypt / hashing failure
	ErrorCodeCryptoFailure = "Key_CryptoFailure"
	// ErrorCodeDatabaseFailure error code for database falure
	ErrorCodeDatabaseFailure = "Key_DBQueryFailure"
	// ErrorCodeDuplicateValue error code for duplicate value
	ErrorCodeDuplicateValue = "Key_AlreadyExists"
	// ErrorCodeEmptyRequestBody error code for empty request body
	ErrorCodeEmptyRequestBody = "Key_EmptyRequestBody"
	// ErrorCodeInternalError error code for internal error
	ErrorCodeInternalError = "Key_InternalError"
	// ErrorCodeInvalidJSON error code for invalid JSON
	ErrorCodeInvalidJSON = "Key_InvalidJSON"
	// ErrorCodeInvalidValue error code for invalid value
	ErrorCodeInvalidValue = "Key_InvalidValue"
	// ErrorCodeNotExists error code for not exists
	ErrorCodeNotExists = "Key_NotExists"
	// ErrorCodeReadWriteFailure error code for io error
	ErrorCodeReadWriteFailure = "Key_ReadWriteFailure"
	// ErrorCodeRequired error code for required fields
	ErrorCodeRequired = "Key_Required"
	// ErrorCodeStringExpected error code for string type
	ErrorCodeStringExpected = "Key_StringExpected"
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
	// MessageUnexpectedErrWhileValidatingXOperationData error message for - unexpected error occured while validation x operation data
	MessageUnexpectedErrWhileValidatingXOperationData = "Unexpected error occured while validating %v data."
)
