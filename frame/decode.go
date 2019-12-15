package frame

import (
	"fmt"
)

func NewDecoder(f Format) (Decoder, error) {
	var decoder DecoderFunc

	switch f {
	case FormatI420:
		decoder = decodeI420
	case FormatNV21:
		decoder = decodeNV21
	case FormatYUY2:
		decoder = decodeYUY2
	default:
		return nil, fmt.Errorf("%s is not supported", f)
	}

	return decoder, nil
}
