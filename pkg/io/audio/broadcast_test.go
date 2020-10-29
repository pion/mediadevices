package audio

import (
	"reflect"
	"testing"

	"github.com/pion/mediadevices/pkg/wave"
)

func TestBroadcast(t *testing.T) {
	chunk := wave.NewFloat32Interleaved(wave.ChunkInfo{
		Len:          8,
		Channels:     2,
		SamplingRate: 48000,
	})

	source := ReaderFunc(func() (wave.Audio, func(), error) {
		return chunk, func() {}, nil
	})

	broadcaster := NewBroadcaster(source, nil)
	readerWithoutCopy1 := broadcaster.NewReader(false)
	readerWithoutCopy2 := broadcaster.NewReader(false)
	actualWithoutCopy1, _, err := readerWithoutCopy1.Read()
	if err != nil {
		t.Fatal(err)
	}
	actualWithoutCopy2, _, err := readerWithoutCopy2.Read()
	if err != nil {
		t.Fatal(err)
	}

	if &actualWithoutCopy1.(*wave.Float32Interleaved).Data[0] != &actualWithoutCopy2.(*wave.Float32Interleaved).Data[0] {
		t.Fatal("Expected underlying buffer for frame with copy to be the same from broadcaster's buffer")
	}

	if !reflect.DeepEqual(chunk, actualWithoutCopy1) {
		t.Fatal("Expected actual frame without copy to be the same with the original")
	}

	readerWithCopy := broadcaster.NewReader(true)
	actualWithCopy, _, err := readerWithCopy.Read()
	if err != nil {
		t.Fatal(err)
	}

	if &actualWithCopy.(*wave.Float32Interleaved).Data[0] == &actualWithoutCopy1.(*wave.Float32Interleaved).Data[0] {
		t.Fatal("Expected underlying buffer for frame with copy to be different from broadcaster's buffer")
	}

	if !reflect.DeepEqual(chunk, actualWithCopy) {
		t.Fatal("Expected actual frame without copy to be the same with the original")
	}
}
