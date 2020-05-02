package error

// NewHTTPResourceNotFound creates an new instance of HTTP Error
func NewHTTPResourceNotFound(resourceName, resourceValue string) HTTPResourceNotFound {
	return HTTPResourceNotFound{"Key_ResourceNotFound", resourceName, resourceValue}
}

// HTTPResourceNotFound represents HTTP 404 error
type HTTPResourceNotFound struct {
	ErrorKey      string `json:"errorKey"`
	ResourceName  string `json:"resourceName"`
	ResourceValue string `json:"resourceValue"`
}

// Error returns the error string
func (e HTTPResourceNotFound) Error() string {
	return e.ErrorKey
}
