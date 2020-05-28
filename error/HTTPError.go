package error

// NewHTTPError creates an new instance of HTTP Error
func NewHTTPError(err string, httpStatus int) HTTPError {
	return HTTPError{ErrorKey: err, HTTPStatus: httpStatus}
}

// HTTPError Represent an error to be sent back on repsonse
type HTTPError struct {
	HTTPStatus int
	ErrorKey   string
}

// Error returns the error string
func (err HTTPError) Error() string {
	return err.ErrorKey
}
