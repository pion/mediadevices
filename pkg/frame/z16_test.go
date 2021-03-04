package frame

import (
	"image"
	"image/color"
	"reflect"
	"testing"
)

func TestDecodeZ16(t *testing.T) {
	const (
		width  = 2
		height = 3
	)
	decoder, err := NewDecoder(FormatZ16)
	if err != nil {
		t.Fatal(err)
	}
	img, _, err := decoder.Decode([]byte{0x00}, width, height)
	if err == nil {
		t.Errorf("expected to get a frame length mismatch")
	}

	input := []byte{
		0x0c, 0x00, 0x20, 0x03,
		0xa3, 0x01, 0x10, 0x00,
		0x56, 0x09, 0x5d, 0x00,
	}
	expected := image.NewGray16(image.Rect(0, 0, width, height))
	expected.Stride = width * 2
	expected.SetGray16(0, 0, color.Gray16{Y: 12})
	expected.SetGray16(1, 0, color.Gray16{Y: 800})
	expected.SetGray16(0, 1, color.Gray16{Y: 419})
	expected.SetGray16(1, 1, color.Gray16{Y: 16})
	expected.SetGray16(0, 2, color.Gray16{Y: 2390})
	expected.SetGray16(1, 2, color.Gray16{Y: 93})

	img, _, err = decoder.Decode(input, width, height)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, img) {
		t.Errorf("Wrong decode result,\nexpected:\n%+v\ngot:\n%+v", expected, img)
	}
}
