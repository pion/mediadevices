package audio

import (
	"io"
	"reflect"
	"testing"

	"github.com/pion/mediadevices/pkg/wave"
	"github.com/pion/mediadevices/pkg/wave/mixer"
)

func TestMixer(t *testing.T) {
	input1 := wave.NewInt16Interleaved(wave.ChunkInfo{Len: 1, Channels: 2, SamplingRate: 1234})
	input1.SetInt16(0, 0, 1)
	input1.SetInt16(0, 1, 3)

	input2 := wave.NewInt16Interleaved(wave.ChunkInfo{Len: 3, Channels: 2, SamplingRate: 1234})
	input2.SetInt16(0, 0, 2)
	input2.SetInt16(0, 1, 4)
	input2.SetInt16(1, 0, 3)
	input2.SetInt16(1, 1, 5)
	input2.SetInt16(2, 0, 4)
	input2.SetInt16(2, 1, 6)

	expected1 := wave.NewInt16Interleaved(wave.ChunkInfo{Len: 1, Channels: 1, SamplingRate: 1234})
	expected1.SetInt16(0, 0, 2)

	expected2 := wave.NewInt16Interleaved(wave.ChunkInfo{Len: 3, Channels: 1, SamplingRate: 1234})
	expected2.SetInt16(0, 0, 3)
	expected2.SetInt16(1, 0, 4)
	expected2.SetInt16(2, 0, 5)

	input := []wave.Audio{
		input1,
		input2,
	}
	expected := []wave.Audio{
		expected1,
		expected2,
	}

	trans := NewChannelMixer(1, &mixer.MonoMixer{})

	var iSent int
	r := trans(ReaderFunc(func() (wave.Audio, error) {
		if iSent < len(input) {
			iSent++
			return input[iSent-1], nil
		}
		return nil, io.EOF
	}))

	for i := 0; ; i++ {
		a, err := r.Read()
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
