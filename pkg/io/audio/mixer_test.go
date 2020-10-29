package audio

import (
	"io"
	"reflect"
	"testing"

	"github.com/pion/mediadevices/pkg/wave"
	"github.com/pion/mediadevices/pkg/wave/mixer"
)

func TestMixer(t *testing.T) {
	input := []wave.Audio{
		&wave.Int16Interleaved{
			Size: wave.ChunkInfo{Len: 1, Channels: 2, SamplingRate: 1234},
			Data: []int16{1, 3},
		},
		&wave.Int16Interleaved{
			Size: wave.ChunkInfo{Len: 3, Channels: 2, SamplingRate: 1234},
			Data: []int16{2, 4, 3, 5, 4, 6},
		},
	}
	expected := []wave.Audio{
		&wave.Int16Interleaved{
			Size: wave.ChunkInfo{Len: 1, Channels: 1, SamplingRate: 1234},
			Data: []int16{2},
		},
		&wave.Int16Interleaved{
			Size: wave.ChunkInfo{Len: 3, Channels: 1, SamplingRate: 1234},
			Data: []int16{3, 4, 5},
		},
	}

	trans := NewChannelMixer(1, &mixer.MonoMixer{})

	var iSent int
	r := trans(ReaderFunc(func() (wave.Audio, func(), error) {
		if iSent < len(input) {
			iSent++
			return input[iSent-1], func() {}, nil
		}
		return nil, func() {}, io.EOF
	}))

	for i := 0; ; i++ {
		a, _, err := r.Read()
		if err != nil {
			if err == io.EOF && i >= len(expected) {
				break
			}
			t.Fatal(err)
		}
		if !reflect.DeepEqual(expected[i], a) {
			t.Errorf("Expected wave[%d]: %v, got: %v", i, expected[i], a)
		}
	}
}
