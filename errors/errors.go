package errors

// IsUnexpectedError returns whether the given error is an UnexpectedError
func IsUnexpectedError(err error) bool {
	_, ok := err.(UnexpectedError)
	return ok
}

// IsValidationError returns whether the given error is an ValidationError
func IsValidationError(err error) bool {
	_, ok := err.(ValidationError)
	return ok
}
