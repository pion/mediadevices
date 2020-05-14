package openh264

import (
	"fmt"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec"
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
func (p *Params) Codec() *webrtc.RTPCodec {
	return webrtc.NewRTPH264Codec(webrtc.DefaultPayloadTypeH264, 90000)
}

// BuildVideoEncoder builds openh264 encoder with given params
func (p *Params) BuildEncoder(track mediadevices.Track) (codec.RTPReadCloser, error) {
	videoTrack, ok := track.(*mediadevices.VideoTrack)
	if !ok {
		return nil, fmt.Errorf("track is not a video track")
	}

	encoder, err := newEncoder(videoTrack, *p)
	if err != nil {
		return nil, err
	}

	return codec.NewRTPReadCloser(
		p.Codec(),
		encoder,
		codec.NewVideoSampler(p.Codec().ClockRate),
	)
}
