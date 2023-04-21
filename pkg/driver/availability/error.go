package availability

import (
	"errors"
)

type Error interface {
	Error() string
}

var ErrUnimplemented Error = errors.New("not implemented")
var ErrBusy Error = errors.New("device or resource busy")
var ErrNoDevice Error = errors.New("no such device")

type errorString struct {
	s string
}

func NewError(text string) Error {
	return &errorString{text}
}

func (e *errorString) Error() string {
	return e.s
}
