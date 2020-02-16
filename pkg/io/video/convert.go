package video

import (
	"fmt"
	"image"
	"image/color"
)

// imageToYCbCr converts src to *image.YCbCr and store it to dst
// Note: conversion can be lossy
func imageToYCbCr(dst *image.YCbCr, src image.Image) {
	if dst == nil {
		panic("dst can't be nil")
	}

	yuvImg, ok := src.(*image.YCbCr)
	if ok {
		*dst = *yuvImg
		return
	}

	bounds := src.Bounds()
	dy := bounds.Dy()
	dx := bounds.Dx()
	flat := dy * dx

	if len(dst.Y)+len(dst.Cb)+len(dst.Cr) < 3*flat {
		i0 := 1 * flat
		i1 := 2 * flat
		i2 := 3 * flat
		if cap(dst.Y) < i2 {
			dst.Y = make([]uint8, i2)
		}
		dst.Y = dst.Y[:i0]
		dst.Cb = dst.Y[i0:i1]
		dst.Cr = dst.Y[i1:i2]
	}
	dst.SubsampleRatio = image.YCbCrSubsampleRatio444
	dst.YStride = dx
	dst.CStride = dx
	dst.Rect = bounds

	switch s := src.(type) {
	case *image.RGBA:
		i := 0
		addr := 0
		for yi := 0; yi < dy; yi++ {
			for xi := 0; xi < dx; xi++ {
				dst.Y[i], dst.Cb[i], dst.Cr[i] = color.RGBToYCbCr(
					s.Pix[addr+0], s.Pix[addr+1], s.Pix[addr+2],
				)
				addr += 4
				i++
			}
		}
	default:
		i := 0
		for yi := 0; yi < dy; yi++ {
			for xi := 0; xi < dx; xi++ {
				// TODO: probably try to get the alpha value with something like
				// https://en.wikipedia.org/wiki/Alpha_compositing
				r, g, b, _ := src.At(xi, yi).RGBA()
				yy, cb, cr := color.RGBToYCbCr(uint8(r/256), uint8(g/256), uint8(b/256))
				dst.Y[i] = yy
				dst.Cb[i] = cb
				dst.Cr[i] = cr
				i++
			}
		}
	}
}

// ToI420 converts r to a new reader that will output images in I420 format
func ToI420(r Reader) Reader {
	var yuvImg image.YCbCr
	return ReaderFunc(func() (image.Image, error) {
		img, err := r.Read()
		if err != nil {
			return nil, err
		}

		imageToYCbCr(&yuvImg, img)
		h := yuvImg.Rect.Dy()

		// Covert pixel format to I420
		switch yuvImg.SubsampleRatio {
		case image.YCbCrSubsampleRatio444:
			addrSrc0 := 0
			addrSrc1 := yuvImg.CStride
			addrDst := 0
			for i := 0; i < h/2; i++ {
				for j := 0; j < yuvImg.CStride/2; j++ {
					cb := uint16(yuvImg.Cb[addrSrc0]) + uint16(yuvImg.Cb[addrSrc1]) +
						uint16(yuvImg.Cb[addrSrc0+1]) + uint16(yuvImg.Cb[addrSrc1+1])
					cr := uint16(yuvImg.Cr[addrSrc0]) + uint16(yuvImg.Cr[addrSrc1]) +
						uint16(yuvImg.Cr[addrSrc0+1]) + uint16(yuvImg.Cr[addrSrc1+1])
					yuvImg.Cb[addrDst] = uint8(cb / 4)
					yuvImg.Cr[addrDst] = uint8(cr / 4)
					addrSrc0 += 2
					addrSrc1 += 2
					addrDst++
				}
				addrSrc0 += yuvImg.CStride
				addrSrc1 += yuvImg.CStride
			}
			yuvImg.CStride = yuvImg.CStride / 2
			cLen := yuvImg.CStride * (h / 2)
			yuvImg.Cb = yuvImg.Cb[:cLen]
			yuvImg.Cr = yuvImg.Cr[:cLen]
		case image.YCbCrSubsampleRatio422:
			addrSrc := 0
			addrDst := 0
			for i := 0; i < h/2; i++ {
				for j := 0; j < yuvImg.CStride; j++ {
					cb := uint16(yuvImg.Cb[addrSrc]) + uint16(yuvImg.Cb[addrSrc+yuvImg.CStride])
					cr := uint16(yuvImg.Cr[addrSrc]) + uint16(yuvImg.Cr[addrSrc+yuvImg.CStride])
					yuvImg.Cb[addrDst] = uint8(cb / 2)
					yuvImg.Cr[addrDst] = uint8(cr / 2)
					addrDst++
					addrSrc++
				}
				addrSrc += yuvImg.CStride
			}
			cLen := yuvImg.CStride * (h / 2)
			yuvImg.Cb = yuvImg.Cb[:cLen]
			yuvImg.Cr = yuvImg.Cr[:cLen]
		case image.YCbCrSubsampleRatio420:
		default:
			return nil, fmt.Errorf("unsupported pixel format: %s", yuvImg.SubsampleRatio)
		}

		yuvImg.SubsampleRatio = image.YCbCrSubsampleRatio420
		return &yuvImg, nil
	})
}

// imageToRGBA converts src to *image.RGBA and store it to dst
func imageToRGBA(dst *image.RGBA, src image.Image) {
	if dst == nil {
		panic("dst can't be nil")
	}

	srcRGBA, ok := src.(*image.RGBA)
	if ok {
		*dst = *srcRGBA
		return
	}

	bounds := src.Bounds()
	dy := bounds.Dy()
	dx := bounds.Dx()

	if len(dst.Pix) < 4*dx*dy {
		dst.Pix = make([]uint8, 4*dx*dy)
	}
	dst.Stride = 4 * dx
	dst.Rect = bounds

	i := 0
	for yi := 0; yi < dy; yi++ {
		for xi := 0; xi < dx; xi++ {
			r, g, b, a := src.At(xi, yi).RGBA()
			dst.Pix[i+0] = uint8(r / 0x100)
			dst.Pix[i+1] = uint8(g / 0x100)
			dst.Pix[i+2] = uint8(b / 0x100)
			dst.Pix[i+3] = uint8(a / 0x100)
			i += 4
		}
	}
}

// ToRGBA converts r to a new reader that will output images in RGBA format
func ToRGBA(r Reader) Reader {
	var dst image.RGBA
	return ReaderFunc(func() (image.Image, error) {
		img, err := r.Read()
		if err != nil {
			return nil, err
		}

		imageToRGBA(&dst, img)
		return &dst, nil
	})
}
