// +build e2e

package mediadevices

import (
	"image"
	"sync"
	"testing"

	"github.com/pion/mediadevices/pkg/codec/x264"
	"github.com/pion/mediadevices/pkg/frame"
)

type mockVideoSource struct {
	width, height int
	pool          sync.Pool
	decoder       frame.Decoder
}

func newMockVideoSource(width, height int) *mockVideoSource {
	decoder, err := frame.NewDecoder(frame.FormatYUY2)
	if err != nil {
		panic(err)
	}

	return &mockVideoSource{
		width:  width,
		height: height,
		pool: sync.Pool{
			New: func() interface{} {
				resolution := width * height
				return make([]byte, resolution*2)
			},
		},
		decoder: decoder,
	}
}

func (source *mockVideoSource) ID() string   { return "" }
func (source *mockVideoSource) Close() error { return nil }
func (source *mockVideoSource) Read() (image.Image, func(), error) {
	raw := source.pool.Get().([]byte)
	decoded, release, err := source.decoder.Decode(raw, source.width, source.height)
	source.pool.Put(raw)
	if err != nil {
		return nil, nil, err
	}

	return decoded, release, nil
}

func BenchmarkEndToEnd(b *testing.B) {
	params, err := x264.NewParams()
	if err != nil {
		b.Fatal(err)
	}
	params.BitRate = 300_000

	videoSource := newMockVideoSource(1920, 1080)
	track := NewVideoTrack(videoSource, nil).(*VideoTrack)
	defer track.Close()

	reader := track.NewReader(false)
	inputProp, err := detectCurrentVideoProp(track.Broadcaster)
	if err != nil {
		b.Fatal(err)
	}

	encodedReader, err := params.BuildVideoEncoder(reader, inputProp)
	if err != nil {
		b.Fatal(err)
	}
	defer encodedReader.Close()

	for i := 0; i < b.N; i++ {
		_, release, err := encodedReader.Read()
		if err != nil {
			b.Fatal(err)
		}
		release()
	}
}
