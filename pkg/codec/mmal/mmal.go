// Package mmal implements a hardware accelerated H264 encoder for raspberry pi.
// This package requires libmmal headers and libraries to be built.
// Reference: https://github.com/raspberrypi/userland/tree/master/interface/mmal
package mmal

// #cgo pkg-config: mmal
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
	engine C.Encoder
	r      video.Reader
	mu     sync.Mutex
	closed bool
	cntr   int
}

func statusToErr(status *C.Status) error {
	return fmt.Errorf("(status = %d) %s", int(status.code), C.GoString(status.msg))
}

func newEncoder(r video.Reader, p prop.Media, params Params) (codec.ReadCloser, error) {
	if params.KeyFrameInterval == 0 {
		params.KeyFrameInterval = 60
	}

	if params.BitRate == 0 {
		params.BitRate = 300000
	}

	e := encoder{
		r: video.ToI420(r),
	}
	status := C.enc_new(C.Params{
		width:              C.int(p.Width),
		height:             C.int(p.Height),
		bitrate:            C.uint(params.BitRate),
		key_frame_interval: C.uint(params.KeyFrameInterval),
	}, &e.engine)
	if status.code != 0 {
		return nil, statusToErr(&status)
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
	imgReal := img.(*image.YCbCr)
	var y, cb, cr C.Slice
	y.data = (*C.uchar)(&imgReal.Y[0])
	y.len = C.int(len(imgReal.Y))
	cb.data = (*C.uchar)(&imgReal.Cb[0])
	cb.len = C.int(len(imgReal.Cb))
	cr.data = (*C.uchar)(&imgReal.Cr[0])
	cr.len = C.int(len(imgReal.Cr))

	var encodedBuffer *C.MMAL_BUFFER_HEADER_T
	status := C.enc_encode(&e.engine, y, cb, cr, &encodedBuffer)
	if status.code != 0 {
		return nil, func() {}, statusToErr(&status)
	}

	// GoBytes copies the C array to Go slice. After this, it's safe to release the C array
	encoded := C.GoBytes(unsafe.Pointer(encodedBuffer.data), C.int(encodedBuffer.length))
	// Release the buffer so that mmal can reuse this memory
	C.mmal_buffer_header_release(encodedBuffer)

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

	e.closed = true
	C.enc_close(&e.engine)
	return nil
}
