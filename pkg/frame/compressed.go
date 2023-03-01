package frame

import (
	"bytes"
	libjpeg "github.com/pixiv/go-libjpeg/jpeg"
	"image"
)

func decodeMJPEG(frame []byte, width, height int) (image.Image, func(), error) {
	img, err := libjpeg.Decode(bytes.NewReader(frame), &libjpeg.DecoderOptions{DCTMethod: libjpeg.DCTIFast})
	return img, func() {}, err
}
