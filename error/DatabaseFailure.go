package error

import "github.com/jinzhu/gorm"

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
	return gorm.IsRecordNotFoundError(e.cause)
}
