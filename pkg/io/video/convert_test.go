package video

import (
	"image"
	"reflect"
	"testing"
)

var imageSizes = map[string][2]int{
	"480p":  [2]int{720, 480},
	"1080p": [2]int{1920, 1080},
}

func TestToI420(t *testing.T) {
	cases := map[string]struct {
		src      image.Image
		expected image.Image
	}{
		"I444": {
			src: &image.YCbCr{
				SubsampleRatio: image.YCbCrSubsampleRatio444,
				Y: []uint8{
					0xF0, 0x10, 0x00, 0x00,
					0x00, 0x00, 0x40, 0x00,
					0x00, 0x00, 0x00, 0x00,
					0x00, 0x80, 0x30, 0x00,
				},
				Cb: []uint8{
					0xF0, 0xF0, 0x80, 0x80,
					0xF0, 0xF0, 0x80, 0x80,
					0x80, 0x80, 0x30, 0x30,
					0x80, 0x80, 0x30, 0x30,
				},
				Cr: []uint8{
					0x10, 0x10, 0x40, 0x40,
					0x10, 0x10, 0x40, 0x40,
					0x80, 0x80, 0x80, 0x80,
					0x80, 0x80, 0x80, 0x80,
				},
				YStride: 4,
				CStride: 4,
				Rect:    image.Rect(0, 0, 4, 4),
			},
			expected: &image.YCbCr{
				SubsampleRatio: image.YCbCrSubsampleRatio420,
				Y: []uint8{
					0xF0, 0x10, 0x00, 0x00,
					0x00, 0x00, 0x40, 0x00,
					0x00, 0x00, 0x00, 0x00,
					0x00, 0x80, 0x30, 0x00,
				},
				Cb: []uint8{
					0xF0, 0x80,
					0x80, 0x30,
				},
				Cr: []uint8{
					0x10, 0x40,
					0x80, 0x80,
				},
				YStride: 4,
				CStride: 2,
				Rect:    image.Rect(0, 0, 4, 4),
			},
		},
		"I422": {
			src: &image.YCbCr{
				SubsampleRatio: image.YCbCrSubsampleRatio422,
				Y: []uint8{
					0xF0, 0x10, 0x00, 0x00,
					0x00, 0x00, 0x40, 0x00,
					0x00, 0x00, 0x00, 0x00,
					0x00, 0x80, 0x30, 0x00,
				},
				Cb: []uint8{
					0xF0, 0x80,
					0xF0, 0x80,
					0x80, 0x30,
					0x80, 0x30,
				},
				Cr: []uint8{
					0x10, 0x40,
					0x10, 0x40,
					0x80, 0x80,
					0x80, 0x80,
				},
				YStride: 4,
				CStride: 2,
				Rect:    image.Rect(0, 0, 4, 4),
			},
			expected: &image.YCbCr{
				SubsampleRatio: image.YCbCrSubsampleRatio420,
				Y: []uint8{
					0xF0, 0x10, 0x00, 0x00,
					0x00, 0x00, 0x40, 0x00,
					0x00, 0x00, 0x00, 0x00,
					0x00, 0x80, 0x30, 0x00,
				},
				Cb: []uint8{
					0xF0, 0x80,
					0x80, 0x30,
				},
				Cr: []uint8{
					0x10, 0x40,
					0x80, 0x80,
				},
				YStride: 4,
				CStride: 2,
				Rect:    image.Rect(0, 0, 4, 4),
			},
		},
		"RGBA": {
			src: &image.RGBA{
				Pix: []uint8{
					0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0xFF,
					0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0xFF,
				},
				Stride: 16,
				Rect:   image.Rect(0, 0, 4, 4),
			},
			expected: &image.YCbCr{
				SubsampleRatio: image.YCbCrSubsampleRatio420,
				Y: []uint8{
					0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00,
					0xFF, 0xFF, 0x00, 0x00,
					0xFF, 0xFF, 0x00, 0x00,
				},
				Cb: []uint8{
					0x80, 0x80,
					0x80, 0x80,
				},
				Cr: []uint8{
					0x80, 0x80,
					0x80, 0x80,
				},
				YStride: 4,
				CStride: 2,
				Rect:    image.Rect(0, 0, 4, 4),
			},
		},
	}
	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			r := ToI420(ReaderFunc(func() (image.Image, func(), error) {
				return c.src, func() {}, nil
			}))
			out, _, err := r.Read()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !reflect.DeepEqual(c.expected, out) {
				t.Errorf("Expected output image:\n%v\ngot:\n%v", c.expected, out)
			}
		})
	}
}

