package error

// NewAPIClientError creates a new API call error
func NewAPIClientError(apiURL string, httpStatusCode *int, httpResponseBody *string, err error) APIClientError {
	return &APIClientErrorImpl{apiURL, httpStatusCode, httpResponseBody, createUnexpectedErrorImpl(ErrorCodeAPICallFailure, err)}
}

// APIClientError represents an database query failure error interface
type APIClientError interface {
	UnexpectedError
	GetAPIURL() string
	GetHTTPStatusCode() *int
	GetHTTPResponseBody() *string
}

type APIClientErrorImpl struct {
	apiURL           string
	httpStatusCode   *int
	httpResponseBody *string
	unexpectedErrorImpl
}

// GetAPIURL gets API URL
func (e *APIClientErrorImpl) GetAPIURL() string {
	return e.apiURL
}

// GetHttpStatusCode gets http status code
func (e *APIClientErrorImpl) GetHTTPStatusCode() *int {
	return e.httpStatusCode
}

// GetHTTPResponseBody gets http status code
func (e *APIClientErrorImpl) GetHTTPResponseBody() *string {
	return e.httpResponseBody
}
