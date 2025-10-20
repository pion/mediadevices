package svtav1

import (
	"errors"
)

var (
	ErrUnknownErrorCode = errors.New("unknown error code")
	ErrInitEncHandler   = errors.New("failed to initialize encoder handler")
	ErrSetEncParam      = errors.New("failed to set encoder parameters")
	ErrEncInit          = errors.New("failed to initialize encoder")
)
