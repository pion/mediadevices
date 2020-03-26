package x264

import (
	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/webrtc/v2"
)

// Params stores libx264 specific encoding parameters.
type Params struct {
	codec.BaseParams

	// Faster preset has lower CPU usage but lower quality
	Preset Preset
}

// Preset represents a set of default configurations from libx264
type Preset int

const (
	PresetUltrafast Preset = iota
	PresetSuperfast
	PresetVeryfast
	PresetFaster
	PresetFast
	PresetMedium
	PresetSlow
	PresetSlower
	PresetVeryslow
	PresetPlacebo
)

// NewParams returns default x264 codec specific parameters.
func NewParams() (Params, error) {
	return Params{
		BaseParams: codec.BaseParams{
			KeyFrameInterval: 60,
		},
	}, nil
}

// Name represents the codec name
func (p *Params) Name() string {
	return webrtc.H264
}

// BuildVideoEncoder builds x264 encoder with given params
func (p *Params) BuildVideoEncoder(r video.Reader, property prop.Media) (codec.ReadCloser, error) {
	return newEncoder(r, property, *p)
}
