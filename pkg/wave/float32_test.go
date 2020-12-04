package wave

import (
	"reflect"
	"testing"
)

func TestFloat32(t *testing.T) {
	cases := map[string]struct {
		in       Audio
		expected [][]float32
	}{
		"Interleaved": {
			in: &Float32Interleaved{
				Data: []float32{
					0.1, -0.5, 0.2, -0.6, 0.3, -0.7, 0.4, -0.8, 0.5, -0.9, 0.6, -1.0, 0.7, -1.1, 0.8, -1.2,
				},
				Size: ChunkInfo{8, 2, 48000},
			},
			expected: [][]float32{
				{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8},
				{-0.5, -0.6, -0.7, -0.8, -0.9, -1.0, -1.1, -1.2},
			},
		},
		"NonInterleaved": {
			in: &Float32NonInterleaved{
				Data: [][]float32{
					{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8},
					{-0.5, -0.6, -0.7, -0.8, -0.9, -1.0, -1.1, -1.2},
				},
				Size: ChunkInfo{8, 2, 48000},
			},
			expected: [][]float32{
				{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8},
				{-0.5, -0.6, -0.7, -0.8, -0.9, -1.0, -1.1, -1.2},
			},
		},
	}
	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			out := make([][]float32, c.in.ChunkInfo().Channels)
			for i := 0; i < c.in.ChunkInfo().Channels; i++ {
				for j := 0; j < c.in.ChunkInfo().Len; j++ {
					out[i] = append(out[i], float32(c.in.At(j, i).(Float32Sample)))
				}
			}
			if !reflect.DeepEqual(c.expected, out) {
				t.Errorf("Sample level differs, expected: %v, got: %v", c.expected, out)
			}
		})
	}
}

func TestFloat32SubAudio(t *testing.T) {
	t.Run("Interleaved", func(t *testing.T) {
		in := &Float32Interleaved{
			Data: []float32{
				0.1, -0.5, 0.2, -0.6, 0.3, -0.7, 0.4, -0.8, 0.5, -0.9, 0.6, -1.0, 0.7, -1.1, 0.8, -1.2,
			},
			Size: ChunkInfo{8, 2, 48000},
		}
		expected := &Float32Interleaved{
			Data: []float32{
				0.3, -0.7, 0.4, -0.8, 0.5, -0.9,
			},
			Size: ChunkInfo{3, 2, 48000},
		}
		out := in.SubAudio(2, 3)
		if !reflect.DeepEqual(expected, out) {
			t.Errorf("SubAudio differs, expected: %v, got: %v", expected, out)
		}
	})
	t.Run("NonInterleaved", func(t *testing.T) {
		in := &Float32NonInterleaved{
			Data: [][]float32{
				{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8},
				{-0.5, -0.6, -0.7, -0.8, -0.9, -1.0, -1.1, -1.2},
			},
			Size: ChunkInfo{8, 2, 48000},
		}
		expected := &Float32NonInterleaved{
			Data: [][]float32{
				{0.3, 0.4, 0.5},
				{-0.7, -0.8, -0.9},
			},
			Size: ChunkInfo{3, 2, 48000},
		}
		out := in.SubAudio(2, 3)
		if !reflect.DeepEqual(expected, out) {
			t.Errorf("SubAudio differs, expected: %v, got: %v", expected, out)
		}
	})
}
