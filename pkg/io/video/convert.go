package video

import (
	"fmt"
	"image"
)

// ToI420 converts r to a new reader that will output images in I420 format
func ToI420(r Reader) Reader {
	return ReaderFunc(func() (image.Image, error) {
		img, err := r.Read()
		if err != nil {
			return nil, err
		}

		// TODO: Not sure how to handle this when it's not YCbCr, maybe try to convert it to YCvCr?
		yuvImg := img.(*image.YCbCr)
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
		return yuvImg, nil
	})
}
