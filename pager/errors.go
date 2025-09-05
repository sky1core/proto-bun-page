package pager

import "fmt"

type PagerError struct {
    Code    string
    Message string
    Details map[string]interface{}
}

func (e *PagerError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func NewInvalidRequestError(msg string) *PagerError {
	return &PagerError{
		Code:    "INVALID_REQUEST",
		Message: msg,
	}
}

func NewInternalError(msg string) *PagerError {
    return &PagerError{
        Code:    "INTERNAL_ERROR",
        Message: msg,
    }
}

func NewStaleCursorError() *PagerError {
    return &PagerError{
        Code:    "STALE_CURSOR",
        Message: "stale cursor",
    }
}
