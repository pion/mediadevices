// +build cgo

package video

import (
	"golang.org/x/image/draw"
	"image"
	"sync"
)

// #include <stdint.h>
// void fastNearestNeighbor(
//     uint8_t* dst, const uint8_t* src,
//     const int ch,
//     const int dw, const int dh, const int dstride,
//     const int sw, const int sh, const int sstride);
// void fastBoxSampling(
//     uint8_t* dst, const uint8_t* src,
//     const int ch,
//     const int dw, const int dh, const int dstride,
//     const int sw, const int sh, const int sstride,
//     uint32_t* tmp);
import "C"

// List of scaling algorithms
var (
	// ScalerFastNearestNeighbor is a CGO version of NearestNeighbor scaler.
	// This is roughly 4-times faster than draw.NearestNeighbor.
	ScalerFastNearestNeighbor = Scaler(&FastNearestNeighbor{})
	// ScalerFastBoxSampling is a CGO implementation of BoxSampling scaler.
	// This is heavyer than NearestNeighbor but keeps detail on down scaling.
	ScalerFastBoxSampling = Scaler(&FastBoxSampling{})
)

var poolBoxSampleBuffer = sync.Pool{
	New: func() interface{} {
		return &([]uint32{})
	},
}

var (
	scalerTestAlgosCGO = map[string]Scaler{
		"FastNearestNeighbor": ScalerFastNearestNeighbor,
	}
	scalerBenchAlgosCGO = map[string]Scaler{
		"FastNearestNeighbor": ScalerFastNearestNeighbor,
		"FastBoxSampling":     ScalerFastBoxSampling,
	}
)

func init() {
	// Append test conditions
	for k, v := range scalerTestAlgosCGO {
		scalerTestAlgos[k] = v
	}
	for k, v := range scalerBenchAlgosCGO {
		scalerBenchAlgos[k] = v
	}
}

// FastNearestNeighbor is a CGO version of NearestNeighbor scaler.
type FastNearestNeighbor struct {
}

// Scale implements the draw.Scaler interface.
func (f *FastNearestNeighbor) Scale(dst draw.Image, dr image.Rectangle, src image.Image, sr image.Rectangle, op draw.Op, opts *draw.Options) {
	switch s := src.(type) {
	case (*image.RGBA):
		d := dst.(*image.RGBA)
		l := d.Stride * d.Rect.Dy()
		if len(d.Pix) < l {
			if cap(d.Pix) < l {
				d.Pix = make([]uint8, l)
			}
			d.Pix = d.Pix[:l]
		}
		C.fastNearestNeighbor(
			(*C.uchar)(&d.Pix[dr.Min.X+d.Stride*dr.Min.Y]),
			(*C.uchar)(&s.Pix[sr.Min.X+s.Stride*sr.Min.Y]),
			4,
			C.int(dr.Dx()), C.int(dr.Dy()), C.int(d.Stride),
			C.int(sr.Dx()), C.int(sr.Dy()), C.int(s.Stride),
		)
	case (*image.Gray):
		d := dst.(*image.Gray)
		l := d.Stride * d.Rect.Dy()
		if len(d.Pix) < l {
			if cap(d.Pix) < l {
				d.Pix = make([]uint8, l)
			}
			d.Pix = d.Pix[:l]
		}
		C.fastNearestNeighbor(
			(*C.uchar)(&d.Pix[dr.Min.X+d.Stride*dr.Min.Y]),
			(*C.uchar)(&s.Pix[sr.Min.X+s.Stride*sr.Min.Y]),
			1,
			C.int(dr.Dx()), C.int(dr.Dy()), C.int(d.Stride),
			C.int(sr.Dx()), C.int(sr.Dy()), C.int(s.Stride),
		)

	case (*rgbLikeYCbCr):
		d := dst.(*rgbLikeYCbCr)
		f.Scale(d.y, dr, s.y, sr, op, opts)
		dr2 := image.Rect(0, 0, d.cb.Stride, len(d.cb.Pix)/d.cb.Stride)
		sr2 := image.Rect(0, 0, s.cb.Stride, len(s.cb.Pix)/s.cb.Stride)
		f.Scale(d.cb, dr2, s.cb, sr2, op, opts)
		f.Scale(d.cr, dr2, s.cr, sr2, op, opts)

	default:
		panic("unimplemented")
	}
}

// FastBoxSampling is a CGO version of Box sampling scaler.
type FastBoxSampling struct {
}

// Scale implements the draw.Scaler interface.
func (f *FastBoxSampling) Scale(dst draw.Image, dr image.Rectangle, src image.Image, sr image.Rectangle, op draw.Op, opts *draw.Options) {
	if sr.Dx() < dr.Dx() && sr.Dy() < dr.Dy() {
		// Box upsampling is equivalent of NearestNeighbor
		(&FastNearestNeighbor{}).Scale(dst, dr, src, sr, op, opts)
		return
	}
	tmp := poolBoxSampleBuffer.Get().(*[]uint32)
	defer poolBoxSampleBuffer.Put(tmp)

	switch s := src.(type) {
	case (*image.RGBA):
		d := dst.(*image.RGBA)
		l := d.Stride * d.Rect.Dy()
		if len(d.Pix) < l {
			if cap(d.Pix) < l {
				d.Pix = make([]uint8, l)
			}
			d.Pix = d.Pix[:l]
		}
		if len(*tmp) < l {
			*tmp = make([]uint32, l)
		}
		C.fastBoxSampling(
			(*C.uchar)(&d.Pix[dr.Min.X+d.Stride*dr.Min.Y]),
			(*C.uchar)(&s.Pix[sr.Min.X+s.Stride*sr.Min.Y]),
			4,
			C.int(dr.Dx()), C.int(dr.Dy()), C.int(d.Stride),
			C.int(sr.Dx()), C.int(sr.Dy()), C.int(s.Stride),
			(*C.uint32_t)(&(*tmp)[0]),
		)
	case (*image.Gray):
		d := dst.(*image.Gray)
		l := d.Stride * d.Rect.Dy()
		if len(d.Pix) < l {
			if cap(d.Pix) < l {
				d.Pix = make([]uint8, l)
			}
			d.Pix = d.Pix[:l]
		}
		if len(*tmp) < l {
			*tmp = make([]uint32, l)
		}
		C.fastBoxSampling(
			(*C.uchar)(&d.Pix[dr.Min.X+d.Stride*dr.Min.Y]),
			(*C.uchar)(&s.Pix[sr.Min.X+s.Stride*sr.Min.Y]),
			1,
			C.int(dr.Dx()), C.int(dr.Dy()), C.int(d.Stride),
			C.int(sr.Dx()), C.int(sr.Dy()), C.int(s.Stride),
			(*C.uint32_t)(&(*tmp)[0]),
		)

	case (*rgbLikeYCbCr):
		d := dst.(*rgbLikeYCbCr)
		f.Scale(d.y, dr, s.y, sr, op, opts)
		dr2 := image.Rect(0, 0, d.cb.Stride, len(d.cb.Pix)/d.cb.Stride)
		sr2 := image.Rect(0, 0, s.cb.Stride, len(s.cb.Pix)/s.cb.Stride)
		f.Scale(d.cb, dr2, s.cb, sr2, op, opts)
		f.Scale(d.cr, dr2, s.cr, sr2, op, opts)

	default:
		panic("unimplemented")
	}
}
