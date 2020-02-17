// +build cgo

package video

import (
	"image/color"
	"testing"
)

func diffUint8(a, b uint8) uint8 {
	d := int(a) - int(b)
	if d < 0 {
		d = -d
	}
	return uint8(d)
}

func TestRGBToYCbCr(t *testing.T) {
	for r := 0; r < 0x100; r += 8 {
		for g := 0; g < 0x100; g += 8 {
			for b := 0; b < 0x100; b += 8 {
				var y, cb, cr uint8
				rgbToYCbCrCGO(&y, &cb, &cr, uint8(r), uint8(g), uint8(b))
				y2, cb2, cr2 := color.RGBToYCbCr(uint8(r), uint8(g), uint8(b))
				if diffUint8(y, y2) > 2 || diffUint8(cb, cb2) > 2 || diffUint8(cr, cr2) > 2 {
					t.Fatalf(
						"rgbToYCbCrCGO differs from color.RGBToYCbCr: in(%d, %d, %d), "+
							"expected(%d, %d, %d), got(%d, %d, %d)",
						r, g, b,
						y2, cb2, cr2,
						y, cb, cr,
					)
				}
			}
		}
	}
}

func TestYCbCrToRGB(t *testing.T) {
	for y := 0; y < 0x100; y += 8 {
		for cb := 0; cb < 0x100; cb += 8 {
			for cr := 0; cr < 0x100; cr += 8 {
				var r, g, b uint8
				yCbCrToRGBCGO(&r, &g, &b, uint8(y), uint8(cb), uint8(cr))
				r2, g2, b2 := color.YCbCrToRGB(uint8(y), uint8(cb), uint8(cr))
				if diffUint8(r, r2) > 2 || diffUint8(g, g2) > 2 || diffUint8(b, b2) > 2 {
					t.Fatalf(
						"yCbCrToRGBCGO differs from color.YCbCrToRGB: in(%d, %d, %d), "+
							"expected(%d, %d, %d), got(%d, %d, %d)",
						y, cb, cr,
						r2, g2, b2,
						r, g, b,
					)
				}
			}
		}
	}
}

func BenchmarkRGBToYCbCr(b *testing.B) {
	b.Run("Go", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, _ = color.RGBToYCbCr(200, 100, 10)
		}
	})
	b.Run("CGO", func(b *testing.B) {
		var y, cb, cr uint8
		// Loop in C code To ignore CGO call overhead
		repeatRGBToYCbCrCGO(b.N, &y, &cb, &cr, 200, 100, 10)
	})
}

func BenchmarkYCbCrToRGB(b *testing.B) {
	b.Run("Go", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, _ = color.YCbCrToRGB(200, 100, 10)
		}
	})
	b.Run("CGO", func(b *testing.B) {
		var rr, gg, bb uint8
		// Loop in C code To ignore CGO call overhead
		repeatYCbCrToRGBCGO(b.N, &rr, &gg, &bb, 200, 100, 10)
	})
}
