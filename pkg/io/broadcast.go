package io

import (
	"fmt"
	"runtime"
	"sync/atomic"
)

const (
	maxDataCount = 0xfffffff
	maskReading  = 0x80000000
)

var errEmptySource = fmt.Errorf("Source can't be nil")

type broadcasterData struct {
	data  interface{}
	count uint32
	err   error
}

// Broadcaster is a generic pull-based broadcaster. Broadcaster is unique in a sense that
// readers can come and go at anytime, and readers don't need to close or notify broadcaster.
type Broadcaster struct {
	source atomic.Value
	// reading (1 bit) + reserved (3 bits) + data count (28 bits)
	state uint32
	// TODO: instead of storing only last data. A fixed length ring buffer
	//       is probably better so that we reduce the amount of data lost.
	last atomic.Value
}

// NewNewBroadcaster creates a new broadcaster.
func NewBroadcaster(source Reader) *Broadcaster {
	var broadcaster Broadcaster
	broadcaster.ReplaceSource(source)

	return &broadcaster
}

// NewReader creates a new reader. Each reader will retrieve the same data from the source.
// copyFn is used to copy the data from the source to individual readers. Broadcaster uses a single
// buffer, this means that slow readers might miss some data. It's also possible that some readers
// miss some data even when they're all reading at the same rate since this is solely up to the Go
// scheduler to decide with goroutines that should run.
func (broadcaster *Broadcaster) NewReader(copyFn func(interface{}) interface{}) Reader {
	var currentCount uint32

	return ReaderFunc(func() (data interface{}, err error) {
		reading := currentCount | maskReading

		// Reader has reached the latest data, should read from the source.
		// Only allow 1 reader to read from the source. When there are more than 1 readers,
		// the other readers will need to share the same data that the first reader gets from
		// the source.
		if atomic.CompareAndSwapUint32(&broadcaster.state, currentCount, reading) {
			nextCount := (currentCount + 1) % maxDataCount
			data, err = broadcaster.source.Load().(Reader).Read()
			broadcaster.last.Store(&broadcasterData{
				data:  data,
				err:   err,
				count: nextCount,
			})

			atomic.StoreUint32(&broadcaster.state, nextCount)

			currentCount = nextCount
			data = copyFn(data)

			return
		}

		// Since it's possible for the first reader to finish reading from the source before the current reader
		// reaches this point, we need to make sure that the global count is still the same and the first reader
		// is still reading.
		// TODO: since it's lockless, it spends a lot of resources in the scheduling.
		for atomic.LoadUint32(&broadcaster.state) == reading {
			// Yield current goroutine to let other goroutines to run instead
			runtime.Gosched()
		}

		last := broadcaster.last.Load().(*broadcasterData)
		data, err, currentCount = copyFn(last.data), last.err, last.count
		return
	})
}

// ReplaceSource replaces the underlying source. This operation is thread safe.
func (broadcaster *Broadcaster) ReplaceSource(source Reader) error {
	if source == nil {
		return errEmptySource
	}

	broadcaster.source.Store(source)
	return nil
}

// ReplaceSource retrieves the underlying source. This operation is thread safe.
func (broadcaster *Broadcaster) Source() Reader {
	return broadcaster.source.Load().(Reader)
}
