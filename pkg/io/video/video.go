package video

import "image"

type Reader interface {
	Read() (img image.Image, err error)
}

type ReaderFunc func() (img image.Image, err error)

func (rf ReaderFunc) Read() (img image.Image, err error) {
	return rf()
}
