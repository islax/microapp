package error

import (
	"fmt"
	"runtime"
	"strconv"
)

// IsUnexpectedError returns whether the given error is an UnexpectedError
func IsUnexpectedError(err error) bool {
	_, ok := err.(UnexpectedError)
	return ok
}

// NewUnexpectedError creates a new unexpected error
func NewUnexpectedError(errCode string, err error) UnexpectedError {
	return createUnexpectedErrorImpl(errCode, err)
}

// NewDataReadWriteError creates a new read write error
func NewDataReadWriteError(err error) UnexpectedError {
	return NewUnexpectedError(ErrorCodeReadWriteFailure, err)
}

// NewCryptoError creates a new encryptions / decryptions error
func NewCryptoError(err error) UnexpectedError {
	return NewUnexpectedError(ErrorCodeCryptoFailure, err)
}

// UnexpectedError represents an unexpected error interface
type UnexpectedError interface {
	Error() string
	GetErrorCode() string
	GetStackTrace() string
	GetCause() error
}

type unexpectedErrorImpl struct {
	errCode    string
	cause      error
	stackTrace string
}

// Error returns the error string
func (e unexpectedErrorImpl) Error() string {
	return fmt.Sprintf("%v:%v", e.errCode, e.cause)
}

// GetErrCode returns the error code
func (e unexpectedErrorImpl) GetCause() error {
	return e.cause
}

// GetErrCode returns the error code
func (e unexpectedErrorImpl) GetErrorCode() string {
	return e.errCode
}

// GetStackTrace returns the error stack trace
func (e unexpectedErrorImpl) GetStackTrace() string {
	return e.stackTrace
}

func createUnexpectedErrorImpl(errCode string, err error) unexpectedErrorImpl {
	const depth = 20
	var ptrs [depth]uintptr
	n := runtime.Callers(2, ptrs[:])
	ptrSlice := ptrs[0:n]
	stack := ""
	for _, pc := range ptrSlice {
		stackFunc := runtime.FuncForPC(pc)
		_, line := stackFunc.FileLine(pc)
		stack = stack + stackFunc.Name() + ":" + strconv.Itoa(line) + "\n"
	}
	return unexpectedErrorImpl{cause: err, stackTrace: stack}
}
