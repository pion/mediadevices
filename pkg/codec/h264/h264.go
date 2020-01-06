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
	"github.com/pion/webrtc/v2"
	"image"
	"unsafe"
)

type h264Encoder struct {
	encoder *C.Encoder
}

var _ codec.VideoEncoder = &h264Encoder{}
var _ codec.VideoEncoderBuilder = codec.VideoEncoderBuilder(NewEncoder)

func init() {
	codec.Register(webrtc.H264, codec.VideoEncoderBuilder(NewEncoder))
}

func NewEncoder(s codec.VideoSetting) (codec.VideoEncoder, error) {
	encoder, err := C.enc_new(C.EncoderOptions{
		width:          C.int(s.Width),
		height:         C.int(s.Height),
		target_bitrate: C.int(s.TargetBitRate),
		max_bitrate:	C.int(s.MaxBitRate),
		max_fps:        C.float(s.FrameRate),
	})
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
