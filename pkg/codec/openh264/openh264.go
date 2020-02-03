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
	"unsafe"

	"github.com/pion/mediadevices/pkg/codec"
	mio "github.com/pion/mediadevices/pkg/io"
	"github.com/pion/mediadevices/pkg/io/video"

	"github.com/pion/webrtc/v2"
)

type encoder struct {
	engine *C.Encoder
	r      video.Reader
	buff   []byte
}

var _ codec.VideoEncoderBuilder = codec.VideoEncoderBuilder(NewEncoder)

func init() {
	codec.Register(webrtc.H264, codec.VideoEncoderBuilder(NewEncoder))
}

func NewEncoder(r video.Reader, prop video.AdvancedProperty) (io.ReadCloser, error) {
	cEncoder, err := C.enc_new(C.EncoderOptions{
		width:          C.int(prop.Width),
		height:         C.int(prop.Height),
		target_bitrate: C.int(prop.BitRate),
		max_fps:        C.float(prop.FrameRate),
	})
	if err != nil {
		// TODO: better error message
		return nil, fmt.Errorf("failed in creating encoder")
	}

	return &encoder{
		engine: cEncoder,
		r:      r,
	}, nil
}

func (e *encoder) Read(p []byte) (n int, err error) {
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

	// TODO: Convert img to YCbCr since openh264 only accepts YCbCr
	// TODO: Convert img to 4:2:0 format which what openh264 accepts
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
	C.enc_free(e.engine)
	return nil
}
