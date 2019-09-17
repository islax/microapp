package model

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/islax/microapp/web"
	uuid "github.com/satori/go.uuid"
)

// Base contains common columns for all tables.
type Base struct {
	ID        uuid.UUID  `gorm:"type:varchar(36);primary_key;"`
	CreatedAt time.Time  `gorm:"column:createdOn"`
	UpdatedAt time.Time  `gorm:"column:modifiedOn;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
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

//NewStringFieldDataWithConstraint creates new FieldData with type string and constraint
func NewStringFieldDataWithConstraint(name string, value interface{}, constraints []*ConstraintDetail) *FieldData {
	return &FieldData{
		Name:        name,
		Value:       value,
		Type:        "string",
		Required:    true,
		Constraints: constraints,
	}
}

// ValidateParams checks string parameters passed to it and returns error in case of blank values.
func ValidateParams(params map[string]interface{}) error {
	errors := make(map[string]string)
	for key, value := range params {
		if (value.(string)) == "" {
			errors[key] = "Key_Required"
		}
	}

	if len(errors) > 0 {
		return web.NewValidationError("Key_InvalidFields", errors)
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
				errors[field.Name] = "Key_StringExpected"
			} else if field.Required && strings.TrimSpace(valAsString) == "" {
				errors[field.Name] = "Key_Required"
			} else if strings.TrimSpace(valAsString) != "" && len(field.Constraints) > 0 {
				for _, constraint := range field.Constraints {
					if constraint.Type == RegEx {
						ok, _ = ValidateString(valAsString, constraint.Type, constraint.ConstraintData.(string))
					} else {
						ok, _ = ValidateString(valAsString, constraint.Type)
					}
					if !ok {
						errors[field.Name] = "Key_InvalidValue"
					}
				}
			}
		}
	}
	if len(errors) > 0 {
		return web.NewValidationError("Key_InvalidFields", errors)
	}
	return nil
}

// ValidateString checks whether the given data conforms to the given constraint. Valid constraints are AlphaNumeric, AlphaNumericAndHyphen, Email, URL and RegEx.
// If the given constraint is RegEx, then the 3rd parameter should contain a valid regular expression.
func ValidateString(value string, constraint ConstraintType, regex ...string) (bool, error) {
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
		if len(regex) < 1 {
			return false, errors.New("If the constraint is 'RegEx', then a valid regex is needed as 3rd parameter")
		}
		regularExpression, err = regexp.Compile(regex[0])
		if err != nil {
			return false, err
		}
	default:
		return false, nil
	}

	return regularExpression.MatchString(value), nil
}

//ConstraintType represents the type of the string
type ConstraintType string

const (
	//AlphaNumeric represents string containing only alphabets and numbers
	AlphaNumeric ConstraintType = "AlphaNumeric"
	//AlphaNumericAndHyphen represents string containing alphabets, numbers and hyphen
	AlphaNumericAndHyphen ConstraintType = "AlphaNumericAndHyphen"
	//Email represents string containing email address
	Email ConstraintType = "Email"
	//URL represents string containing URL
	URL ConstraintType = "URL"
	//RegEx represents string containing regular expression
	RegEx ConstraintType = "RegEx"
)
