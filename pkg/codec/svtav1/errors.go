package svtav1

import "errors"

// #cgo pkg-config: SvtAv1Enc
// #include "bridge.h"
import "C"

var (
	ErrUnknownErrorCode = errors.New("unknown error code")
	ErrInitEncHandler   = errors.New("failed to initialize encoder handler")
	ErrSetEncParam      = errors.New("failed to set encoder parameters")
	ErrEncInit          = errors.New("failed to initialize encoder")
	ErrSendPicture      = errors.New("failed to send picture")
	ErrGetPacket        = errors.New("failed to get packet")
)

func errFromC(ret C.int) error {
	switch ret {
	case 0:
		return nil
	case C.ERR_INIT_ENC_HANDLER:
		return ErrInitEncHandler
	case C.ERR_SET_ENC_PARAM:
		return ErrSetEncParam
	case C.ERR_ENC_INIT:
		return ErrEncInit
	case C.ERR_SEND_PICTURE:
		return ErrSendPicture
	case C.ERR_GET_PACKET:
		return ErrGetPacket
	default:
		return ErrUnknownErrorCode
	}
}
