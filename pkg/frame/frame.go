package frame

import "image"

type Decoder interface {
	Decode(frame []byte, width, height int) (image.Image, error)
}

// DecoderFunc is a proxy type for Decoder
type decoderFunc func(frame []byte, width, height int) (image.Image, error)

func (f decoderFunc) Decode(frame []byte, width, height int) (image.Image, error) {
	return f(frame, width, height)
}
