package frame

import (
	"bytes"
	"image"
	"image/jpeg"
)

func decodeMJPEG(frame []byte, width, height int) (image.Image, error) {
	return jpeg.Decode(bytes.NewReader(frame))
}
