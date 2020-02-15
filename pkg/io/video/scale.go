package video

import (
	"errors"
	"image"

	"golang.org/x/image/draw"
)

// Scaler represents scaling algorithm
type Scaler draw.Scaler

// List of scaling algorithms
var (
	ScalerNearestNeighbor = Scaler(draw.NearestNeighbor)
	ScalerApproxBiLinear  = Scaler(draw.ApproxBiLinear)
	ScalerBiLinear        = Scaler(draw.BiLinear)
	ScalerCatmullRom      = Scaler(draw.CatmullRom)
)

var errUnsupportedImageType = errors.New("scaling: unsupported image type")

// Scale returns video scaling transform.
// Setting scaler=nil to use default scaler. (ScalerNearestNeighbor)
// Negative width or height value will keep the aspect ratio of incoming image.
//
// Note: computation cost to scale YCbCr format is 10 times higher than RGB
// due to the implementation in x/image/draw package.
func Scale(width, height int, scaler Scaler) TransformFunc {
	return func(r Reader) Reader {
		if scaler == nil {
			scaler = ScalerNearestNeighbor
		}

		var rect image.Rectangle
		var imgScaled image.Image
		if width > 0 && height > 0 {
			rect = image.Rect(0, 0, width, height)
		} else if width <= 0 && height <= 0 {
			panic("Both width and height are negative!")
		}

		ycbcrNeedRealloc := func(i1 *image.YCbCr, i2 image.Image) bool {
			if i2 == nil {
				return true
			}
			dst, ok := i2.(*image.YCbCr)
			if !ok || i1.SubsampleRatio != dst.SubsampleRatio {
				return true
			}
			return false
		}
		fixedRect := func(rect image.Rectangle, sr image.YCbCrSubsampleRatio) image.Rectangle {
			switch sr {
			case image.YCbCrSubsampleRatio444:
			case image.YCbCrSubsampleRatio422:
				rect.Max.X /= 2
			case image.YCbCrSubsampleRatio420:
				rect.Max.X /= 2
				rect.Max.Y /= 2
			}
			return rect
		}

		planes := [3]struct {
			src *image.Gray
			dst *image.Gray
		}{}
		for i := range planes {
			planes[i].src, planes[i].dst = &image.Gray{}, &image.Gray{}
		}
		src := &rgbLikeYCbCr{y: planes[0].src, cb: planes[1].src, cr: planes[2].src}
		dst := &rgbLikeYCbCr{y: planes[0].dst, cb: planes[1].dst, cr: planes[2].dst}

		return ReaderFunc(func() (image.Image, error) {
			img, err := r.Read()
			if err != nil {
				return nil, err
			}

			if imgScaled == nil {
				if height <= 0 {
					h := img.Bounds().Dy() * width / img.Bounds().Dx()
					rect = image.Rect(0, 0, width, h)
				} else if width <= 0 {
					w := img.Bounds().Dx() * height / img.Bounds().Dy()
					rect = image.Rect(0, 0, w, height)
				}
			}

			switch v := img.(type) {
			case *image.RGBA:
				if imgScaled == nil || imgScaled.ColorModel() != img.ColorModel() {
					imgScaled = image.NewRGBA(rect)
				}
				dst := imgScaled.(*image.RGBA)
				scaler.Scale(dst, rect, img, img.Bounds(), draw.Over, nil)

			case *image.YCbCr:
				if ycbcrNeedRealloc(v, imgScaled) {
					imgNew := image.NewYCbCr(rect, v.SubsampleRatio)
					imgScaled = imgNew
					*dst.y = image.Gray{Pix: imgNew.Y, Stride: imgNew.YStride, Rect: rect}
					*dst.cb = image.Gray{
						Pix: imgNew.Cb, Stride: imgNew.CStride, Rect: fixedRect(rect, v.SubsampleRatio),
					}
					*dst.cr = image.Gray{
						Pix: imgNew.Cr, Stride: imgNew.CStride, Rect: fixedRect(rect, v.SubsampleRatio),
					}
				}
				// Scale each plane
				*src.y = image.Gray{Pix: v.Y, Stride: v.YStride, Rect: v.Rect}
				*src.cb = image.Gray{
					Pix: v.Cb, Stride: v.CStride, Rect: fixedRect(v.Rect, v.SubsampleRatio),
				}
				*src.cr = image.Gray{
					Pix: v.Cr, Stride: v.CStride, Rect: fixedRect(v.Rect, v.SubsampleRatio),
				}
				scaler.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

			default:
				return nil, errUnsupportedImageType
			}

			return imgScaled, nil
		})
	}
}
