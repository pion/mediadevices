// +build cgo

package video

import (
	"image"
)

// #include "convert_cgo.h"
import "C"

// CGO version of the functions will be selected at runtime.
// All functions switched at runtime must be declared also in convert_nocgo.go.
const hasCGOConvert = true

func i444ToI420(img *image.YCbCr) {
	h := img.Rect.Dy()
	C.i444ToI420CGO(
		(*C.uchar)(&img.Cb[0]), (*C.uchar)(&img.Cr[0]),
		C.int(img.CStride), C.int(h),
	)
	img.CStride = img.CStride / 2
	cLen := img.CStride * (h / 2)
	img.Cb = img.Cb[:cLen]
	img.Cr = img.Cr[:cLen]
}

func i422ToI420(img *image.YCbCr) {
	h := img.Rect.Dy()
	C.i422ToI420CGO(
		(*C.uchar)(&img.Cb[0]), (*C.uchar)(&img.Cr[0]),
		C.int(img.CStride), C.int(h),
	)
	cLen := img.CStride * (h / 2)
	img.Cb = img.Cb[:cLen]
	img.Cr = img.Cr[:cLen]
}

func rgbToYCbCrCGO(y, cb, cr *uint8, r, g, b uint8) { // For testing
	C.rgbToYCbCrCGO(
		(*C.uchar)(y), (*C.uchar)(cb), (*C.uchar)(cr),
		C.uchar(r), C.uchar(g), C.uchar(b),
	)
}

func repeatRGBToYCbCrCGO(n int, y, cb, cr *uint8, r, g, b uint8) { // For testing
	C.repeatRGBToYCbCrCGO(
		C.int(n),
		(*C.uchar)(y), (*C.uchar)(cb), (*C.uchar)(cr),
		C.uchar(r), C.uchar(g), C.uchar(b),
	)
}

func yCbCrToRGBCGO(r, g, b *uint8, y, cb, cr uint8) { // For testing
	C.yCbCrToRGBCGO(
		(*C.uchar)(r), (*C.uchar)(g), (*C.uchar)(b),
		C.uchar(y), C.uchar(cb), C.uchar(cr),
	)
}

func repeatYCbCrToRGBCGO(n int, r, g, b *uint8, y, cb, cr uint8) { // For testing
	C.repeatYCbCrToRGBCGO(
		C.int(n),
		(*C.uchar)(r), (*C.uchar)(g), (*C.uchar)(b),
		C.uchar(y), C.uchar(cb), C.uchar(cr),
	)
}

func i444ToRGBA(dst *image.RGBA, src *image.YCbCr) {
	C.i444ToRGBACGO(
		(*C.uchar)(&dst.Pix[0]),
		(*C.uchar)(&src.Y[0]),
		(*C.uchar)(&src.Cb[0]),
		(*C.uchar)(&src.Cr[0]),
		C.int(src.Rect.Dx()),
		C.int(src.Rect.Dy()),
	)
}

func rgbaToI444(dst *image.YCbCr, src *image.RGBA) {
	C.rgbaToI444(
		(*C.uchar)(&dst.Y[0]),
		(*C.uchar)(&dst.Cb[0]),
		(*C.uchar)(&dst.Cr[0]),
		(*C.uchar)(&src.Pix[0]),
		C.int(src.Rect.Dx()),
		C.int(src.Rect.Dy()),
	)
}
