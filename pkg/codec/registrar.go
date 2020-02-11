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

func BuildVideoEncoder(r video.Reader, p prop.Media) (io.ReadCloser, error) {
	b, ok := videoEncoders[p.CodecName]
	if !ok {
		return nil, fmt.Errorf("codec: can't find %s video encoder", p.CodecName)
	}

	return b(r, p)
}

func BuildAudioEncoder(r audio.Reader, p prop.Media) (io.ReadCloser, error) {
	b, ok := audioEncoders[p.CodecName]
	if !ok {
		return nil, fmt.Errorf("codec: can't find %s audio encoder", p.CodecName)
	}

	return b(r, p)
}
