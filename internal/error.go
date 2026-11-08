package internal

import "fmt"

func Err(err error) *InternalError {
	return &InternalError{err: err}
}

type InternalError struct {
	err error
}

func (e *InternalError) Error() string {
	return fmt.Sprintf("internal error: %s", e.err.Error())
}

func (e *InternalError) Unwrap() error {
	return e.err
}

