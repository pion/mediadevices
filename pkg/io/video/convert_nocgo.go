// +build !cgo

package video

import (
	"image"
	"image/color"
)

const hasCGOConvert = false

func i444ToI420(img *image.YCbCr) {
	h := img.Rect.Dy()
	addrSrc0 := 0
	addrSrc1 := img.CStride
	addrDst := 0
	for i := 0; i < h/2; i++ {
		for j := 0; j < img.CStride/2; j++ {
			cb := uint16(img.Cb[addrSrc0]) + uint16(img.Cb[addrSrc1]) +
				uint16(img.Cb[addrSrc0+1]) + uint16(img.Cb[addrSrc1+1])
			cr := uint16(img.Cr[addrSrc0]) + uint16(img.Cr[addrSrc1]) +
				uint16(img.Cr[addrSrc0+1]) + uint16(img.Cr[addrSrc1+1])
			img.Cb[addrDst] = uint8(cb / 4)
			img.Cr[addrDst] = uint8(cr / 4)
			addrSrc0 += 2
			addrSrc1 += 2
			addrDst++
		}
		addrSrc0 += img.CStride
		addrSrc1 += img.CStride
	}
	img.CStride = img.CStride / 2
	cLen := img.CStride * (h / 2)
	img.Cb = img.Cb[:cLen]
	img.Cr = img.Cr[:cLen]
}

func i422ToI420(img *image.YCbCr) {
	h := img.Rect.Dy()
	addrSrc := 0
	addrDst := 0
	for i := 0; i < h/2; i++ {
		for j := 0; j < img.CStride; j++ {
			cb := uint16(img.Cb[addrSrc]) + uint16(img.Cb[addrSrc+img.CStride])
			cr := uint16(img.Cr[addrSrc]) + uint16(img.Cr[addrSrc+img.CStride])
			img.Cb[addrDst] = uint8(cb / 2)
			img.Cr[addrDst] = uint8(cr / 2)
			addrDst++
			addrSrc++
		}
		addrSrc += img.CStride
	}
	cLen := img.CStride * (h / 2)
	img.Cb = img.Cb[:cLen]
	img.Cr = img.Cr[:cLen]
}

func i444ToRGBA(dst *image.RGBA, src *image.YCbCr) {
	dx := src.Rect.Dx()
	dy := src.Rect.Dy()
	i := 0
	j := 0
	for yi := 0; yi < dy; yi++ {
		for xi := 0; xi < dx; xi++ {
			r, g, b := color.YCbCrToRGB(src.Y[j], src.Cb[j], src.Cr[j])
			dst.Pix[i+0] = uint8(r)
			dst.Pix[i+1] = uint8(g)
			dst.Pix[i+2] = uint8(b)
			dst.Pix[i+3] = 0xff
			i += 4
		}
	}
}

func rgbaToI444(dst *image.YCbCr, src *image.RGBA) {
	i := 0
	addr := 0
	dx := src.Rect.Dx()
	dy := src.Rect.Dy()
	for yi := 0; yi < dy; yi++ {
		for xi := 0; xi < dx; xi++ {
			dst.Y[i], dst.Cb[i], dst.Cr[i] = color.RGBToYCbCr(
				src.Pix[addr+0], src.Pix[addr+1], src.Pix[addr+2],
			)
			addr += 4
			i++
		}
	}
}
