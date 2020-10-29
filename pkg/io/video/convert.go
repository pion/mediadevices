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
		rgbaToI444(dst, s)
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
	return ReaderFunc(func() (image.Image, func(), error) {
		img, _, err := r.Read()
		if err != nil {
			return nil, func() {}, err
		}

		imageToYCbCr(&yuvImg, img)

		// Covert pixel format to I420
		switch yuvImg.SubsampleRatio {
		case image.YCbCrSubsampleRatio444:
			i444ToI420(&yuvImg)
		case image.YCbCrSubsampleRatio422:
			i422ToI420(&yuvImg)
		case image.YCbCrSubsampleRatio420:
		default:
			return nil, func() {}, fmt.Errorf("unsupported pixel format: %s", yuvImg.SubsampleRatio)
		}

		yuvImg.SubsampleRatio = image.YCbCrSubsampleRatio420
		return &yuvImg, func() {}, nil
	})
}

// imageToRGBA converts src to *image.RGBA and store it to dst
func imageToRGBA(dst *image.RGBA, src image.Image) {
	if dst == nil {
		panic("dst can't be nil")
	}

	if srcRGBA, ok := src.(*image.RGBA); ok {
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

	if srcYCbCr, ok := src.(*image.YCbCr); ok &&
		srcYCbCr.SubsampleRatio == image.YCbCrSubsampleRatio444 {
		i444ToRGBA(dst, srcYCbCr)
		return
	}

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
	return ReaderFunc(func() (image.Image, func(), error) {
		img, _, err := r.Read()
		if err != nil {
			return nil, func() {}, err
		}

		imageToRGBA(&dst, img)
		return &dst, func() {}, nil
	})
}
