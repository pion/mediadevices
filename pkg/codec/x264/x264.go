// Package x264 implements H264 encoder.
// This package requires libx264 headers and libraries to be built.
// Reference: https://code.videolan.org/videolan/x264/blob/master/example.c
package x264

// #cgo pkg-config: x264
// #include "bridge.h"
import "C"
import (
	"errors"
	"fmt"
	"image"
	"io"
	"sync"
	"unsafe"

	"github.com/pion/mediadevices/pkg/codec"
	mio "github.com/pion/mediadevices/pkg/io"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/webrtc/v2"
)

const (
	// maxRF is a limit for x264 compression level
	// TODO: Probably remove this hardcoded value.
	//       I only saw that 51 was also hardcoded in their source.
	maxRF      = 51
	maxQuality = 10
)

type encoder struct {
	engine *C.Encoder
	buff   []byte
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

func init() {
	codec.Register(webrtc.H264, codec.VideoEncoderBuilder(newEncoder))
}

func newEncoder(r video.Reader, p prop.Media) (io.ReadCloser, error) {
	if p.KeyFrameInterval == 0 {
		p.KeyFrameInterval = 60
	}

	quality := 5
	switch cp := p.CodecParams.(type) {
	case nil:
	case Params:
		quality = cp.Quality
	default:
		return nil, errors.New("unsupported CodecParams type")
	}

	var rf C.float
	// first reverse quality value since rf is the inverse of Quality,
	// and add 1 to map [1,maxQuality] to [maxQuality-1,0]
	rf = C.float(maxQuality - quality)
	// Then, map to x264 RF range, [0,maxQuality-1] to [1,maxRF].
	// [1,maxRF] because 0 is lossless and constrained baseline doesn't support
	// lossless.
	rf *= (maxRF - 1)
	rf /= (maxQuality - 1)
	rf++

	param := C.x264_param_t{
		i_csp:        C.X264_CSP_I420,
		i_width:      C.int(p.Width),
		i_height:     C.int(p.Height),
		i_keyint_max: C.int(p.KeyFrameInterval),
	}
	param.rc.f_rf_constant = rf

	var rc C.int
	engine := C.enc_new(param, &rc)
	if err := errFromC(rc); err != nil {
		return nil, err
	}

	e := encoder{
		engine: engine,
		r:      video.ToI420(r),
	}
	return &e, nil
}

func (e *encoder) Read(p []byte) (int, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return 0, io.EOF
	}

	if e.buff != nil {
		n, err := mio.Copy(p, e.buff)
		if err == nil {
			e.buff = nil
		}
		return n, err
	}

	img, err := e.r.Read()
	if err != nil {
		return 0, err
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
		return 0, err
	}

	encoded := C.GoBytes(unsafe.Pointer(s.data), s.data_len)
	n, err := mio.Copy(p, encoded)
	if err != nil {
		e.buff = encoded
	}
	return n, err
}

func (e *encoder) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	var rc C.int
	C.enc_close(e.engine, &rc)
	e.closed = true
	return nil
}
