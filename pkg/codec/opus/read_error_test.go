//go:build opustest

package opus

import (
	"testing"

	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/wave"
)

// TestReadEncodeErrorDoesNotPanic reproduces a panic in encoder.Read: when
// opus_encode fails (returns a negative error code), the code still evaluates
// encoded[:n:n] with a negative index, which panics with "slice bounds out of
// range". The expected behavior is to return the error instead.
//
// Trigger: opus_encode returns OPUS_BAD_ARG when frame_size is invalid. At
// 48kHz the minimum frame size is 120 samples and it must be a multiple of
// 120, so 121 samples deterministically produces a negative return value.
func TestReadEncodeErrorDoesNotPanic(t *testing.T) {
	engine := newTestEncoder(48000, 1)
	if engine == nil {
		t.Skip("failed to create opus encoder engine")
	}
	defer destroyTestEncoder(engine)

	reader := audio.ReaderFunc(func() (wave.Audio, func(), error) {
		return wave.NewInt16Interleaved(wave.ChunkInfo{
			Len:          121, // invalid frame size for 48kHz: must be a multiple of 120
			SamplingRate: 48000,
			Channels:     1,
		}), func() {}, nil
	})

	e := &encoder{engine: engine, reader: reader}

	data, release, err := e.Read()
	if release != nil {
		release()
	}

	if err == nil {
		t.Fatal("expected an error when opus_encode fails, got nil")
	}
	if data != nil {
		t.Fatalf("expected nil data on encode failure, got %d bytes", len(data))
	}
}
