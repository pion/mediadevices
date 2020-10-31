package frame

import "image"

type buffer struct {
	image image.Image
	raw   []uint8
}
