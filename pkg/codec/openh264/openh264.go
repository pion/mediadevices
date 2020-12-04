package openh264

// #include <string.h>
// #include <openh264/codec_api.h>
// #include <errno.h>
// #include "bridge.hpp"
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

func newEncoder(r video.Reader, p prop.Media, params Params) (codec.ReadCloser, error) {
	if params.BitRate == 0 {
		params.BitRate = 100000
	}

	var rv C.int
	cEncoder := C.enc_new(C.EncoderOptions{
		width:          C.int(p.Width),
		height:         C.int(p.Height),
		target_bitrate: C.int(params.BitRate),
		max_fps:        C.float(p.FrameRate),
	}, &rv)
	if err := errResult(rv); err != nil {
		return nil, fmt.Errorf("failed in creating encoder: %v", err)
	}

	return &encoder{
		engine: cEncoder,
		r:      video.ToI420(r),
	}, nil
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
	bounds := yuvImg.Bounds()
	var rv C.int
	s := C.enc_encode(e.engine, C.Frame{
		y:      unsafe.Pointer(&yuvImg.Y[0]),
		u:      unsafe.Pointer(&yuvImg.Cb[0]),
		v:      unsafe.Pointer(&yuvImg.Cr[0]),
		height: C.int(bounds.Max.Y - bounds.Min.Y),
		width:  C.int(bounds.Max.X - bounds.Min.X),
	}, &rv)
	if err := errResult(rv); err != nil {
		return nil, func() {}, fmt.Errorf("failed in encoding: %v", err)
	}

	encoded := C.GoBytes(unsafe.Pointer(s.data), s.data_len)
	return encoded, func() {}, nil
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

	e.closed = true

	var rv C.int
	C.enc_free(e.engine, &rv)
	return errResult(rv)
}
