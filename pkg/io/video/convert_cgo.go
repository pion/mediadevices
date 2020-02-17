// +build cgo

package video

import (
	"image"
)

// void i444ToI420CGO(
//     unsigned char* cb, unsigned char* cr,
//     const int stride, const int h);
// void i422ToI420CGO(
//     unsigned char* cb, unsigned char* cr,
//     const int stride, const int h);
import "C"

func init() {
	hasCGOConvert = true
}

func i444ToI420CGO(img *image.YCbCr) {
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

func i422ToI420CGO(img *image.YCbCr) {
	h := img.Rect.Dy()
	C.i422ToI420CGO(
		(*C.uchar)(&img.Cb[0]), (*C.uchar)(&img.Cr[0]),
		C.int(img.CStride), C.int(h),
	)
	cLen := img.CStride * (h / 2)
	img.Cb = img.Cb[:cLen]
	img.Cr = img.Cr[:cLen]
}
