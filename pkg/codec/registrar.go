package codec

import (
	"fmt"
	"io"

	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
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

func BuildVideoEncoder(name string, r video.Reader, p prop.Video) (io.ReadCloser, error) {
	b, ok := videoEncoders[name]
	if !ok {
		return nil, fmt.Errorf("codec: can't find %s video encoder", name)
	}

	return b(r, p)
}

func BuildAudioEncoder(name string, r audio.Reader, p prop.Audio) (io.ReadCloser, error) {
	b, ok := audioEncoders[name]
	if !ok {
		return nil, fmt.Errorf("codec: can't find %s audio encoder", name)
	}

	return b(r, p)
}
