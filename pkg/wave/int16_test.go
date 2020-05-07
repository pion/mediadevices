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
				Data: []int16{
					1, -5, 2, -6, 3, -7, 4, -8, 5, -9, 6, -10, 7, -11, 8, -12,
				},
				Size: ChunkInfo{8, 2, 48000},
			},
			expected: [][]int16{
				{1, 2, 3, 4, 5, 6, 7, 8},
				{-5, -6, -7, -8, -9, -10, -11, -12},
			},
		},
		"NonInterleaved": {
			in: &Int16NonInterleaved{
				Data: [][]int16{
					{1, 2, 3, 4, 5, 6, 7, 8},
					{-5, -6, -7, -8, -9, -10, -11, -12},
				},
				Size: ChunkInfo{8, 2, 48000},
			},
			expected: [][]int16{
				{1, 2, 3, 4, 5, 6, 7, 8},
				{-5, -6, -7, -8, -9, -10, -11, -12},
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
