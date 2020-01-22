package codec

import (
	"fmt"
	"github.com/pion/mediadevices/pkg/io/audio"
	"io"
)

var (
	videoEncoders = make(map[string]VideoEncoderBuilder)
	audioEncoders = make(map[string]AudioEncoderBuilder)
)

func Register(name string, builder interface{}) {
	switch b := builder.(type) {
	case VideoEncoderBuilder:
		videoEncoders[name] = b
	case AudioEncoderBuilder:
		audioEncoders[name] = b
	}
}

func BuildVideoEncoder(name string, s VideoSetting) (VideoEncoder, error) {
	b, ok := videoEncoders[name]
	if !ok {
		return nil, fmt.Errorf("codec: can't find %s video encoder", name)
	}

	return b(s)
}

func BuildAudioEncoder(name string, r audio.Reader, s AudioSetting) (io.ReadCloser, error) {
	b, ok := audioEncoders[name]
	if !ok {
		return nil, fmt.Errorf("codec: can't find %s audio encoder", name)
	}

	return b(r, s)
}
