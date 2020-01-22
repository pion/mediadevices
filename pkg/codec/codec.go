package codec

import (
	"io"

	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
)

type VideoSetting struct {
	Width, Height             int
	TargetBitRate, MaxBitRate int
	FrameRate                 float32
}

type VideoEncoderBuilder func(r video.Reader, s VideoSetting) (io.ReadCloser, error)

type AudioSetting struct {
	InSampleRate, OutSampleRate int
	// Latency in ms
	Latency float64
}

type AudioEncoderBuilder func(r audio.Reader, s AudioSetting) (io.ReadCloser, error)
