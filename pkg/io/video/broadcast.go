package video

import (
	"fmt"
	"image"

	"github.com/pion/mediadevices/pkg/io"
)

var errEmptySource = fmt.Errorf("Source can't be nil")

// Broadcaster is a specialized video broadcaster.
type Broadcaster struct {
	ioBroadcaster *io.Broadcaster
}

// NewNewBroadcaster creates a new broadcaster.
func NewBroadcaster(source Reader) *Broadcaster {
	broadcaster := io.NewBroadcaster(io.ReaderFunc(func() (interface{}, error) {
		return source.Read()
	}))

	return &Broadcaster{broadcaster}
}

// NewReader creates a new reader. Each reader will retrieve the same data from the source.
// copyFn is used to copy the data from the source to individual readers. Broadcaster uses a small ring
// buffer, this means that slow readers might miss some data if they're really late and the data is no longer
// in the ring buffer.
func (broadcaster *Broadcaster) NewReader(copyFrame bool) Reader {
	copyFn := func(src interface{}) interface{} { return src }

	if copyFrame {
		buffer := NewFrameBuffer(0)
		copyFn = func(src interface{}) interface{} {
			realSrc, _ := src.(image.Image)
			buffer.StoreCopy(realSrc)
			return buffer.Load()
		}
	}

	reader := broadcaster.ioBroadcaster.NewReader(copyFn)
	return ReaderFunc(func() (image.Image, error) {
		data, err := reader.Read()
		img, _ := data.(image.Image)
		return img, err
	})
}

// ReplaceSource replaces the underlying source. This operation is thread safe.
func (broadcaster *Broadcaster) ReplaceSource(source Reader) error {
	return broadcaster.ioBroadcaster.ReplaceSource(io.ReaderFunc(func() (interface{}, error) {
		return source.Read()
	}))
}

// ReplaceSource retrieves the underlying source. This operation is thread safe.
func (broadcaster *Broadcaster) Source() Reader {
	source := broadcaster.ioBroadcaster.Source()
	return ReaderFunc(func() (image.Image, error) {
		data, err := source.Read()
		img, _ := data.(image.Image)
		return img, err
	})
}
