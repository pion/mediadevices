package audio

import (
	"github.com/pion/mediadevices/pkg/wave"
)

type Reader interface {
	Read() (wave.Audio, error)
}

type ReaderFunc func() (wave.Audio, error)

func (rf ReaderFunc) Read() (wave.Audio, error) {
	return rf()
}

// TransformFunc produces a new Reader that will produces a transformed audio
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
