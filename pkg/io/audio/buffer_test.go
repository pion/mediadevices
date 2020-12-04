package audio

import (
	"io"
	"reflect"
	"testing"

	"github.com/pion/mediadevices/pkg/wave"
)

func TestBuffer(t *testing.T) {
	input := []wave.Audio{
		&wave.Int16Interleaved{
			Size: wave.ChunkInfo{Len: 1, Channels: 2, SamplingRate: 1234},
			Data: []int16{1, 2},
		},
		&wave.Int16Interleaved{
			Size: wave.ChunkInfo{Len: 3, Channels: 2, SamplingRate: 1234},
			Data: []int16{3, 4, 5, 6, 7, 8},
		},
		&wave.Int16Interleaved{
			Size: wave.ChunkInfo{Len: 2, Channels: 2, SamplingRate: 1234},
			Data: []int16{9, 10, 11, 12},
		},
		&wave.Int16Interleaved{
			Size: wave.ChunkInfo{Len: 7, Channels: 2, SamplingRate: 1234},
			Data: []int16{13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26},
		},
	}
	expected := []wave.Audio{
		&wave.Int16Interleaved{
			Size: wave.ChunkInfo{Len: 3, Channels: 2, SamplingRate: 1234},
			Data: []int16{1, 2, 3, 4, 5, 6},
		},
		&wave.Int16Interleaved{
			Size: wave.ChunkInfo{Len: 3, Channels: 2, SamplingRate: 1234},
			Data: []int16{7, 8, 9, 10, 11, 12},
		},
		&wave.Int16Interleaved{
			Size: wave.ChunkInfo{Len: 3, Channels: 2, SamplingRate: 1234},
			Data: []int16{13, 14, 15, 16, 17, 18},
		},
		&wave.Int16Interleaved{
			Size: wave.ChunkInfo{Len: 3, Channels: 2, SamplingRate: 1234},
			Data: []int16{19, 20, 21, 22, 23, 24},
		},
	}

	trans := NewBuffer(3)

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
