package error

// NewAPICallError creates a new API call error
func NewAPICallError(apiURL string, httpStatusCode *int, httpResponseBody *string, err error) APICallError {
	return &apiCallErrorImpl{apiURL, httpStatusCode, httpResponseBody, createUnexpectedErrorImpl(ErrorCodeAPICallFailure, err)}
}

// APICallError represents an database query failure error interface
type APICallError interface {
	UnexpectedError
	GetAPIURL() string
	GetHTTPStatusCode() *int
	GetHTTPResponseBody() *string
}

type apiCallErrorImpl struct {
	apiURL           string
	httpStatusCode   *int
	httpResponseBody *string
	unexpectedErrorImpl
}

// GetAPIURL gets API URL
func (e *apiCallErrorImpl) GetAPIURL() string {
	return e.apiURL
}

// GetHttpStatusCode gets http status code
func (e *apiCallErrorImpl) GetHTTPStatusCode() *int {
	return e.httpStatusCode
}

// GetHTTPResponseBody gets http status code
func (e *apiCallErrorImpl) GetHTTPResponseBody() *string {
	return e.httpResponseBody
}
