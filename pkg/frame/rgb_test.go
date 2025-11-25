package frame

import (
	"fmt"
	"testing"
)

func BenchmarkDecodeBGRA(b *testing.B) {
	sizes := []struct {
		width, height int
	}{
		{640, 480},
		{1920, 1080},
	}
	for _, sz := range sizes {
		sz := sz
		b.Run(fmt.Sprintf("%dx%d", sz.width, sz.height), func(b *testing.B) {
			input := make([]byte, sz.width*sz.height*4)
			for i := 0; i < b.N; i++ {
				_, _, err := decodeBGRA(input, sz.width, sz.height)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkDecodeARGB(b *testing.B) {
	sizes := []struct {
		width, height int
	}{
		{640, 480},
		{1920, 1080},
	}
	for _, sz := range sizes {
		sz := sz
		b.Run(fmt.Sprintf("%dx%d", sz.width, sz.height), func(b *testing.B) {
			input := make([]byte, sz.width*sz.height*4)
			for i := 0; i < b.N; i++ {
				_, _, err := decodeARGB(input, sz.width, sz.height)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
