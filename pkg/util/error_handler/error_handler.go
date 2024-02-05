package error_handler

import "fmt"

type JinxError struct {
	ErrorCode int
	Err       error
}

func (e *JinxError) Error() string {
	return fmt.Sprintf("Status %d: %v", e.ErrorCode, e.Err)
}

func NewJinxError(errorCode int, err error) *JinxError {
	return &JinxError{
		ErrorCode: errorCode,
		Err:       err,
	}
}
