package frame

// Return a function to get the number of bytes a frame will occupy in the given format
var FrameSizeMap = map[Format]frameSizeFunc{
	FormatI420:  frameSizeI420,
	FormatNV21:  frameSizeNV21,
	FormatNV12:  frameSizeNV21, // NV12 and NV21 have the same frame size
	FormatYUY2:  frameSizeYUY2,
	FormatUYVY:  frameSizeYUY2, // UYVY and YUY2 have the same frame size
	FormatMJPEG: frameSizeMJPEG,
	FormatZ16:   frameSizeZ16,
}

type frameSizeFunc func(width, height int) uint

func frameSizeYUY2(width, height int) uint {
	yi := width * height
	// ci := yi / 2
	// fi := yi + 2*ci
	fi := 2 * yi
	return uint(fi)
}

func frameSizeI420(width, height int) uint {
	yi := width * height
	cbi := yi + width*height/4
	cri := cbi + width*height/4
	return uint(cri)
}

func frameSizeNV21(width, height int) uint {
	yi := width * height
	ci := yi + width*height/2
	return uint(ci)
}

func frameSizeZ16(width, height int) uint {
	expectedSize := 2 * (width * height)
	return uint(expectedSize)
}

func frameSizeMJPEG(width, height int) uint {
	// MJPEG is a compressed format, so we don't know the size
	panic("Not possible to get frame size with MJPEG format. Since it is a compressed format, so we don't know the size")
}
