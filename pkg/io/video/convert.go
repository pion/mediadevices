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
		b := make([]uint8, i2)
		dst.Y = b[:i0:i0]
		dst.Cb = b[i0:i1:i1]
		dst.Cr = b[i1:i2:i2]
	}
	dst.SubsampleRatio = image.YCbCrSubsampleRatio444
	dst.YStride = dx
	dst.CStride = dx
	dst.Rect = bounds

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
			for i := 0; i < h/2; i++ {
				addrSrc := i * 2 * yuvImg.CStride
				addrDst := i * yuvImg.CStride / 2
				for j := 0; j < yuvImg.CStride/2; j++ {
					j2 := j * 2
					cb := uint16(yuvImg.Cb[addrSrc+j2]) + uint16(yuvImg.Cb[addrSrc+yuvImg.CStride+j2]) +
						uint16(yuvImg.Cb[addrSrc+j2+1]) + uint16(yuvImg.Cb[addrSrc+yuvImg.CStride+j2+1])
					cr := uint16(yuvImg.Cr[addrSrc+j2]) + uint16(yuvImg.Cr[addrSrc+yuvImg.CStride+j2]) +
						uint16(yuvImg.Cr[addrSrc+j2+1]) + uint16(yuvImg.Cr[addrSrc+yuvImg.CStride+j2+1])
					yuvImg.Cb[addrDst+j] = uint8(cb / 4)
					yuvImg.Cr[addrDst+j] = uint8(cr / 4)
				}
			}
			yuvImg.CStride = yuvImg.CStride / 2
			cLen := yuvImg.CStride * (h / 2)
			yuvImg.Cb = yuvImg.Cb[:cLen:cLen]
			yuvImg.Cr = yuvImg.Cr[:cLen:cLen]
		case image.YCbCrSubsampleRatio422:
			for i := 0; i < h/2; i++ {
				addrSrc := i * 2 * yuvImg.CStride
				addrDst := i * yuvImg.CStride
				for j := 0; j < yuvImg.CStride; j++ {
					cb := uint16(yuvImg.Cb[addrSrc+j]) + uint16(yuvImg.Cb[addrSrc+yuvImg.CStride+j])
					cr := uint16(yuvImg.Cr[addrSrc+j]) + uint16(yuvImg.Cr[addrSrc+yuvImg.CStride+j])
					yuvImg.Cb[addrDst+j] = uint8(cb / 2)
					yuvImg.Cr[addrDst+j] = uint8(cr / 2)
				}
			}
			cLen := yuvImg.CStride * (h / 2)
			yuvImg.Cb = yuvImg.Cb[:cLen:cLen]
			yuvImg.Cr = yuvImg.Cr[:cLen:cLen]
		case image.YCbCrSubsampleRatio420:
		default:
			return nil, fmt.Errorf("unsupported pixel format: %s", yuvImg.SubsampleRatio)
		}

		yuvImg.SubsampleRatio = image.YCbCrSubsampleRatio420
		return &yuvImg, nil
	})
}
