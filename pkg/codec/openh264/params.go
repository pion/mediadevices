package openh264

import (
	"io"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/webrtc/v2"
)

// Params stores libopenh264 specific encoding parameters.
type Params struct {
	codec.BaseParams
}

// NewParams returns default openh264 codec specific parameters.
func NewParams() (Params, error) {
	return Params{
		BaseParams: codec.BaseParams{
			BitRate: 100000,
		},
	}, nil
}

// Name represents the codec name
func (p *Params) Name() string {
	return webrtc.H264
}

// BuildVideoEncoder builds openh264 encoder with given params
func (p *Params) BuildVideoEncoder(r video.Reader, property prop.Media) (io.ReadCloser, error) {
	return newEncoder(r, property, *p)
}
