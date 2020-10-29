// Package x264 implements H264 encoder.
// This package requires libx264 headers and libraries to be built.
// Reference: https://code.videolan.org/videolan/x264/blob/master/example.c
package x264

// #cgo pkg-config: x264
// #include "bridge.h"
import "C"
import (
	"fmt"
	"image"
	"io"
	"sync"
	"unsafe"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

type encoder struct {
	engine *C.Encoder
	r      video.Reader
	mu     sync.Mutex
	closed bool
}

type cerror int

func (e cerror) Error() string {
	switch e {
	case C.ERR_DEFAULT_PRESET:
		return errDefaultPreset.Error()
	case C.ERR_APPLY_PROFILE:
		return errApplyProfile.Error()
	case C.ERR_ALLOC_PICTURE:
		return errAllocPicture.Error()
	case C.ERR_OPEN_ENGINE:
		return errOpenEngine.Error()
	case C.ERR_ENCODE:
		return errEncode.Error()
	default:
		return "unknown error"
	}
}

func errFromC(rc C.int) error {
	if rc == 0 {
		return nil
	}
	return cerror(rc)
}

var (
	errInitEngine    = fmt.Errorf("failed to initialize x264")
	errDefaultPreset = fmt.Errorf("failed to set default preset")
	errApplyProfile  = fmt.Errorf("failed to apply profile")
	errAllocPicture  = fmt.Errorf("failed to alloc picture")
	errOpenEngine    = fmt.Errorf("failed to open x264")
	errEncode        = fmt.Errorf("failed to encode")
)

func newEncoder(r video.Reader, p prop.Media, params Params) (codec.ReadCloser, error) {
	if params.KeyFrameInterval == 0 {
		params.KeyFrameInterval = 60
	}

	// Convert from bit/s to kbit/s because x264 uses kbit/s instead.
	// Reference: https://code.videolan.org/videolan/x264/-/blob/7923c5818b50a3d8816eed222a7c43b418a73b36/encoder/ratecontrol.c#L657
	params.BitRate /= 1000

	param := C.x264_param_t{
		i_csp:        C.X264_CSP_I420,
		i_width:      C.int(p.Width),
		i_height:     C.int(p.Height),
		i_keyint_max: C.int(params.KeyFrameInterval),
	}
	param.rc.i_bitrate = C.int(params.BitRate)
	param.rc.i_vbv_max_bitrate = param.rc.i_bitrate
	param.rc.i_vbv_buffer_size = param.rc.i_vbv_max_bitrate * 2

	var rc C.int
	// cPreset will be freed in C.enc_new
	cPreset := C.CString(fmt.Sprint(params.Preset))
	engine := C.enc_new(param, cPreset, &rc)
	if err := errFromC(rc); err != nil {
		return nil, err
	}

	e := encoder{
		engine: engine,
		r:      video.ToI420(r),
	}
	return &e, nil
}

func (e *encoder) Read() ([]byte, func(), error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return nil, func() {}, io.EOF
	}

	img, _, err := e.r.Read()
	if err != nil {
		return nil, func() {}, err
	}
	yuvImg := img.(*image.YCbCr)

	var rc C.int
	s := C.enc_encode(
		e.engine,
		(*C.uchar)(&yuvImg.Y[0]),
		(*C.uchar)(&yuvImg.Cb[0]),
		(*C.uchar)(&yuvImg.Cr[0]),
		&rc,
	)
	if err := errFromC(rc); err != nil {
		return nil, func() {}, err
	}

	encoded := C.GoBytes(unsafe.Pointer(s.data), s.data_len)
	return encoded, func() {}, err
}

func (e *encoder) SetBitRate(b int) error {
	panic("SetBitRate is not implemented")
}

func (e *encoder) ForceKeyFrame() error {
	panic("ForceKeyFrame is not implemented")
}

func (e *encoder) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return nil
	}

	var rc C.int
	C.enc_close(e.engine, &rc)
	e.closed = true
	return nil
}
