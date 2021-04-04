package error

import (
	"errors"

	"gorm.io/gorm"
)

// NewDatabaseError creates a new database error
func NewDatabaseError(err error) DatabaseError {
	return &databaseErrorImpl{createUnexpectedErrorImpl(ErrorCodeDatabaseFailure, err)}
}

// DatabaseError represents an database query failure error interface
type DatabaseError interface {
	UnexpectedError
	IsRecordNotFoundError() bool
}

type databaseErrorImpl struct {
	unexpectedErrorImpl
}

func (e *databaseErrorImpl) IsRecordNotFoundError() bool {
	return errors.Is(e.cause, gorm.ErrRecordNotFound)
}
