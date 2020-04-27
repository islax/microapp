package errors

import (
	"fmt"
)

// ValidationError is an error indicating error in validations
type ValidationError struct {
	ErrorKey string            `json:"errorKey"`
	Errors   map[string]string `json:"errors"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("Error: [%s - %s]", e.ErrorKey, e.Errors)
}

// NewValidationError creates an new instance of Validation Error
func NewValidationError(err string, failedValidations map[string]string) ValidationError {
	return ValidationError{ErrorKey: err, Errors: failedValidations}
}
