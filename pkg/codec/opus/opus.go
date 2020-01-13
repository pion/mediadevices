package opus

import (
	"fmt"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/webrtc/v2"
	"gopkg.in/hraban/opus.v2"
)

type encoder struct {
	engine *opus.Encoder
	buff   []byte
}

var _ codec.AudioEncoder = &encoder{}
var _ codec.AudioEncoderBuilder = codec.AudioEncoderBuilder(NewEncoder)

func init() {
	codec.Register(webrtc.Opus, codec.AudioEncoderBuilder(NewEncoder))
}

func NewEncoder(s codec.AudioSetting) (codec.AudioEncoder, error) {
	channels := 1 // mono
	engine, err := opus.NewEncoder(48000, channels, opus.AppVoIP)
	if err != nil {
		return nil, err
	}

	buffSize := 1024
	buff := make([]byte, buffSize)
	return &encoder{engine, buff}, nil
}

func (e *encoder) Encode(b []int16) ([]byte, error) {
	frameSize := len(b) // must be interleaved if stereo
	frameSizeMs := float32(frameSize) * 1000 / 48000
	switch frameSizeMs {
	case 2.5, 5, 10, 20, 40, 60:
		// Good.
	default:
		return nil, fmt.Errorf("Illegal frame size: %d bytes (%f ms)", frameSize, frameSizeMs)
	}

	n, err := e.engine.Encode(b, e.buff)
	if err != nil {
		return nil, err
	}

	return e.buff[:n], nil
}

func (e *encoder) Close() error {
	return nil
}
