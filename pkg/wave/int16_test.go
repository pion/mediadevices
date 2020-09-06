package wave

import (
	"reflect"
	"testing"
)

func TestInt16(t *testing.T) {
	cases := map[string]struct {
		in       Audio
		expected [][]int16
	}{
		"Interleaved": {
			in: &Int16Interleaved{
				Data: []uint8{
					0, 1, 1, 2, 2, 3, 3, 4,
					4, 5, 5, 6, 6, 7, 7, 8,
				},
				Size: ChunkInfo{4, 2, 48000},
			},
			expected: [][]int16{
				{(0 << 8) | 1, (2 << 8) | 3, (4 << 8) | 5, (6 << 8) | 7},
				{(1 << 8) | 2, (3 << 8) | 4, (5 << 8) | 6, (7 << 8) | 8},
			},
		},
		"NonInterleaved": {
			in: &Int16NonInterleaved{
				Data: [][]uint8{
					{0, 1, 2, 3, 4, 5, 6, 7},
					{1, 2, 3, 4, 5, 6, 7, 8},
				},
				Size: ChunkInfo{4, 2, 48000},
			},
			expected: [][]int16{
				{(0 << 8) | 1, (2 << 8) | 3, (4 << 8) | 5, (6 << 8) | 7},
				{(1 << 8) | 2, (3 << 8) | 4, (5 << 8) | 6, (7 << 8) | 8},
			},
		},
	}
	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			out := make([][]int16, c.in.ChunkInfo().Channels)
			for i := 0; i < c.in.ChunkInfo().Channels; i++ {
				for j := 0; j < c.in.ChunkInfo().Len; j++ {
					out[i] = append(out[i], int16(c.in.At(j, i).(Int16Sample)))
				}
			}
			if !reflect.DeepEqual(c.expected, out) {
				t.Errorf("Sample level differs, expected: %v, got: %v", c.expected, out)
			}
		})
	}
}

func TestInt32SubAudio(t *testing.T) {
	t.Run("Interleaved", func(t *testing.T) {
		in := &Int16Interleaved{
			Data: []uint8{
				1, 2, 3, 4, 5, 6, 7, 8,
				9, 10, 11, 12, 13, 14, 15, 16,
			},
			Size: ChunkInfo{4, 2, 48000},
		}
		expected := &Int16Interleaved{
			Data: []uint8{
				9, 10, 11, 12, 13, 14, 15, 16,
			},
			Size: ChunkInfo{2, 2, 48000},
		}
		out := in.SubAudio(2, 2)
		if !reflect.DeepEqual(expected, out) {
			t.Errorf("SubAudio differs, expected: %v, got: %v", expected, out)
		}
	})
	t.Run("NonInterleaved", func(t *testing.T) {
		in := &Int16NonInterleaved{
			Data: [][]uint8{
				{1, 2, 5, 6, 9, 10, 13, 14},
				{3, 4, 7, 8, 11, 12, 15, 16},
			},
			Size: ChunkInfo{4, 2, 48000},
		}
		expected := &Int16NonInterleaved{
			Data: [][]uint8{
				{9, 10, 13, 14},
				{11, 12, 15, 16},
			},
			Size: ChunkInfo{2, 2, 48000},
		}
		out := in.SubAudio(2, 2)
		if !reflect.DeepEqual(expected, out) {
			t.Errorf("SubAudio differs, expected: %v, got: %v", expected, out)
		}
	})
}
