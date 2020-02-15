package video

import (
	"errors"
	"image"
	"sync"

	"golang.org/x/image/draw"
)

type Scaler draw.Scaler

var (
	ScalerNearestNeighbor = Scaler(draw.NearestNeighbor)
	ScalerApproxBiLinear  = Scaler(draw.ApproxBiLinear)
	ScalerBiLinear        = Scaler(draw.BiLinear)
	ScalerCatmullRom      = Scaler(draw.CatmullRom)
)

var errUnsupportedImageType = errors.New("scaling: unsupported image type")

// Scale returns video scaling transform.
// Setting scaler=nil to use default scaler. (ScalerNearestNeighbor)
func Scale(width, height int, scaler Scaler) TransformFunc {
	return func(r Reader) Reader {
		if scaler == nil {
			scaler = ScalerNearestNeighbor
		}

		rect := image.Rect(0, 0, width, height)
		var imgScaled image.Image

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

		planes := [3]struct {
			src *image.Gray
			dst *image.Gray
		}{}
		for i := range planes {
			planes[i].src, planes[i].dst = &image.Gray{}, &image.Gray{}
		}
		scalers := [3]draw.Scaler{}

		return ReaderFunc(func() (image.Image, error) {
			img, err := r.Read()
			if err != nil {
				return nil, err
			}

			switch v := img.(type) {
			case *image.RGBA:
				if imgScaled == nil || imgScaled.ColorModel() != img.ColorModel() {
					imgScaled = image.NewRGBA(rect)
					if s, ok := scaler.(*draw.Kernel); ok {
						// *draw.Kernel has optimized version for size fixed scaling
						dr := rect
						sr := v.Bounds()
						scalers[0] = s.NewScaler(dr.Dx(), dr.Dy(), sr.Dx(), sr.Dy())
					} else {
						scalers[0] = scaler
					}
				}
				dst := imgScaled.(*image.RGBA)
				scalers[0].Scale(dst, rect, img, img.Bounds(), draw.Over, nil)

			case *image.YCbCr:
				if ycbcrNeedRealloc(v, imgScaled) {
					i := image.NewYCbCr(rect, v.SubsampleRatio)
					imgScaled = i
					*planes[0].dst = image.Gray{Pix: i.Y, Stride: i.YStride, Rect: rect}
					*planes[1].dst = image.Gray{Pix: i.Cr, Stride: i.CStride, Rect: rect}
					*planes[2].dst = image.Gray{Pix: i.Cb, Stride: i.CStride, Rect: rect}

					if s, ok := scaler.(*draw.Kernel); ok {
						// *draw.Kernel has optimized version for size fixed scaling
						dr := rect
						sr := v.Bounds()
						scalers[0] = s.NewScaler(dr.Dx(), dr.Dy(), sr.Dx(), sr.Dy())
						switch v.SubsampleRatio {
						case image.YCbCrSubsampleRatio444:
							scalers[1] = s.NewScaler(dr.Dx(), dr.Dy(), sr.Dx(), sr.Dy())
							scalers[2] = s.NewScaler(dr.Dx(), dr.Dy(), sr.Dx(), sr.Dy())
						case image.YCbCrSubsampleRatio422:
							scalers[1] = s.NewScaler(dr.Dx()/2, dr.Dy(), sr.Dx()/2, sr.Dy())
							scalers[2] = s.NewScaler(dr.Dx()/2, dr.Dy(), sr.Dx()/2, sr.Dy())
						case image.YCbCrSubsampleRatio420:
							scalers[1] = s.NewScaler(dr.Dx()/2, dr.Dy()/2, sr.Dx()/2, sr.Dy()/2)
							scalers[2] = s.NewScaler(dr.Dx()/2, dr.Dy()/2, sr.Dx()/2, sr.Dy()/2)
						}
					} else {
						for i := range scalers {
							scalers[i] = scaler
						}
					}

					switch v.SubsampleRatio {
					case image.YCbCrSubsampleRatio444:
					case image.YCbCrSubsampleRatio422:
						planes[1].dst.Rect.Max.X /= 2
						planes[2].dst.Rect.Max.X /= 2
					case image.YCbCrSubsampleRatio420:
						planes[1].dst.Rect.Max.X /= 2
						planes[1].dst.Rect.Max.Y /= 2
						planes[2].dst.Rect.Max.X /= 2
						planes[2].dst.Rect.Max.Y /= 2
					}
				}
				// Scale each plane
				*planes[0].src = image.Gray{Pix: v.Y, Stride: v.YStride, Rect: v.Rect}
				*planes[1].src = image.Gray{Pix: v.Cr, Stride: v.CStride, Rect: v.Rect}
				*planes[2].src = image.Gray{Pix: v.Cb, Stride: v.CStride, Rect: v.Rect}
				switch v.SubsampleRatio {
				case image.YCbCrSubsampleRatio444:
				case image.YCbCrSubsampleRatio422:
					planes[1].src.Rect.Max.X /= 2
					planes[2].src.Rect.Max.X /= 2
				case image.YCbCrSubsampleRatio420:
					planes[1].src.Rect.Max.X /= 2
					planes[1].src.Rect.Max.Y /= 2
					planes[2].src.Rect.Max.X /= 2
					planes[2].src.Rect.Max.Y /= 2
				}
				// TODO: add option to enable multi threading
				var wg sync.WaitGroup
				wg.Add(3)
				for i, p := range planes {
					i, p := i, p
					go func() {
						scalers[i].Scale(p.dst, p.dst.Rect, p.src, p.src.Rect, draw.Over, nil)
						wg.Done()
					}()
				}
				wg.Wait()

			default:
				return nil, errUnsupportedImageType
			}

			return imgScaled, nil
		})
	}
}
