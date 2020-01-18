package audio

type Reader interface {
	Read(samples [][2]float32) (n int, err error)
}

type ReaderFunc func(samples [][2]float32) (n int, err error)

func (rf ReaderFunc) Read(samples [][2]float32) (n int, err error) {
	return rf(samples)
}
