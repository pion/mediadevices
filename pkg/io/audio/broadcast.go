package audio

import (
	"errors"

	"github.com/pion/mediadevices/pkg/io"
	"github.com/pion/mediadevices/pkg/wave"
)

var errEmptySource = errors.New("Source can't be nil")

// Broadcaster is a specialized video broadcaster.
type Broadcaster struct {
	ioBroadcaster *io.Broadcaster
}

type BroadcasterConfig struct {
	Core *io.BroadcasterConfig
}

// NewBroadcaster creates a new broadcaster. Source is expected to drop chunks
// when any of the readers is slower than the source.
func NewBroadcaster(source Reader, config *BroadcasterConfig) *Broadcaster {
	var coreConfig *io.BroadcasterConfig

	if config != nil {
		coreConfig = config.Core
	}

	broadcaster := io.NewBroadcaster(io.ReaderFunc(func() (interface{}, func(), error) {
		return source.Read()
	}), coreConfig)

	return &Broadcaster{broadcaster}
}

// NewReader creates a new reader. Each reader will retrieve the same data from the source.
// copyFn is used to copy the data from the source to individual readers. Broadcaster uses a small ring
// buffer, this means that slow readers might miss some data if they're really late and the data is no longer
// in the ring buffer.
func (broadcaster *Broadcaster) NewReader(copyChunk bool) Reader {
	copyFn := func(src interface{}) interface{} { return src }

	if copyChunk {
		buffer := wave.NewBuffer()
		copyFn = func(src interface{}) interface{} {
			realSrc, _ := src.(wave.Audio)
			buffer.StoreCopy(realSrc)
			return buffer.Load()
		}
	}

	reader := broadcaster.ioBroadcaster.NewReader(copyFn)
	return ReaderFunc(func() (wave.Audio, func(), error) {
		data, _, err := reader.Read()
		chunk, _ := data.(wave.Audio)
		return chunk, func() {}, err
	})
}

// ReplaceSource replaces the underlying source. This operation is thread safe.
func (broadcaster *Broadcaster) ReplaceSource(source Reader) error {
	return broadcaster.ioBroadcaster.ReplaceSource(io.ReaderFunc(func() (interface{}, func(), error) {
		return source.Read()
	}))
}

// Source retrieves the underlying source. This operation is thread safe.
func (broadcaster *Broadcaster) Source() Reader {
	source := broadcaster.ioBroadcaster.Source()
	return ReaderFunc(func() (wave.Audio, func(), error) {
		data, _, err := source.Read()
		img, _ := data.(wave.Audio)
		return img, func() {}, err
	})
}
