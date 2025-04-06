package availability

import (
	"errors"
)

var (
	ErrUnimplemented = NewError("not implemented")
	ErrBusy          = NewError("device or resource busy")
	ErrNoDevice      = NewError("no such device")
)

type errorString struct {
	s string
}

func NewError(text string) error {
	return &errorString{text}
}

func IsError(err error) bool {
	var target *errorString
	return errors.As(err, &target)
}

func (e *errorString) Error() string {
	return e.s
}
