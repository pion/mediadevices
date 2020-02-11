package video

import (
	"image"
)

type Reader interface {
	Read() (img image.Image, err error)
}

type ReaderFunc func() (img image.Image, err error)

func (rf ReaderFunc) Read() (img image.Image, err error) {
	return rf()
}

// TransformFunc produces a new Reader that will produces a transformed video
type TransformFunc func(r Reader) Reader

// Merge merges transforms and produces a new TransformFunc that will execute
// transforms in order
func Merge(transforms ...TransformFunc) TransformFunc {
	return func(r Reader) Reader {
		for _, transform := range transforms {
			if transform == nil {
				continue
			}

			r = transform(r)
		}

		return r
	}
}
