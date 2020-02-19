package openh264

// #cgo CFLAGS: -I${SRCDIR}/../../../cvendor/include
// #cgo CXXFLAGS: -I${SRCDIR}/../../../cvendor/include
// #cgo LDFLAGS: ${SRCDIR}/../../../cvendor/lib/openh264/libopenh264.a
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
	mio "github.com/pion/mediadevices/pkg/io"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"

	"github.com/pion/webrtc/v2"
)

type encoder struct {
	engine *C.Encoder
	r      video.Reader
	buff   []byte

	mu     sync.Mutex
	closed bool
}

var _ codec.VideoEncoderBuilder = codec.VideoEncoderBuilder(NewEncoder)

func init() {
	codec.Register(webrtc.H264, codec.VideoEncoderBuilder(NewEncoder))
}

func NewEncoder(r video.Reader, p prop.Media) (io.ReadCloser, error) {
	if p.BitRate == 0 {
		p.BitRate = 100000
	}

	cEncoder, err := C.enc_new(C.EncoderOptions{
		width:          C.int(p.Width),
		height:         C.int(p.Height),
		target_bitrate: C.int(p.BitRate),
		max_fps:        C.float(p.FrameRate),
	})
	if err != nil {
		// TODO: better error message
		return nil, fmt.Errorf("failed in creating encoder")
	}

	return &encoder{
		engine: cEncoder,
		r:      video.ToI420(r),
	}, nil
}

func (e *encoder) Read(p []byte) (n int, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return 0, io.EOF
	}

	if e.buff != nil {
		n, err = mio.Copy(p, e.buff)
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
	bounds := yuvImg.Bounds()
	s, err := C.enc_encode(e.engine, C.Frame{
		y:      unsafe.Pointer(&yuvImg.Y[0]),
		u:      unsafe.Pointer(&yuvImg.Cb[0]),
		v:      unsafe.Pointer(&yuvImg.Cr[0]),
		height: C.int(bounds.Max.Y - bounds.Min.Y),
		width:  C.int(bounds.Max.X - bounds.Min.X),
	})
	if err != nil {
		// TODO: better error message
		return 0, fmt.Errorf("failed in encoding")
	}

	encoded := C.GoBytes(unsafe.Pointer(s.data), s.data_len)
	n, err = mio.Copy(p, encoded)
	if err != nil {
		e.buff = encoded
	}

	return n, err
}

func (e *encoder) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.closed = true

	C.enc_free(e.engine)
	return nil
}
