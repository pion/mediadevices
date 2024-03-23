package frame

import (
	"fmt"
	"image"
	"math/bits"
	"unsafe"
)

func decodeARGB(frame []byte, width, height int) (image.Image, func(), error) {
	size := 4 * width * height
	if size > len(frame) {
		return nil, func() {}, fmt.Errorf("frame length (%d) less than expected (%d)", len(frame), size)
	}
	r := image.Rect(0, 0, width, height)
	for i := 0; i < size; i += 4 {
		*(*uint32)(unsafe.Pointer(&frame[i])) = func(v uint32) uint32 {
			return (v & 0xFF00FF00) | (v&0xFF)<<16 | (v&0xFF0000)>>16
		}(*(*uint32)(unsafe.Pointer(&frame[i])))
		//frame[i], frame[i+2] = frame[i+2], frame[i]
	}
	return &image.RGBA{
		Pix:    frame[:size:size],
		Stride: 4 * r.Dx(),
		Rect:   r,
	}, func() {}, nil
}

func decodeBGRA(frame []byte, width, height int) (image.Image, func(), error) {
	size := 4 * width * height
	if size > len(frame) {
		return nil, func() {}, fmt.Errorf("frame length (%d) less than expected (%d)", len(frame), size)
	}
	r := image.Rect(0, 0, width, height)
	for i := 0; i < size; i += 4 {
		*(*uint32)(unsafe.Pointer(&frame[i])) = bits.RotateLeft32(*(*uint32)(unsafe.Pointer(&frame[i])), -8)
		//frame[i], frame[i+1], frame[i+2], frame[i+3] = frame[i+1], frame[i+2], frame[i+3], frame[i]
	}
	return &image.RGBA{
		Pix:    frame[:size:size],
		Stride: 4 * r.Dx(),
		Rect:   r,
	}, func() {}, nil
}
