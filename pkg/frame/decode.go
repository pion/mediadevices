package frame

import (
	"fmt"
)

func NewDecoder(f Format) (Decoder, error) {
	var decoder decoderFunc

	switch f {
	case FormatI420:
		decoder = decodeI420
	case FormatNV21:
		decoder = decodeNV21
	case FormatYUY2:
		decoder = decodeYUY2
	case FormatUYVY:
		decoder = decodeUYVY
	case FormatMJPEG:
		decoder = decodeMJPEG
	default:
		return nil, fmt.Errorf("%s is not supported", f)
	}

	return decoder, nil
}
