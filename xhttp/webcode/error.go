package webcode

import (
	"fmt"
)

// Error .
func (e *Error) Error() string {
	return e.GetMessage()
}

// NewError .
func NewError(code Code, message interface{}) *Error {
	err := &Error{Code: int32(code)}
	switch message.(type) {
	case error:
		err.Message = message.(error).Error()
	case string:
		err.Message = message.(string)
	default:
		err.Message = fmt.Sprintf("%v", message)
	}
	return err
}
