//go:build !cgo
// +build !cgo

package video

import (
	"golang.org/x/image/draw"
	"image"
)

// ScalerFastBoxSampling mock scaler for nocgo implementation to pass tests for CGO_ENABLED=0
var (
	ScalerFastBoxSampling = Scaler(&FastBoxSampling{})
)

// FastBoxSampling mock implementation for nocgo implementation
// TODO implement nocgo FastBoxSampling scaling algorithm
type FastBoxSampling struct {
}

func (f *FastBoxSampling) Scale(_ draw.Image, _ image.Rectangle, _ image.Image, _ image.Rectangle, _ draw.Op, _ *draw.Options) {
}
