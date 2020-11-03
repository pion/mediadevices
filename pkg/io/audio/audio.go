package audio

import (
	"github.com/pion/mediadevices/pkg/wave"
)

type Reader interface {
	// Read reads data from the source. The caller is responsible to release the memory that's associated
	// with data by calling the given release function. When err is not nil, the caller MUST NOT call release
	// as data is going to be nil (no memory was given). Otherwise, the caller SHOULD call release after
	// using the data. The caller is NOT REQUIRED to call release, as this is only a part of memory management
	// optimization. If release is not called, the source is forced to allocate a new memory, which also means
	// there will be new allocations during streaming, and old unused memory will become garbage. As a consequence,
	// these garbage will put a lot of pressure to the garbage collector and makes it to run more often and finish
	// slower as the heap memory usage increases and more garbage to collect.
	Read() (chunk wave.Audio, release func(), err error)
}

type ReaderFunc func() (chunk wave.Audio, release func(), err error)

func (rf ReaderFunc) Read() (chunk wave.Audio, release func(), err error) {
	chunk, release, err = rf()
	return
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
