package frame

import (
	"fmt"
	"image"
)

func decodeI420(frame []byte, width, height int) (image.Image, func(), error) {
	yi := width * height
	cbi := yi + width*height/4
	cri := cbi + width*height/4

	if cri > len(frame) {
		return nil, func() {}, fmt.Errorf("frame length (%d) less than expected (%d)", len(frame), cri)
	}

	return &image.YCbCr{
		Y:              frame[:yi],
		YStride:        width,
		Cb:             frame[yi:cbi],
		Cr:             frame[cbi:cri],
		CStride:        width / 2,
		SubsampleRatio: image.YCbCrSubsampleRatio420,
		Rect:           image.Rect(0, 0, width, height),
	}, func() {}, nil
}

func decodeNV21(frame []byte, width, height int) (image.Image, func(), error) {
	yi := width * height
	ci := yi + width*height/2

	if ci > len(frame) {
		return nil, func() {}, fmt.Errorf("frame length (%d) less than expected (%d)", len(frame), ci)
	}

	var cb, cr []byte
	for i := yi; i < ci; i += 2 {
		cr = append(cr, frame[i])
		cb = append(cb, frame[i+1])
	}

	return &image.YCbCr{
		Y:              frame[:yi],
		YStride:        width,
		Cb:             cb,
		Cr:             cr,
		CStride:        width / 2,
		SubsampleRatio: image.YCbCrSubsampleRatio420,
		Rect:           image.Rect(0, 0, width, height),
	}, func() {}, nil
}

func decodeNV12(frame []byte, width, height int) (image.Image, func(), error) {
	img, release, err := decodeNV21(frame, width, height)
	if err != nil {
		return img, release, err
	}

	// The only difference between NV21 and NV12 is the chroma order, so simply swap them
	yuv := img.(*image.YCbCr)
	yuv.Cb, yuv.Cr = yuv.Cr, yuv.Cb
	return yuv, release, err
}
