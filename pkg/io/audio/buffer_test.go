package audio

import (
	"io"
	"reflect"
	"testing"

	"github.com/pion/mediadevices/pkg/wave"
)

func TestBuffer(t *testing.T) {
	input1 := wave.NewInt16Interleaved(wave.ChunkInfo{Len: 1, Channels: 2, SamplingRate: 1234})
	input1.SetInt16(0, 0, 1)
	input1.SetInt16(0, 1, 2)

	input2 := wave.NewInt16Interleaved(wave.ChunkInfo{Len: 3, Channels: 2, SamplingRate: 1234})
	input2.SetInt16(0, 0, 3)
	input2.SetInt16(0, 1, 4)
	input2.SetInt16(1, 0, 5)
	input2.SetInt16(1, 1, 6)
	input2.SetInt16(2, 0, 7)
	input2.SetInt16(2, 1, 8)

	input3 := wave.NewInt16Interleaved(wave.ChunkInfo{Len: 2, Channels: 2, SamplingRate: 1234})
	input3.SetInt16(0, 0, 9)
	input3.SetInt16(0, 1, 10)
	input3.SetInt16(1, 0, 11)
	input3.SetInt16(1, 1, 12)

	input4 := wave.NewInt16Interleaved(wave.ChunkInfo{Len: 7, Channels: 2, SamplingRate: 1234})
	input4.SetInt16(0, 0, 13)
	input4.SetInt16(0, 1, 14)
	input4.SetInt16(1, 0, 15)
	input4.SetInt16(1, 1, 16)
	input4.SetInt16(2, 0, 17)
	input4.SetInt16(2, 1, 18)
	input4.SetInt16(3, 0, 19)
	input4.SetInt16(3, 1, 20)
	input4.SetInt16(4, 0, 21)
	input4.SetInt16(4, 1, 22)
	input4.SetInt16(5, 0, 23)
	input4.SetInt16(5, 1, 24)
	input4.SetInt16(6, 0, 25)
	input4.SetInt16(6, 1, 26)

	expected1 := wave.NewInt16Interleaved(wave.ChunkInfo{Len: 3, Channels: 2, SamplingRate: 1234})
	expected1.SetInt16(0, 0, 1)
	expected1.SetInt16(0, 1, 2)
	expected1.SetInt16(1, 0, 3)
	expected1.SetInt16(1, 1, 4)
	expected1.SetInt16(2, 0, 5)
	expected1.SetInt16(2, 1, 6)

	expected2 := wave.NewInt16Interleaved(wave.ChunkInfo{Len: 3, Channels: 2, SamplingRate: 1234})
	expected2.SetInt16(0, 0, 7)
	expected2.SetInt16(0, 1, 8)
	expected2.SetInt16(1, 0, 9)
	expected2.SetInt16(1, 1, 10)
	expected2.SetInt16(2, 0, 11)
	expected2.SetInt16(2, 1, 12)

	expected3 := wave.NewInt16Interleaved(wave.ChunkInfo{Len: 3, Channels: 2, SamplingRate: 1234})
	expected3.SetInt16(0, 0, 13)
	expected3.SetInt16(0, 1, 14)
	expected3.SetInt16(1, 0, 15)
	expected3.SetInt16(1, 1, 16)
	expected3.SetInt16(2, 0, 17)
	expected3.SetInt16(2, 1, 18)

	expected4 := wave.NewInt16Interleaved(wave.ChunkInfo{Len: 3, Channels: 2, SamplingRate: 1234})
	expected4.SetInt16(0, 0, 19)
	expected4.SetInt16(0, 1, 20)
	expected4.SetInt16(1, 0, 21)
	expected4.SetInt16(1, 1, 22)
	expected4.SetInt16(2, 0, 23)
	expected4.SetInt16(2, 1, 24)

	input := []wave.Audio{
		input1,
		input2,
		input3,
		input4,
	}
	expected := []wave.Audio{
		expected1,
		expected2,
		expected3,
		expected4,
	}

	trans := NewBuffer(3)

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
