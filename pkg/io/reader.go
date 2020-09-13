package io

// Reader is a generic data reader. In the future, interface{} should be replaced by a generic type
// to provide strong type.
type Reader interface {
	Read() (interface{}, error)
}

// ReaderFunc is a proxy type for Reader
type ReaderFunc func() (interface{}, error)

func (f ReaderFunc) Read() (interface{}, error) {
	return f()
}
