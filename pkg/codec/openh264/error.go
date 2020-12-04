package openh264

import (
	"fmt"
)

import "C"

type eResult int

const (
	retSuccess      eResult = 0
	retFailed       eResult = -1
	retInvalidParam eResult = -2
	retOutOfMemory  eResult = -3
	retNotSupported eResult = -4
	retUnexpected   eResult = -5
	retNeedReinit   eResult = -6
)

func (e eResult) Error() string {
	switch e {
	case retFailed:
		return "failed"
	case retInvalidParam:
		return "invalid param"
	case retOutOfMemory:
		return "out of memory"
	case retNotSupported:
		return "not supported"
	case retUnexpected:
		return "unexpected"
	case retNeedReinit:
		return "need reinit"
	default:
		return fmt.Sprintf("unknown error (%d)", e)
	}
}

func errResult(e C.int) error {
	if e == 0 {
		return nil
	}
	return eResult(e)
}
