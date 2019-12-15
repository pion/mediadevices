package frame

import (
	"fmt"
)

func NewDecoder(f Format) (Decoder, error) {
	var decoder DecoderFunc

	switch f {
	case FormatYUVI420:
		decoder = decodeI420
	case FormatYUVNV21:
		decoder = decodeNV21
	case FormatYUVYUY2:
		decoder = decodeYUY2
	default:
		return nil, fmt.Errorf("unsupported format")
	}

	return decoder, nil
}
