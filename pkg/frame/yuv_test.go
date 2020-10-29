package frame

import (
	"fmt"
	"image"
	"reflect"
	"testing"
)

func TestDecodeYUY2(t *testing.T) {
	const (
		width  = 2
		height = 2
	)
	input := []byte{
		// Y    Cb     Y    Cr
		0x01, 0x82, 0x03, 0x84,
		0x05, 0x86, 0x07, 0x88,
	}
	expected := &image.YCbCr{
		Y:              []byte{0x01, 0x03, 0x05, 0x07},
		YStride:        width,
		Cb:             []byte{0x82, 0x86},
		Cr:             []byte{0x84, 0x88},
		CStride:        width / 2,
		SubsampleRatio: image.YCbCrSubsampleRatio422,
		Rect:           image.Rect(0, 0, width, height),
	}

	img, _, err := decodeYUY2(input, width, height)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, img) {
		t.Errorf("Wrong decode result,\nexpected:\n%+v\ngot:\n%+v", expected, img)
	}
}

func TestDecodeUYVY(t *testing.T) {
	const (
		width  = 2
		height = 2
	)
	input := []byte{
		//Cb     Y    Cr     Y
		0x82, 0x01, 0x84, 0x03,
		0x86, 0x05, 0x88, 0x07,
	}
	expected := &image.YCbCr{
		Y:              []byte{0x01, 0x03, 0x05, 0x07},
		YStride:        width,
		Cb:             []byte{0x82, 0x86},
		Cr:             []byte{0x84, 0x88},
		CStride:        width / 2,
		SubsampleRatio: image.YCbCrSubsampleRatio422,
		Rect:           image.Rect(0, 0, width, height),
	}

	img, _, err := decodeUYVY(input, width, height)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, img) {
		t.Errorf("Wrong decode result,\nexpected:\n%+v\ngot:\n%+v", expected, img)
	}
}

func BenchmarkDecodeYUY2(b *testing.B) {
	sizes := []struct {
		width, height int
	}{
		{640, 480},
		{1920, 1080},
	}
	for _, sz := range sizes {
		sz := sz
		b.Run(fmt.Sprintf("%dx%d", sz.width, sz.height), func(b *testing.B) {
			input := make([]byte, sz.width*sz.height*2)
			for i := 0; i < b.N; i++ {
				_, _, err := decodeYUY2(input, sz.width, sz.height)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
