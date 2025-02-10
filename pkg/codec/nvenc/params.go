package nvenc

import (
	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

type Params struct {
	codec.BaseParams
}

func NewParams() (Params, error) {
	return Params{
		BaseParams: codec.BaseParams{
			KeyFrameInterval: 60,
		},
	}, nil
}

func (p *Params) RTPCodec() *codec.RTPCodec {
	return codec.NewRTPH264Codec(90000)
}

func (p *Params) BuildVideoEncoder(r video.Reader, property prop.Media) (codec.ReadCloser, error) {
	return newEncoder(r, property, *p)
}
