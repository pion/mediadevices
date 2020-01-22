package codec

import (
	"github.com/pion/mediadevices/pkg/io/audio"
	"image"
	"io"
)

type VideoEncoder interface {
	Encode(img image.Image) ([]byte, error)
	Close() error
}

type VideoDecoder interface {
	Decode([]byte) (image.Image, error)
	Close() error
}

type VideoSetting struct {
	Width, Height             int
	TargetBitRate, MaxBitRate int
	FrameRate                 float32
}

type VideoEncoderBuilder func(s VideoSetting) (VideoEncoder, error)
type VideoDecoderBuilder func(s VideoSetting) (VideoDecoder, error)

type AudioSetting struct {
	InSampleRate, OutSampleRate int
	// Latency in ms
	Latency float64
}

type AudioEncoderBuilder func(r audio.Reader, s AudioSetting) (io.ReadCloser, error)
