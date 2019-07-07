package model

import (
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
func ValidateParams(args ...Param) error {
	errors := make(map[string]string)
	for _, arg := range args {
		if (arg.v) == "" {
			errors[arg.k] = "Key_Required"
		}
	}

	if len(errors) > 0 {
		return web.NewValidationError("Key_InvalidFields", errors)
	}

	return nil
}

// Param instances are passed to ValidateParams to check empty values
type Param struct {
	k string
	v string
}
