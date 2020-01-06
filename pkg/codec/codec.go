package codec

import "image"

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
