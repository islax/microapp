package error

//NOTE: Please maintain in ascending order
const (
	// ErrorCodeAPICallFailure error code for API call failure
	ErrorCodeAPICallFailure = "Key_APICallFailure"
	// ErrorCodeCryptoFailure error code for encrypt / decrypt / hashing failure
	ErrorCodeCryptoFailure = "Key_CryptoFailure"
	// ErrorCodeDatabaseFailure error code for database falure
	ErrorCodeDatabaseFailure = "Key_DBQueryFailure"
	// ErrorCodeDuplicateValue error code for duplicate value
	ErrorCodeDuplicateValue = "Key_AlreadyExists"
	// ErrorCodeEmptyRequestBody error code for empty request body
	ErrorCodeEmptyRequestBody = "Key_EmptyRequestBody"
	// ErrorCodeHTTPCreateRequestFailure error code for http request creation failure
	ErrorCodeHTTPCreateRequestFailure = "Key_HTTPCreateRequestFailure"
	// ErrorCodeInvalidFormData error code for form parsing error
	ErrorCodeInvalidFormData = "Key_InvalidFormData"
	// ErrorCodeInternalError error code for internal error
	ErrorCodeInternalError = "Key_InternalError"
	// ErrorCodeInvalidFields error code for invalid fields
	ErrorCodeInvalidFields = "Key_InvalidFields"
	// ErrorCodeInvalidJSON error code for invalid JSON
	ErrorCodeInvalidJSON = "Key_InvalidJSON"
	// ErrorCodeInvalidPublicKey error code for invalid public cert
	ErrorCodeInvalidPublicKey = "Key_InvalidPublicKey"
	// ErrorCodeInvalidRequestPayload error code for invalid request payload
	ErrorCodeInvalidRequestPayload = "Key_InvalidRequestPayload"
	// ErrorCodeInvalidValue error code for invalid value
	ErrorCodeInvalidValue = "Key_InvalidValue"
	// ErrorCodeJSONMarshalFailure error code for json marshal failures
	ErrorCodeJSONMarshalFailure = "Key_JSONMarshalFailure"
	// ErrorCodeNotExists error code for not exists
	ErrorCodeNotExists = "Key_NotExists"
	// ErrorCodeReadWriteFailure error code for io error
	ErrorCodeReadWriteFailure = "Key_ReadWriteFailure"
	// ErrorCodeRequired error code for required fields
	ErrorCodeRequired = "Key_Required"
	// ErrorCodeStringExpected error code for string type
	ErrorCodeStringExpected = "Key_StringExpected"
)
