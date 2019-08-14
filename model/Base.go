package model

import (
	"regexp"
	"time"

	"github.com/islax/microapp/web"
	uuid "github.com/satori/go.uuid"
)

// Base contains common columns for all tables.
type Base struct {
	ID        uuid.UUID  `gorm:"type:varchar(36);primary_key;"`
	CreatedAt time.Time  `gorm:"column:createdOn"`
	UpdatedAt time.Time  `gorm:"column:modifiedOn"`
	DeletedAt *time.Time `sql:"index" gorm:"column:deletedOn"`
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

// ValidateString checks whether the given string conforms to the given constraint. Valid constraints are ANC - Alphanumeric, ANH - Alphanumeric & hyphen, URL - URL, EML - Email.
// If the given constraint doesnot match any of the predefined constraint then it will be treated as a regular expression.
func ValidateString(value string, constraint string) (bool, error) {
	var regularExpression *regexp.Regexp
	var err error
	switch constraint {
	case "ANC":
		regularExpression, _ = regexp.Compile("^[A-Za-z0-9]+$")
	case "ANH":
		regularExpression, _ = regexp.Compile("^[A-Za-z0-9-]+$")
	case "URL":
		regularExpression, _ = regexp.Compile("(http:\\/\\/www\\.|https:\\/\\/www\\.|http:\\/\\/|https:\\/\\/)?[a-z0-9]+([\\-\\.]{1}[a-z0-9]+)*\\.[a-z]{2,5}(:[0-9]{1,5})?(\\/.*)?$")
	case "EML":
		regularExpression, _ = regexp.Compile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	default:
		regularExpression, err = regexp.Compile(constraint)
		if err != nil {
			return false, err
		}
	}

	return regularExpression.MatchString(value), nil
}
