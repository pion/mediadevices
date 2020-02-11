package audio

type Reader interface {
	Read(samples [][2]float32) (n int, err error)
}

type ReaderFunc func(samples [][2]float32) (n int, err error)

func (rf ReaderFunc) Read(samples [][2]float32) (n int, err error) {
	return rf(samples)
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
