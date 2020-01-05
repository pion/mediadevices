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
