package error

// NewAPICallError creates a new API call error
func NewAPICallError(httpStatusCode int, httpResponseBody string, err error) APICallError {
	return &apiCallErrorImpl{httpStatusCode, httpResponseBody, createUnexpectedErrorImpl(ErrorCodeAPICallFailure, err)}
}

// APICallError represents an database query failure error interface
type APICallError interface {
	UnexpectedError
	GetHTTPStatusCode() int
	GetHTTPResponseBody() string
}

type apiCallErrorImpl struct {
	httpStatusCode   int
	httpResponseBody string
	unexpectedErrorImpl
}

// GetHttpStatusCode gets http status code
func (e *apiCallErrorImpl) GetHTTPStatusCode() int {
	return e.httpStatusCode
}

// GetHTTPResponseBody gets http status code
func (e *apiCallErrorImpl) GetHTTPResponseBody() string {
	return e.httpResponseBody
}
