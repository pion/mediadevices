package io

// Reader is a generic data reader. In the future, interface{} should be replaced by a generic type
// to provide strong type.
type Reader interface {
	// Read reads data from the source. The caller is responsible to release the memory that's associated
	// with data by calling the given release function. When err is not nil, the caller MUST NOT call release
	// as data is going to be nil (no memory was given). Otherwise, the caller SHOULD call release after
	// using the data. The caller is NOT REQUIRED to call release, as this is only a part of memory management
	// optimization. If release is not called, the source is forced to allocate a new memory, which also means
	// there will be new allocations during streaming, and old unused memory will become garbage. As a consequence,
	// these garbage will put a lot of pressure to the garbage collector and makes it to run more often and finish
	// slower as the heap memory usage increases and more garbage to collect.
	Read() (data interface{}, release func(), err error)
}

// ReaderFunc is a proxy type for Reader
type ReaderFunc func() (data interface{}, release func(), err error)

func (f ReaderFunc) Read() (data interface{}, release func(), err error) {
	data, release, err = f()
	return
}
