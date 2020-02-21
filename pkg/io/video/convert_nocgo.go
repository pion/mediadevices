// +build !cgo

package video

import (
	"image"
)

func i444ToI420CGO(img *image.YCbCr) {
	panic("CGO was disabled on build")
}

func i422ToI420CGO(img *image.YCbCr) {
	panic("CGO was disabled on build")
}

func i444ToRGBACGO(dst *image.RGBA, src *image.YCbCr) {
	panic("CGO was disabled on build")
}

func rgbaToI444CGO(dst *image.YCbCr, src *image.RGBA) {
	panic("CGO was disabled on build")
}
