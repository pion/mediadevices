package frame

type Format string

const (
	// FormatI420 https://www.fourcc.org/pixel-format/yuv-i420/
	FormatI420 Format = "I420"
	// FormatI444 is a YUV format without sub-sampling
	FormatI444 Format = "I444"
	// FormatNV21 https://www.fourcc.org/pixel-format/yuv-nv21/
	FormatNV21 = "NV21"
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
