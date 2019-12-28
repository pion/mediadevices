package frame

import (
	"fmt"
	"image"
)

func decodeI420(frame []byte, width, height int) (image.Image, error) {
	yi := width * height
	cbi := yi + width*height/4
	cri := cbi + width*height/4

	if cri > len(frame) {
		return nil, fmt.Errorf("frame length (%d) less than expected (%d)", len(frame), cri)
	}

	return &image.YCbCr{
		Y:              frame[:yi],
		YStride:        width,
		Cb:             frame[yi:cbi],
		Cr:             frame[cbi:cri],
		CStride:        width / 2,
		SubsampleRatio: image.YCbCrSubsampleRatio420,
		Rect:           image.Rect(0, 0, width, height),
	}, nil
}

func decodeNV21(frame []byte, width, height int) (image.Image, error) {
	yi := width * height
	ci := yi + width*height/2

	if ci > len(frame) {
		return nil, fmt.Errorf("frame length (%d) less than expected (%d)", len(frame), ci)
	}

	var cb, cr []byte
	for i := yi; i < ci; i += 2 {
		cb = append(cb, frame[i])
		cr = append(cr, frame[i+1])
	}

	return &image.YCbCr{
		Y:              frame[:yi],
		YStride:        width,
		Cb:             cb,
		Cr:             cr,
		CStride:        width / 2,
		SubsampleRatio: image.YCbCrSubsampleRatio420,
		Rect:           image.Rect(0, 0, width, height),
	}, nil
}

func decodeYUY2(frame []byte, width, height int) (image.Image, error) {
	yi := width * height
	ci := yi / 2
	fi := yi + 2*ci

	if len(frame) != fi {
		return nil, fmt.Errorf("frame length (%d) less than expected (%d)", len(frame), fi)
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
	}, nil
}
