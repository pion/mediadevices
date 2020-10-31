// +build !cgo

package frame

import (
	"fmt"
	"image"
	"sync"
)

func decodeYUY2() decoderFunc {
	pool := sync.Pool{
		New: func() interface{} {
			return &buffer{
				image: &image.YCbCr{},
			}
		},
	}

	return func(frame []byte, width, height int) (image.Image, func(), error) {
		buff := pool.Get().(*buffer)

		yi := width * height
		ci := yi / 2
		fi := yi + 2*ci

		if len(frame) != fi {
			pool.Put(buff)
			return nil, func() {}, fmt.Errorf("frame length (%d) less than expected (%d)", len(frame), fi)
		}

		if len(buff.raw) < fi {
			need := fi - len(buff.raw)
			buff.raw = append(buff.raw, make([]uint8, need)...)
		}
		y := buff.raw[:yi:yi]
		cb := buff.raw[yi : yi+ci : yi+ci]
		cr := buff.raw[yi+ci : fi : fi]

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

		img := buff.image.(*image.YCbCr)
		img.Y = y
		img.YStride = width
		img.Cb = cb
		img.Cr = cr
		img.CStride = width / 2
		img.SubsampleRatio = image.YCbCrSubsampleRatio422
		img.Rect.Min.X = 0
		img.Rect.Min.Y = 0
		img.Rect.Max.X = width
		img.Rect.Max.Y = height
		return img, func() { pool.Put(buff) }, nil
	}
}

func decodeUYVY() decoderFunc {
	return func(frame []byte, width, height int) (image.Image, func(), error) {
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
}
