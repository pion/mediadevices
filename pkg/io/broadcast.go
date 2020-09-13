package io

import (
	"fmt"
	"runtime"
	"sync/atomic"
)

const (
	maskReading         = 1 << 63
	broadcasterRingSize = 128
)

var errEmptySource = fmt.Errorf("Source can't be nil")

type broadcasterData struct {
	data  interface{}
	count uint32
	err   error
}

type broadcasterRing struct {
	buffer []atomic.Value
	// reading (1 bit) + reserved (31 bits) + data count (32 bits)
	state uint64
}

func newBroadcasterRing() *broadcasterRing {
	return &broadcasterRing{buffer: make([]atomic.Value, broadcasterRingSize)}
}

func (ring *broadcasterRing) index(count uint32) int {
	return int(count) % len(ring.buffer)
}

func (ring *broadcasterRing) acquire(count uint32) func(*broadcasterData) {
	// Reader has reached the latest data, should read from the source.
	// Only allow 1 reader to read from the source. When there are more than 1 readers,
	// the other readers will need to share the same data that the first reader gets from
	// the source.
	state := uint64(count)
	if atomic.CompareAndSwapUint64(&ring.state, state, state|maskReading) {
		return func(data *broadcasterData) {
			i := ring.index(count)
			ring.buffer[i].Store(data)
			atomic.StoreUint64(&ring.state, uint64(count+1))
		}
	}

	return nil
}

func (ring *broadcasterRing) get(count uint32) *broadcasterData {
	for {
		reading := uint64(count) | maskReading
		// TODO: since it's lockless, it spends a lot of resources in the scheduling.
		for atomic.LoadUint64(&ring.state) == reading {
			// Yield current goroutine to let other goroutines to run instead
			runtime.Gosched()
		}

		i := ring.index(count)
		data := ring.buffer[i].Load().(*broadcasterData)
		if data.count == count {
			return data
		}

		count++
	}
}

// Broadcaster is a generic pull-based broadcaster. Broadcaster is unique in a sense that
// readers can come and go at anytime, and readers don't need to close or notify broadcaster.
type Broadcaster struct {
	source atomic.Value
	buffer *broadcasterRing
}

// NewNewBroadcaster creates a new broadcaster.
func NewBroadcaster(source Reader) *Broadcaster {
	var broadcaster Broadcaster
	broadcaster.buffer = newBroadcasterRing()
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
	// since ring buffer's state starts with 0, each reader count should start with MAX(uint32),
	// so that it'll get incremented to zero (with wrap around)
	currentCount--

	return ReaderFunc(func() (data interface{}, err error) {
		currentCount++
		if push := broadcaster.buffer.acquire(currentCount); push != nil {
			data, err = broadcaster.source.Load().(Reader).Read()
			push(&broadcasterData{
				data:  data,
				err:   err,
				count: currentCount,
			})
		} else {
			ringData := broadcaster.buffer.get(currentCount)
			data, err, currentCount = ringData.data, ringData.err, ringData.count
		}

		data = copyFn(data)
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
