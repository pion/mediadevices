package frame

type Format uint

const (
	// YUV Formats

	// FormatYUVI420 https://www.fourcc.org/pixel-format/yuv-i420/
	FormatYUVI420 Format = iota
	// FormatYUVNV21 https://www.fourcc.org/pixel-format/yuv-nv21/
	FormatYUVNV21
	// FormatYUVYUY2 https://www.fourcc.org/pixel-format/yuv-yuy2/
	FormatYUVYUY2
)

// YUV aliases

// FormatYUVYUYV is an alias of FormatYUVYUY2
const FormatYUVYUYV = FormatYUVYUY2
