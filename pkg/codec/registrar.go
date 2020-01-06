package codec

import (
	"fmt"
)

var (
	videoEncoders = make(map[string]VideoEncoderBuilder)
	videoDecoders = make(map[string]VideoDecoderBuilder)
)

func Register(name string, builder interface{}) {
	switch b := builder.(type) {
	case VideoEncoderBuilder:
		videoEncoders[name] = b
	case VideoDecoderBuilder:
		videoDecoders[name] = b
	}
}

func BuildVideoEncoder(name string, s VideoSetting) (VideoEncoder, error) {
	b, ok := videoEncoders[name]
	if !ok {
		return nil, fmt.Errorf("codec: can't find %s video encoder", name)
	}

	return b(s)
}

func BuildVideoDecoder(name string, s VideoSetting) (VideoDecoder, error) {
	b, ok := videoDecoders[name]
	if !ok {
		return nil, fmt.Errorf("codec: can't find %s video decoder", name)
	}

	return b(s)
}
