package video

import (
	"errors"
	"image"

	"golang.org/x/image/draw"
)

var (
	scalerTestAlgos = map[string]Scaler{
		"NearestNeighbor": ScalerNearestNeighbor,
	}
	scalerBenchAlgos = map[string]Scaler{
		"NearestNeighbor": ScalerNearestNeighbor,
		"ApproxBiLinear":  ScalerApproxBiLinear,
		"BiLinear":        ScalerBiLinear,
	}
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
		scalerCached := ScalerNearestNeighbor
		if scaler != nil {
			scalerCached = scaler
		}

		// cacheScaler makes cached version of Scaler if available.
		cacheScaler := func(dRect, sRect image.Rectangle) {}
		if kernel, ok := scaler.(interface {
			NewScaler(int, int, int, int) draw.Scaler
		}); ok {
			cacheScaler = func(dRect, sRect image.Rectangle) {
				scalerCached = kernel.NewScaler(dRect.Dx(), dRect.Dy(), sRect.Dx(), sRect.Dy())
			}
		}

		var rect image.Rectangle
		var imgScaled image.Image
		if width > 0 && height > 0 {
			rect = image.Rect(0, 0, width, height)
		} else if width <= 0 && height <= 0 {
			panic("Both width and height are negative!")
		}

		// updateRect updates output image size
		updateRect := func(r image.Rectangle) {}
		if height <= 0 {
			updateRect = func(r image.Rectangle) {
				h := r.Dy() * width / r.Dx()
				rect = image.Rect(0, 0, width, h)
			}
		} else if width <= 0 {
			updateRect = func(r image.Rectangle) {
				w := r.Dx() * height / r.Dy()
				rect = image.Rect(0, 0, w, height)
			}
		}

		// fixedRect returns Rectangle of chroma plane
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

		src := &rgbLikeYCbCr{y: &image.Gray{}, cb: &image.Gray{}, cr: &image.Gray{}}
		dst := &rgbLikeYCbCr{y: &image.Gray{}, cb: &image.Gray{}, cr: &image.Gray{}}

		// ycbcrRealloc reallocs image.YCbCr if needed
		ycbcrRealloc := func(i1 *image.YCbCr) {
			if imgScaled == nil || imgScaled.ColorModel() != i1.ColorModel() {
				updateRect(i1.Rect)
				cacheScaler(rect, i1.Rect)
				imgScaled = image.NewYCbCr(rect, i1.SubsampleRatio)
			}
			imgDst := imgScaled.(*image.YCbCr)
			if imgDst.Rect.Dx() != i1.Rect.Dx() || imgDst.Rect.Dy() != i1.Rect.Dy() {
				updateRect(i1.Rect)
				cacheScaler(rect, i1.Rect)
			}

			yDx := rect.Dx()
			yDy := rect.Dy()
			cRect := fixedRect(rect, i1.SubsampleRatio)
			cDx := cRect.Dx()
			cDy := cRect.Dx()
			yLen := yDx * yDy
			cLen := cDx * cDy
			if len(imgDst.Y) < yLen {
				if cap(imgDst.Y) < yLen {
					imgDst.Y = make([]uint8, yLen)
				}
				imgDst.Y = imgDst.Y[:yLen]
			}
			if len(imgDst.Cr) < cLen {
				if cap(imgDst.Cr) < cLen {
					imgDst.Cr = make([]uint8, cLen)
				}
				imgDst.Cr = imgDst.Cr[:cLen]
			}
			if len(imgDst.Cb) < cLen {
				if cap(imgDst.Cb) < cLen {
					imgDst.Cb = make([]uint8, cLen)
				}
				imgDst.Cb = imgDst.Cb[:cLen]
			}
			*dst.y = image.Gray{Pix: imgDst.Y, Stride: imgDst.YStride, Rect: rect}
			*dst.cb = image.Gray{Pix: imgDst.Cb, Stride: imgDst.CStride, Rect: cRect}
			*dst.cr = image.Gray{Pix: imgDst.Cr, Stride: imgDst.CStride, Rect: cRect}
		}
		// ycbcrRealloc reallocs image.RGBA if needed
		rgbaRealloc := func(i1 *image.RGBA) {
			if imgScaled == nil || imgScaled.ColorModel() != i1.ColorModel() {
				updateRect(i1.Rect)
				cacheScaler(rect, i1.Rect)
				imgScaled = image.NewRGBA(rect)
			}
			imgDst := imgScaled.(*image.RGBA)
			if imgDst.Rect.Dx() != i1.Rect.Dx() || imgDst.Rect.Dy() != i1.Rect.Dy() {
				updateRect(i1.Rect)
				cacheScaler(rect, i1.Rect)
			}

			l := 4 * rect.Dx() * rect.Dy()
			if len(imgDst.Pix) < l {
				if cap(imgDst.Pix) < l {
					imgDst.Pix = make([]uint8, l)
				}
				imgDst.Pix = imgDst.Pix[:l]
			}
		}

		return ReaderFunc(func() (image.Image, func(), error) {
			img, _, err := r.Read()
			if err != nil {
				return nil, func() {}, err
			}

			switch v := img.(type) {
			case *image.RGBA:
				rgbaRealloc(v)
				dst := imgScaled.(*image.RGBA)
				scalerCached.Scale(dst, rect, v, v.Rect, draw.Src, nil)

				cloned := *dst // clone metadata
				return &cloned, func() {}, nil

			case *image.YCbCr:
				ycbcrRealloc(v)
				// Scale each plane
				*src.y = image.Gray{Pix: v.Y, Stride: v.YStride, Rect: v.Rect}
				*src.cb = image.Gray{
					Pix: v.Cb, Stride: v.CStride, Rect: fixedRect(v.Rect, v.SubsampleRatio),
				}
				*src.cr = image.Gray{
					Pix: v.Cr, Stride: v.CStride, Rect: fixedRect(v.Rect, v.SubsampleRatio),
				}
				scalerCached.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Src, nil)

				cloned := *(imgScaled.(*image.YCbCr)) // clone metadata
				return &cloned, func() {}, nil

			default:
				return nil, func() {}, errUnsupportedImageType
			}
		})
	}
}
