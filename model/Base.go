package model

import (
	"errors"
	"regexp"
	"strings"
	"time"

	microappError "github.com/islax/microapp/error"
	uuid "github.com/satori/go.uuid"
)

// Base contains common columns for all tables.
type Base struct {
	ID        uuid.UUID  `gorm:"type:varchar(36);primary_key;"`
	CreatedAt time.Time  `gorm:"column:createdOn"`
	UpdatedAt time.Time  `gorm:"column:modifiedOn"`
	DeletedAt *time.Time `sql:"index" gorm:"column:deletedOn"`
}

//FieldData represents the data associated with a field
type FieldData struct {
	Name        string
	Value       interface{}
	Type        string
	Required    bool
	Constraints []*ConstraintDetail
}

//ConstraintDetail respresents constraint detail
type ConstraintDetail struct {
	Type           ConstraintType
	ConstraintData interface{}
}

//NewStringFieldData creates new FieldData with type string and no constraint
func NewStringFieldData(name string, value interface{}) *FieldData {
	return &FieldData{
		Name:     name,
		Value:    value,
		Type:     "string",
		Required: true,
	}
}

// NewStringFieldDataWithConstraint creates new FieldData with type string and constraint
func NewStringFieldDataWithConstraint(name string, value interface{}, required bool, constraints []*ConstraintDetail) *FieldData {
	return &FieldData{
		Name:        name,
		Value:       value,
		Type:        "string",
		Required:    required,
		Constraints: constraints,
	}
}

// ValidateParams checks string parameters passed to it and returns error in case of blank values.
func ValidateParams(params map[string]interface{}) error {
	errors := make(map[string]string)
	for key, value := range params {
		if (value.(string)) == "" {
			errors[key] = microappError.ErrorCodeRequired
		}
	}

	if len(errors) > 0 {
		return microappError.NewInvalidFieldsError(errors)
	}

	return nil
}

// ValidateFields checks string parameters passed to it and returns error in case of blank values.
func ValidateFields(fields []*FieldData) error {
	errors := make(map[string]string)
	for _, field := range fields {
		if field.Type == "string" {
			valAsString, ok := field.Value.(string)
			if !ok {
				errors[field.Name] = microappError.ErrorCodeStringExpected
			} else if field.Required && strings.TrimSpace(valAsString) == "" {
				errors[field.Name] = microappError.ErrorCodeRequired
			} else if strings.TrimSpace(valAsString) != "" && len(field.Constraints) > 0 {
				for _, constraint := range field.Constraints {
					ok, _ = ValidateString(valAsString, constraint.Type, constraint.ConstraintData)
					if !ok {
						errors[field.Name] = microappError.ErrorCodeInvalidValue
					}
				}
			}
		}
	}
	if len(errors) > 0 {
		return microappError.NewInvalidFieldsError(errors)
	}
	return nil
}

// ValidateString checks whether the given data conforms to the given constraint. Valid constraints are AlphaNumeric, AlphaNumericAndHyphen, Email, URL and RegEx.
// If the given constraint is RegEx, then the constraintData should contain a valid regular expression.
func ValidateString(value string, constraint ConstraintType, constraintData interface{}) (bool, error) {
	var regularExpression *regexp.Regexp
	var err error
	switch constraint {
	case AlphaNumeric:
		regularExpression, _ = regexp.Compile("^[A-Za-z0-9]+$")
	case AlphaNumericAndHyphen:
		regularExpression, _ = regexp.Compile("^[A-Za-z0-9-]+$")
	case URL:
		regularExpression, _ = regexp.Compile("(http:\\/\\/www\\.|https:\\/\\/www\\.|http:\\/\\/|https:\\/\\/)?[a-z0-9]+([\\-\\.]{1}[a-z0-9]+)*\\.[a-z]{2,5}(:[0-9]{1,5})?(\\/.*)?$")
	case Email:
		regularExpression, _ = regexp.Compile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	case RegEx:
		if constraintData == nil {
			return false, microappError.NewUnexpectedError(microappError.ErrorCodeRequired, errors.New("If the constraint is 'RegEx', then a valid regex is needed as constraintData"))
		}
		regularExpression, err = regexp.Compile(constraintData.(string))
		if err != nil {
			return false, microappError.NewUnexpectedError(microappError.ErrorCodeInvalidValue, err)
		}
	case In:
		if constraintData == nil {
			return false, microappError.NewUnexpectedError(microappError.ErrorCodeRequired, errors.New("If the constraint is 'In', then a string slice containing valid values is needed as constraintData"))
		}

		validValues, stringSliceType := constraintData.([]string)
		if !stringSliceType {
			return false, microappError.NewUnexpectedError(microappError.ErrorCodeRequired, errors.New("If the constraint is 'In', then a string slice containing valid values is needed as constraintData"))
		}
		for _, validValue := range validValues {
			if value == validValue {
				return true, nil
			}
		}
		return false, nil
	case UUID:
		if _, err := uuid.FromString(value); err != nil {
			return false, nil
		}
		return true, nil
	default:
		return false, nil
	}

	return regularExpression.MatchString(value), nil
}

//ConstraintType represents the type of the string
type ConstraintType string

const (
	// AlphaNumeric represents string containing only alphabets and numbers
	AlphaNumeric ConstraintType = "AlphaNumeric"
	// AlphaNumericAndHyphen represents string containing alphabets, numbers and hyphen
	AlphaNumericAndHyphen ConstraintType = "AlphaNumericAndHyphen"
	// Email represents string containing email address
	Email ConstraintType = "Email"
	// RegEx represents string containing regular expression
	RegEx ConstraintType = "RegEx"
	// In represents string value from a pre-defined set
	In ConstraintType = "In"
	// URL represents string containing URL
	URL ConstraintType = "URL"
	// UUID represents string containing UUID
	UUID ConstraintType = "UUID"
)
