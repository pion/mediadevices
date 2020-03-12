package codec

import (
	"errors"
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

// Register a new EncoderBuilder for the name.
func Register(name string, builder interface{}) {
	switch b := builder.(type) {
	case VideoEncoderBuilder:
		videoEncoders[name] = b
	case AudioEncoderBuilder:
		audioEncoders[name] = b
	}
}

// BuildVideoEncoder builds encoder for the codec p.CodecName.
func BuildVideoEncoder(r video.Reader, p prop.Media) (io.ReadCloser, error) {
	b, ok := videoEncoders[p.CodecName]
	if !ok {
		return nil, fmt.Errorf("codec: can't find %s video encoder", p.CodecName)
	}

	return b(r, p)
}

// BuildAudioEncoder builds encoder for the codec p.CodecName.
func BuildAudioEncoder(r audio.Reader, p prop.Media) (io.ReadCloser, error) {
	b, ok := audioEncoders[p.CodecName]
	if !ok {
		return nil, fmt.Errorf("codec: can't find %s audio encoder", p.CodecName)
	}

	return b(r, p)
}

// NamedVideoCodec stores codec and its name.
type NamedVideoCodec struct {
	Name  string
	Codec func(video.Reader, prop.Media) (io.ReadCloser, error)
}

// NamedAudioCodec stores codec and its name.
type NamedAudioCodec struct {
	Name  string
	Codec func(audio.Reader, prop.Media) (io.ReadCloser, error)
}

// VideoEncoderFallbacks packs multiple codecs as a new VideoEncoderBuilder.
// First codec is tried to initialize first, and the next codec will be tried if failed.
func VideoEncoderFallbacks(codecs ...NamedVideoCodec) VideoEncoderBuilder {
	return VideoEncoderBuilder(func(r video.Reader, p prop.Media) (io.ReadCloser, error) {
		codecParams := map[string]interface{}{}
		switch cp := p.CodecParams.(type) {
		case nil:
		case (map[string]interface{}):
			codecParams = cp
		default:
			return nil, errors.New("CodecParams of VideoEncoderFallbacks must be map[string]interface{}")
		}

		var rc io.ReadCloser
		var err error
		for _, codec := range codecs {
			if cp, ok := codecParams[codec.Name]; ok {
				p.CodecParams = cp
			} else {
				p.CodecParams = nil
			}
			rc, err = codec.Codec(r, p)
			if err == nil {
				return rc, nil
			}
		}
		return nil, err
	})
}

// AudioEncoderFallbacks packs multiple codecs as a new AudioEncoderBuilder.
// First codec is tried to initialize first, and the next codec will be tried if failed.
func AudioEncoderFallbacks(codecs ...NamedAudioCodec) AudioEncoderBuilder {
	return AudioEncoderBuilder(func(r audio.Reader, p prop.Media) (io.ReadCloser, error) {
		codecParams := map[string]interface{}{}
		switch cp := p.CodecParams.(type) {
		case nil:
		case (map[string]interface{}):
			codecParams = cp
		default:
			return nil, errors.New("CodecParams of AudioEncoderFallbacks must be map[string]interface{}")
		}

		var rc io.ReadCloser
		var err error
		for _, codec := range codecs {
			if cp, ok := codecParams[codec.Name]; ok {
				p.CodecParams = cp
			} else {
				p.CodecParams = nil
			}
			rc, err = codec.Codec(r, p)
			if err == nil {
				return rc, nil
			}
		}
		return nil, err
	})
}