func TestToRGBA(t *testing.T) {
	cases := map[string]struct {
		src      image.Image
		expected image.Image
	}{
		"I444": {
			src: &image.YCbCr{
				SubsampleRatio: image.YCbCrSubsampleRatio420,
				Y: []uint8{
					0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00,
					0xFF, 0xFF, 0x00, 0x00,
					0xFF, 0xFF, 0x00, 0x00,
				},
				Cb: []uint8{
					0x80, 0x80,
					0x80, 0x80,
				},
				Cr: []uint8{
					0x80, 0x80,
					0x80, 0x80,
				},
				YStride: 4,
				CStride: 2,
				Rect:    image.Rect(0, 0, 4, 4),
			},
			expected: &image.RGBA{
				Pix: []uint8{
					0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0xFF,
					0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0xFF,
				},
				Stride: 16,
				Rect:   image.Rect(0, 0, 4, 4),
			},
		},
	}
	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			r := ToRGBA(ReaderFunc(func() (image.Image, func(), error) {
				return c.src, func() {}, nil
			}))
			out, _, err := r.Read()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !reflect.DeepEqual(c.expected, out) {
				t.Errorf("Expected output image:\n%v\ngot:\n%v", c.expected, out)
			}
		})
	}
}

func BenchmarkToI420(b *testing.B) {
	for name, sz := range imageSizes {
		cases := map[string]image.Image{
			"I444": image.NewYCbCr(image.Rect(0, 0, sz[0], sz[1]), image.YCbCrSubsampleRatio444),
			"I422": image.NewYCbCr(image.Rect(0, 0, sz[0], sz[1]), image.YCbCrSubsampleRatio422),
			"I420": image.NewYCbCr(image.Rect(0, 0, sz[0], sz[1]), image.YCbCrSubsampleRatio420),
			"RGBA": image.NewRGBA(image.Rect(0, 0, sz[0], sz[1])),
		}
		b.Run(name, func(b *testing.B) {
			for name, img := range cases {
				img := img
				b.Run(name, func(b *testing.B) {
					r := ToI420(ReaderFunc(func() (image.Image, func(), error) {
						return img, func() {}, nil
					}))

					for i := 0; i < b.N; i++ {
						_, _, err := r.Read()
						if err != nil {
							b.Fatalf("Unexpected error: %v", err)
						}
					}
				})
			}
		})
	}
}

func BenchmarkToRGBA(b *testing.B) {
	for name, sz := range imageSizes {
		cases := map[string]image.Image{
			"I444": image.NewYCbCr(image.Rect(0, 0, sz[0], sz[1]), image.YCbCrSubsampleRatio444),
			"I422": image.NewYCbCr(image.Rect(0, 0, sz[0], sz[1]), image.YCbCrSubsampleRatio422),
			"I420": image.NewYCbCr(image.Rect(0, 0, sz[0], sz[1]), image.YCbCrSubsampleRatio420),
			"RGBA": image.NewRGBA(image.Rect(0, 0, sz[0], sz[1])),
		}
		b.Run(name, func(b *testing.B) {
			for name, img := range cases {
				img := img
				b.Run(name, func(b *testing.B) {
					r := ToRGBA(ReaderFunc(func() (image.Image, func(), error) {
						return img, func() {}, nil
					}))

					for i := 0; i < b.N; i++ {
						_, _, err := r.Read()
						if err != nil {
							b.Fatalf("Unexpected error: %v", err)
						}
					}
				})
			}
		})
	}
}
