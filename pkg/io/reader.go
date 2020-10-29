package io

// Reader is a generic data reader. In the future, interface{} should be replaced by a generic type
// to provide strong type.
type Reader interface {
	Read() (data interface{}, release func(), err error)
}

// ReaderFunc is a proxy type for Reader
type ReaderFunc func() (data interface{}, release func(), err error)

func (f ReaderFunc) Read() (data interface{}, release func(), err error) {
	data, release, err = f()
	return
}
