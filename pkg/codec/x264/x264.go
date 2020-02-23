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
	"unsafe"

	"github.com/pion/mediadevices/pkg/codec"
	mio "github.com/pion/mediadevices/pkg/io"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/webrtc/v2"
)

type encoder struct {
	engine *C.Encoder
	buff   []byte
	r      video.Reader
}

var (
	errInitEngine   = fmt.Errorf("failed to initialize x264")
	errApplyProfile = fmt.Errorf("failed to apply profile")
	errAllocPicture = fmt.Errorf("failed to alloc picture")
	errOpenEngine   = fmt.Errorf("failed to open x264")
	errEncode       = fmt.Errorf("failed to encode")
)

func init() {
	codec.Register(webrtc.H264, codec.VideoEncoderBuilder(newEncoder))
}

func newEncoder(r video.Reader, p prop.Media) (io.ReadCloser, error) {
	if p.KeyFrameInterval == 0 {
		p.KeyFrameInterval = 60
	}

	engine, err := C.enc_new(C.x264_param_t{
		i_csp:        C.X264_CSP_I420,
		i_width:      C.int(p.Width),
		i_height:     C.int(p.Height),
		i_keyint_max: C.int(p.KeyFrameInterval),
	})
	if err != nil {
		return nil, errInitEngine
	}

	e := encoder{
		engine: engine,
		r:      video.ToI420(r),
	}
	return &e, nil
}

func (e *encoder) Read(p []byte) (int, error) {
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
	s := C.enc_encode(
		e.engine,
		(*C.uchar)(&yuvImg.Y[0]),
		(*C.uchar)(&yuvImg.Cb[0]),
		(*C.uchar)(&yuvImg.Cr[0]),
	)
	if s.data_len < 0 {
		return 0, errEncode
	}

	encoded := C.GoBytes(unsafe.Pointer(s.data), s.data_len)
	n, err := mio.Copy(p, encoded)
	if err != nil {
		e.buff = encoded
	}
	return n, err
}

func (e *encoder) Close() error {
	C.enc_close(e.engine)
	return nil
}
