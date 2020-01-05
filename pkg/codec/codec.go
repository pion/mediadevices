package codec

import "image"

type Encoder interface {
	Encode(img image.Image) ([]byte, error)
	Close() error
}

type Decoder interface {
	Decode([]byte) (image.Image, error)
	Close() error
}
