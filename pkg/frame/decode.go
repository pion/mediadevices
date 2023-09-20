package frame

import (
	"fmt"
)

type Format string

const (
	// FormatI420 https://wiki.videolan.org/YUV#I420
	FormatI420 Format = "I420"
	// FormatI444 is a YUV format without sub-sampling
	FormatI444 Format = "I444"
	// FormatNV21 https://www.kernel.org/doc/html/v5.9/userspace-api/media/v4l/pixfmt-nv12.html
	FormatNV21 = "NV21"
	// FormatNV12 https://www.kernel.org/doc/html/v5.9/userspace-api/media/v4l/pixfmt-nv12.html
	FormatNV12 = "NV12"
	// FormatYUY2 https://www.kernel.org/doc/html/v5.9/userspace-api/media/v4l/pixfmt-yuyv.html
	// YUY2 is what Windows calls YUYV
	FormatYUY2 = "YUY2"
	// FormatYUYV https://www.kernel.org/doc/html/v5.9/userspace-api/media/v4l/pixfmt-yuyv.html
	FormatYUYV = "YUYV"
	// FormatUYVY https://www.kernel.org/doc/html/v5.9/userspace-api/media/v4l/pixfmt-uyvy.html
	FormatUYVY = "UYVY"

	// FormatRGBA https://www.kernel.org/doc/html/v5.9/userspace-api/media/v4l/pixfmt-rgb.html
	FormatRGBA Format = "RGBA"

	// FormatMJPEG https://wiki.videolan.org/MJPEG
	FormatMJPEG = "MJPEG"

	// FormatZ16 https://www.kernel.org/doc/html/v5.9/userspace-api/media/v4l/pixfmt-z16.html
	FormatZ16 = "Z16"
)

var decoderMap = map[Format]decoderFunc{
	FormatI420:  decodeI420,
	FormatNV21:  decodeNV21,
	FormatNV12:  decodeNV12,
	FormatYUY2:  decodeYUY2,
	FormatYUYV:  decodeYUY2,
	FormatUYVY:  decodeUYVY,
	FormatMJPEG: decodeMJPEG,
	FormatZ16:   decodeZ16,
}

func NewDecoder(f Format) (Decoder, error) {
	decoder, ok := decoderMap[f]

	if !ok {
		return nil, fmt.Errorf("%s is not supported", f)
	}

	return decoder, nil
}
