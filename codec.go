package mediadevices

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/webrtc/v2"
)

// CodecSelector is a container of video and audio encoder builders, which later will be used
// for codec matching.
type CodecSelector struct {
	videoEncoders []codec.VideoEncoderBuilder
	audioEncoders []codec.AudioEncoderBuilder
}

// CodecSelectorOption is a type for specifying CodecSelector options
type CodecSelectorOption func(*CodecSelector)

// WithVideoEncoders replace current video codecs with listed encoders
func WithVideoEncoders(encoders ...codec.VideoEncoderBuilder) CodecSelectorOption {
	return func(t *CodecSelector) {
		t.videoEncoders = encoders
	}
}

// WithVideoEncoders replace current audio codecs with listed encoders
func WithAudioEncoders(encoders ...codec.AudioEncoderBuilder) CodecSelectorOption {
	return func(t *CodecSelector) {
		t.audioEncoders = encoders
	}
}

// NewCodecSelector constructs CodecSelector with given variadic options
func NewCodecSelector(opts ...CodecSelectorOption) *CodecSelector {
	var track CodecSelector

	for _, opt := range opts {
		opt(&track)
	}

	return &track
}

// Populate lets the webrtc engine be aware of supported codecs that are contained in CodecSelector
func (selector *CodecSelector) Populate(setting *webrtc.MediaEngine) {
	for _, encoder := range selector.videoEncoders {
		setting.RegisterCodec(encoder.RTPCodec().RTPCodec)
	}

	for _, encoder := range selector.audioEncoders {
		setting.RegisterCodec(encoder.RTPCodec().RTPCodec)
	}
}

func (selector *CodecSelector) selectVideoCodecByNames(reader video.Reader, inputProp prop.Media, codecNames ...string) (codec.ReadCloser, *codec.RTPCodec, error) {
	var selectedEncoder codec.VideoEncoderBuilder
	var encodedReader codec.ReadCloser
	var errReasons []string
	var err error

outer:
	for _, wantCodec := range codecNames {
		for _, encoder := range selector.videoEncoders {
			if encoder.RTPCodec().Name == wantCodec {
				encodedReader, err = encoder.BuildVideoEncoder(reader, inputProp)
				if err == nil {
					selectedEncoder = encoder
					break outer
				}
			}

			errReasons = append(errReasons, fmt.Sprintf("%s: %s", encoder.RTPCodec().Name, err))
		}
	}

	if selectedEncoder == nil {
		return nil, nil, errors.New(strings.Join(errReasons, "\n\n"))
	}

	return encodedReader, selectedEncoder.RTPCodec(), nil
}

func (selector *CodecSelector) selectVideoCodec(reader video.Reader, inputProp prop.Media, codecs ...*webrtc.RTPCodec) (codec.ReadCloser, *codec.RTPCodec, error) {
	var codecNames []string

	for _, codec := range codecs {
		codecNames = append(codecNames, codec.Name)
	}

	return selector.selectVideoCodecByNames(reader, inputProp, codecNames...)
}

func (selector *CodecSelector) selectAudioCodecByNames(reader audio.Reader, inputProp prop.Media, codecNames ...string) (codec.ReadCloser, *codec.RTPCodec, error) {
	var selectedEncoder codec.AudioEncoderBuilder
	var encodedReader codec.ReadCloser
	var errReasons []string
	var err error

outer:
	for _, wantCodec := range codecNames {
		for _, encoder := range selector.audioEncoders {
			if encoder.RTPCodec().Name == wantCodec {
				encodedReader, err = encoder.BuildAudioEncoder(reader, inputProp)
				if err == nil {
					selectedEncoder = encoder
					break outer
				}
			}

			errReasons = append(errReasons, fmt.Sprintf("%s: %s", encoder.RTPCodec().Name, err))
		}
	}

	if selectedEncoder == nil {
		return nil, nil, errors.New(strings.Join(errReasons, "\n\n"))
	}

	return encodedReader, selectedEncoder.RTPCodec(), nil
}

func (selector *CodecSelector) selectAudioCodec(reader audio.Reader, inputProp prop.Media, codecs ...*webrtc.RTPCodec) (codec.ReadCloser, *codec.RTPCodec, error) {
	var codecNames []string

	for _, codec := range codecs {
		codecNames = append(codecNames, codec.Name)
	}

	return selector.selectAudioCodecByNames(reader, inputProp, codecNames...)
}
