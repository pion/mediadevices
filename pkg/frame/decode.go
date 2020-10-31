package frame

import (
	"fmt"
)

func NewDecoder(f Format) (Decoder, error) {
	var buildDecoder func() decoderFunc

	switch f {
	case FormatI420:
		buildDecoder = decodeI420
	case FormatNV21:
		buildDecoder = decodeNV21
	case FormatYUY2:
		buildDecoder = decodeYUY2
	case FormatUYVY:
		buildDecoder = decodeUYVY
	case FormatMJPEG:
		buildDecoder = decodeMJPEG
	default:
		return nil, fmt.Errorf("%s is not supported", f)
	}

	return buildDecoder(), nil
}
