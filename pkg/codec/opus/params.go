package opus

import (
	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/webrtc/v2"
)

// Params stores opus specific encoding parameters.
type Params struct {
	codec.BaseParams
}

// NewParams returns default opus codec specific parameters.
func NewParams() (Params, error) {
	return Params{}, nil
}

// Name represents the codec name
func (p *Params) Name() string {
	return webrtc.Opus
}

// BuildAudioEncoder builds opus encoder with given params
func (p *Params) BuildAudioEncoder(r audio.Reader, property prop.Media) (codec.ReadCloser, error) {
	return newEncoder(r, property, *p)
}
