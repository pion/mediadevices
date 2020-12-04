// +build !cgo

package frame

import (
	"fmt"
	"image"
)

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

	fast := 0
	slow := 0
	for i := 0; i < fi; i += 4 {
		y[fast] = frame[i]
		cb[slow] = frame[i+1]
		y[fast+1] = frame[i+2]
		cr[slow] = frame[i+3]
		fast += 2
		slow++
	}

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

	fast := 0
	slow := 0
	for i := 0; i < fi; i += 4 {
		cb[slow] = frame[i]
		y[fast] = frame[i+1]
		cr[slow] = frame[i+2]
		y[fast+1] = frame[i+3]
		fast += 2
		slow++
	}

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
