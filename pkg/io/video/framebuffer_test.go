package video

import (
	"image"
	"math/rand"
	"reflect"
	"testing"
)

func randomize(arr []uint8) {
	for i := range arr {
		arr[i] = uint8(rand.Uint32())
	}
}

func BenchmarkFrameBufferCopyOptimized(b *testing.B) {
	frameBuffer := NewFrameBuffer(0)
	resolution := image.Rect(0, 0, 1920, 1080)
	src := image.NewYCbCr(resolution, image.YCbCrSubsampleRatio420)

	for i := 0; i < b.N; i++ {
		frameBuffer.StoreCopy(src)
	}
}

func BenchmarkFrameBufferCopyNaive(b *testing.B) {
	resolution := image.Rect(0, 0, 1920, 1080)
	src := image.NewYCbCr(resolution, image.YCbCrSubsampleRatio420)
	var dst image.Image

	for i := 0; i < b.N; i++ {
		clone := *src
		clone.Cb = make([]uint8, len(src.Cb))
		clone.Cr = make([]uint8, len(src.Cr))
		clone.Y = make([]uint8, len(src.Y))

		copy(clone.Cb, src.Cb)
		copy(clone.Cr, src.Cr)
		copy(clone.Y, src.Y)
		dst = &clone
		_ = dst
	}
}

func TestFrameBufferStoreCopyAndLoad(t *testing.T) {
	resolution := image.Rect(0, 0, 16, 8)
	rgbaLike := image.NewRGBA64(resolution)
	randomize(rgbaLike.Pix)
	testCases := map[string]struct {
		New    func() image.Image
		Update func(image.Image)
	}{
		"Alpha": {
			New: func() image.Image {
				return (*image.Alpha)(rgbaLike)
			},
			Update: func(src image.Image) {
				img := src.(*image.Alpha)
				randomize(img.Pix)
			},
		},
		"Alpha16": {
			New: func() image.Image {
				return (*image.Alpha16)(rgbaLike)
			},
			Update: func(src image.Image) {
				img := src.(*image.Alpha16)
				randomize(img.Pix)
			},
		},
		"CMYK": {
			New: func() image.Image {
				return (*image.CMYK)(rgbaLike)
			},
			Update: func(src image.Image) {
				img := src.(*image.CMYK)
				randomize(img.Pix)
			},
		},
		"Gray": {
			New: func() image.Image {
				return (*image.Gray)(rgbaLike)
			},
			Update: func(src image.Image) {
				img := src.(*image.Gray)
				randomize(img.Pix)
			},
		},
		"Gray16": {
			New: func() image.Image {
				return (*image.Gray16)(rgbaLike)
			},
			Update: func(src image.Image) {
				img := src.(*image.Gray16)
				randomize(img.Pix)
			},
		},
		"NRGBA": {
			New: func() image.Image {
				return (*image.NRGBA)(rgbaLike)
			},
			Update: func(src image.Image) {
				img := src.(*image.NRGBA)
				randomize(img.Pix)
			},
		},
		"NRGBA64": {
			New: func() image.Image {
				return (*image.NRGBA64)(rgbaLike)
			},
			Update: func(src image.Image) {
				img := src.(*image.NRGBA64)
				randomize(img.Pix)
			},
		},
		"RGBA": {
			New: func() image.Image {
				return (*image.RGBA)(rgbaLike)
			},
			Update: func(src image.Image) {
				img := src.(*image.RGBA)
				randomize(img.Pix)
			},
		},
		"RGBA64": {
			New: func() image.Image {
				return (*image.RGBA64)(rgbaLike)
			},
			Update: func(src image.Image) {
				img := src.(*image.RGBA64)
				randomize(img.Pix)
			},
		},
		"NYCbCrA": {
			New: func() image.Image {
				img := image.NewNYCbCrA(resolution, image.YCbCrSubsampleRatio420)
				randomize(img.Y)
				randomize(img.Cb)
				randomize(img.Cr)
				randomize(img.A)
				img.CStride = 10
				img.YStride = 5
				return img
			},
			Update: func(src image.Image) {
				img := src.(*image.NYCbCrA)
				randomize(img.Y)
				randomize(img.Cb)
				randomize(img.Cr)
				randomize(img.A)
				img.CStride = 3
				img.YStride = 2
			},
		},
		"YCbCr": {
			New: func() image.Image {
				img := image.NewYCbCr(resolution, image.YCbCrSubsampleRatio420)
				randomize(img.Y)
				randomize(img.Cb)
				randomize(img.Cr)
				img.CStride = 10
				img.YStride = 5
				return img
			},
			Update: func(src image.Image) {
				img := src.(*image.YCbCr)
				randomize(img.Y)
				randomize(img.Cb)
				randomize(img.Cr)
				img.CStride = 3
				img.YStride = 2
			},
		},
	}

	frameBuffer := NewFrameBuffer(0)

	for name, testCase := range testCases {
		// Since the test also wants to make sure that Copier can convert from 1 type to another,
		// t.Run is not ideal since it'll run the tests separately
		t.Log("Testing", name)

		src := testCase.New()
		frameBuffer.StoreCopy(src)
		if !reflect.DeepEqual(frameBuffer.Load(), src) {
			t.Fatal("Expected the copied image to be identical with the source")
		}

		testCase.Update(src)
		frameBuffer.StoreCopy(src)
		if !reflect.DeepEqual(frameBuffer.Load(), src) {
			t.Fatal("Expected the copied image to be identical with the source after an update in source")
		}
	}
}
