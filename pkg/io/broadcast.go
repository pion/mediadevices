package io

import (
	"fmt"
	"sync/atomic"
	"time"
)

const (
	maskReading                = 1 << 63
	defaultBroadcasterRingSize = 32
	// TODO: If the data source has fps greater than 30, they'll see some
	// 			 fps fluctuation. But, 30 fps should be enough for general cases.
	defaultBroadcasterRingPollDuration = time.Millisecond * 33
)

var errEmptySource = fmt.Errorf("Source can't be nil")

type broadcasterData struct {
	data  interface{}
	count uint32
	err   error
}

type broadcasterRing struct {
	// reading (1 bit) + reserved (31 bits) + data count (32 bits)
	// IMPORTANT: state has to be the first element in struct, otherwise LoadUint64 will panic in 32 bits systems
	//            due to unallignment
	state        uint64
	buffer       []atomic.Value
	pollDuration time.Duration
}

func newBroadcasterRing(size uint, pollDuration time.Duration) *broadcasterRing {
	return &broadcasterRing{buffer: make([]atomic.Value, size), pollDuration: pollDuration}
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
			time.Sleep(ring.pollDuration)
		}

		i := ring.index(count)
		data := ring.buffer[i].Load().(*broadcasterData)
		if data.count == count {
			return data
		}

		count++
	}
}

func (ring *broadcasterRing) lastCount() uint32 {
	// ring.state always keeps track the next count, so we need to subtract it by 1 to get the
	// last count
	return uint32(atomic.LoadUint64(&ring.state)) - 1
}

// Broadcaster is a generic pull-based broadcaster. Broadcaster is unique in a sense that
// readers can come and go at anytime, and readers don't need to close or notify broadcaster.
type Broadcaster struct {
	source atomic.Value
	buffer *broadcasterRing
}

// BroadcasterConfig is a config to control broadcaster behaviour
type BroadcasterConfig struct {
	// BufferSize configures the underlying ring buffer size that's being used
	// to avoid data lost for late readers. The default value is 32.
	BufferSize uint
	// PollDuration configures the sleep duration in waiting for new data to come.
	// The default value is 33 ms.
	PollDuration time.Duration
}

// NewBroadcaster creates a new broadcaster. Source is expected to drop frames
// when any of the readers is slower than the source.
func NewBroadcaster(source Reader, config *BroadcasterConfig) *Broadcaster {
	pollDuration := defaultBroadcasterRingPollDuration
	var bufferSize uint = defaultBroadcasterRingSize
	if config != nil {
		if config.PollDuration != 0 {
			pollDuration = config.PollDuration
		}

		if config.BufferSize != 0 {
			bufferSize = config.BufferSize
		}
	}

	var broadcaster Broadcaster
	broadcaster.buffer = newBroadcasterRing(bufferSize, pollDuration)
	broadcaster.ReplaceSource(source)

	return &broadcaster
}

// NewReader creates a new reader. Each reader will retrieve the same data from the source.
// copyFn is used to copy the data from the source to individual readers. Broadcaster uses a small ring
// buffer, this means that slow readers might miss some data if they're really late and the data is no longer
// in the ring buffer.
func (broadcaster *Broadcaster) NewReader(copyFn func(interface{}) interface{}) Reader {
	currentCount := broadcaster.buffer.lastCount()

	return ReaderFunc(func() (data interface{}, release func(), err error) {
		currentCount++
		if push := broadcaster.buffer.acquire(currentCount); push != nil {
			data, _, err = broadcaster.source.Load().(Reader).Read()
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
