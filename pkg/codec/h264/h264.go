package h264

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
	"github.com/pion/mediadevices/pkg/codec"
	"image"
	"unsafe"
)

type Options struct {
	Width        int
	Height       int
	Bitrate      int
	MaxFrameRate float32
}

// https://github.com/cisco/openh264/wiki/TypesAndStructures#sencparambase
func (h *Options) translate() C.EncoderOptions {
	return C.EncoderOptions{
		width:          C.int(h.Width),
		height:         C.int(h.Height),
		target_bitrate: C.int(h.Bitrate),
		max_fps:        C.float(h.MaxFrameRate),
	}
}

type h264Encoder struct {
	encoder *C.Encoder
}

var _ codec.Encoder = &h264Encoder{}

func NewEncoder(opts Options) (codec.Encoder, error) {
	encoder, err := C.enc_new(opts.translate())
	if err != nil {
		// TODO: better error message
		return nil, fmt.Errorf("failed in creating encoder")
	}

	e := h264Encoder{
		encoder: encoder,
	}
	return &e, nil
}

func (e *h264Encoder) Encode(img image.Image) ([]byte, error) {
	// TODO: Convert img to YCbCr since openh264 only accepts YCbCr
	// TODO: Convert img to 4:2:0 format which what openh264 accepts
	yuvImg := img.(*image.YCbCr)
	bounds := yuvImg.Bounds()
	s, err := C.enc_encode(e.encoder, C.Frame{
		y:      unsafe.Pointer(&yuvImg.Y[0]),
		u:      unsafe.Pointer(&yuvImg.Cb[0]),
		v:      unsafe.Pointer(&yuvImg.Cr[0]),
		height: C.int(bounds.Max.Y - bounds.Min.Y),
		width:  C.int(bounds.Max.X - bounds.Min.X),
	})
	if err != nil {
		// TODO: better error message
		return nil, fmt.Errorf("failed in encoding")
	}

	return C.GoBytes(unsafe.Pointer(s.data), s.data_len), nil
}

func (e *h264Encoder) Close() error {
	C.enc_free(e.encoder)
	return nil
}
