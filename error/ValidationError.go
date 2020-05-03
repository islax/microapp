package error

import (
	"fmt"
)

// IsValidationError returns whether the given error is an ValidationError
func IsValidationError(err error) bool {
	_, ok := err.(ValidationError)
	return ok
}

// NewInvalidFieldsError creates a new invalid fields validation Error.
// 'failedFieldValidations' - map key should be the name of the field and value should be the error code.
func NewInvalidFieldsError(failedFieldValidations map[string]string) ValidationError {
	return ValidationError{ErrorKey: ErrorCodeInvalidFields, Errors: failedFieldValidations}
}

// NewInvalidRequestPayloadError creates a new invalid request payload validation Error.
func NewInvalidRequestPayloadError(errorCode string) ValidationError {
	return ValidationError{ErrorKey: ErrorCodeInvalidRequestPayload, Errors: map[string]string{"payload": errorCode}}
}

// NewValidationError creates a new invalid fields validation Error.
func NewValidationError(errKey string, failedValidations map[string]string) ValidationError {
	return ValidationError{ErrorKey: errKey, Errors: failedValidations}
}

// ValidationError is an error indicating error in validations
type ValidationError struct {
	ErrorKey string            `json:"errorKey"`
	Errors   map[string]string `json:"errors"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("Error: [%s - %s]", e.ErrorKey, e.Errors)
}
