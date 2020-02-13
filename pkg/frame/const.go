package frame

type Format string

const (
	// YUV Formats

	// FormatI420 https://www.fourcc.org/pixel-format/yuv-i420/
	FormatI420 Format = "I420"
	// FormatI444 is a YUV format without sub-sampling
	FormatI444 Format = "I444"
	// FormatNV21 https://www.fourcc.org/pixel-format/yuv-nv21/
	FormatNV21 = "NV21"
	// FormatYUY2 https://www.fourcc.org/pixel-format/yuv-yuy2/
	FormatYUY2 = "YUY2"

	// Compressed Formats

	// FormatMJPEG https://www.fourcc.org/mjpg/
	FormatMJPEG = "MJPEG"
)

// YUV aliases

// FormatYUYV is an alias of FormatYUY2
const FormatYUYV = FormatYUY2
