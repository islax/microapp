package errors

import (
	"fmt"
	"runtime"
	"strconv"
)

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

// NewUnexpectedError creates a new unexpected error
func NewUnexpectedError(errCode string, err error) UnexpectedError {
	return createUnexpectedErrorImpl(errCode, err)
}

// NewDataReadWriteError creates a new read write error
func NewDataReadWriteError(err error) UnexpectedError {
	return NewUnexpectedError("Key_ReadWriteFailure", err)
}

// NewCryptoError creates a new encryptions / decryptions error
func NewCryptoError(err error) UnexpectedError {
	return NewUnexpectedError("Key_CryptoFailure", err)
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
