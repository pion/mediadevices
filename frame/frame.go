package frame

import "image"

type Decoder interface {
	Decode(frame []byte, width, height int) (*image.YCbCr, error)
}

// DecoderFunc is a proxy type for Decoder
type DecoderFunc func(frame []byte, width, height int) (*image.YCbCr, error)

func (f DecoderFunc) Decode(frame []byte, width, height int) (*image.YCbCr, error) {
	return f(frame, width, height)
}
