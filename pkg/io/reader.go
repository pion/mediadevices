package io

// Reader is a generic reader. When generic is ready, interface{} will be replaced
// with a generic type and will provide type safety.
type Reader interface {
	Read() (interface{}, error)
}

// ReaderFunc is a proxy type to make easier for users to implement Reader
type ReaderFunc func() (interface{}, error)

func (f ReaderFunc) Read() (interface{}, error) {
	return f()
}
