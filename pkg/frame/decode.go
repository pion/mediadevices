package frame

import (
	"fmt"
)

type Format string

const (
	// FormatI420 https://www.fourcc.org/pixel-format/yuv-i420/
	FormatI420 Format = "I420"
	// FormatI444 is a YUV format without sub-sampling
	FormatI444 Format = "I444"
	// FormatNV21 https://www.fourcc.org/pixel-format/yuv-nv21/
	FormatNV21 = "NV21"
	// FormatNV12 https://www.fourcc.org/pixel-format/yuv-nv12/
	FormatNV12 = "NV12"
	// FormatYUY2 https://www.fourcc.org/pixel-format/yuv-yuy2/
	FormatYUY2 = "YUY2"
	// FormatUYVY https://www.fourcc.org/pixel-format/yuv-uyvy/
	FormatUYVY = "UYVY"

	// FormatRGBA https://www.fourcc.org/pixel-format/rgb-rgba/
	FormatRGBA Format = "RGBA"

	// FormatMJPEG https://www.fourcc.org/mjpg/
	FormatMJPEG = "MJPEG"
)

const FormatYUYV = FormatYUY2

var decoderMap = map[Format]decoderFunc{
	FormatI420:  decodeI420,
	FormatNV21:  decodeNV21,
	FormatNV12:  decodeNV12,
	FormatYUY2:  decodeYUY2,
	FormatUYVY:  decodeUYVY,
	FormatMJPEG: decodeMJPEG,
}

func NewDecoder(f Format) (Decoder, error) {
	decoder, ok := decoderMap[f]

	if !ok {
		return nil, fmt.Errorf("%s is not supported", f)
	}

	return decoder, nil
}
