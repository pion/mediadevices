package svtav1

import (
	"time"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

// Params stores libx264 specific encoding parameters.
type Params struct {
	codec.BaseParams

	// Preset configuration number of SVT-AV1
	// 1-3: extremely high efficiency but heavy
	// 4-6: a balance of efficiency and reasonable compute time
	// 7-13: real-time encoding
	Preset int

	StartingBufferLevel time.Duration
	OptimalBufferLevel  time.Duration
	MaximumBufferSize   time.Duration
}

// NewParams returns default x264 codec specific parameters.
func NewParams() (Params, error) {
	return Params{
		BaseParams: codec.BaseParams{
			KeyFrameInterval: 60,
		},
		Preset:              9,
		StartingBufferLevel: 400 * time.Millisecond,
		OptimalBufferLevel:  200 * time.Millisecond,
		MaximumBufferSize:   500 * time.Millisecond,
	}, nil
}

// RTPCodec represents the codec metadata
func (p *Params) RTPCodec() *codec.RTPCodec {
	return codec.NewRTPAV1Codec(90000)
}

// BuildVideoEncoder builds x264 encoder with given params
func (p *Params) BuildVideoEncoder(r video.Reader, property prop.Media) (codec.ReadCloser, error) {
	return newEncoder(r, property, *p)
}
