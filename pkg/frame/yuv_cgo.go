// +build cgo

package frame

import (
	"fmt"
	"image"
)

// #include <stdint.h>
// void decodeYUY2CGO(uint8_t* y, uint8_t* cb, uint8_t* cr, uint8_t* yuy2, int width, int height);
// void decodeUYVYCGO(uint8_t* y, uint8_t* cb, uint8_t* cr, uint8_t* uyvy, int width, int height);
import "C"

func decodeYUY2(frame []byte, width, height int) (image.Image, func(), error) {
	yi := width * height
	ci := yi / 2
	fi := yi + 2*ci

	if len(frame) != fi {
		return nil, func() {}, fmt.Errorf("frame length (%d) less than expected (%d)", len(frame), fi)
	}

	y := make([]byte, yi)
	cb := make([]byte, ci)
	cr := make([]byte, ci)

	C.decodeYUY2CGO(
		(*C.uchar)(&y[0]),
		(*C.uchar)(&cb[0]),
		(*C.uchar)(&cr[0]),
		(*C.uchar)(&frame[0]),
		C.int(width), C.int(height),
	)

	return &image.YCbCr{
		Y:              y,
		YStride:        width,
		Cb:             cb,
		Cr:             cr,
		CStride:        width / 2,
		SubsampleRatio: image.YCbCrSubsampleRatio422,
		Rect:           image.Rect(0, 0, width, height),
	}, func() {}, nil
}

func decodeUYVY(frame []byte, width, height int) (image.Image, func(), error) {
	yi := width * height
	ci := yi / 2
	fi := yi + 2*ci

	if len(frame) != fi {
		return nil, func() {}, fmt.Errorf("frame length (%d) less than expected (%d)", len(frame), fi)
	}

	y := make([]byte, yi)
	cb := make([]byte, ci)
	cr := make([]byte, ci)

	C.decodeUYVYCGO(
		(*C.uchar)(&y[0]),
		(*C.uchar)(&cb[0]),
		(*C.uchar)(&cr[0]),
		(*C.uchar)(&frame[0]),
		C.int(width), C.int(height),
	)

	return &image.YCbCr{
		Y:              y,
		YStride:        width,
		Cb:             cb,
		Cr:             cr,
		CStride:        width / 2,
		SubsampleRatio: image.YCbCrSubsampleRatio422,
		Rect:           image.Rect(0, 0, width, height),
	}, func() {}, nil
}
