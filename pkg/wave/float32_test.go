package wave

import (
	"math"
	"reflect"
	"testing"
)

func float32ToUint8(vs ...float32) []uint8 {
	var b []uint8

	for _, v := range vs {
		s := math.Float32bits(v)
		b = append(b, uint8(s>>24), uint8(s>>16), uint8(s>>8), uint8(s))
	}

	return b
}

func TestFloat32(t *testing.T) {
	expected := [][]float32{
		{0.0, 1.0, 2.0, 3.0},
		{4.0, 5.0, 6.0, 7.0},
	}

	cases := map[string]struct {
		in       Audio
		expected [][]float32
	}{
		"Interleaved": {
			in: &Float32Interleaved{
				Data: float32ToUint8(
					0.0, 4.0, 1.0, 5.0,
					2.0, 6.0, 3.0, 7.0,
				),
				Size: ChunkInfo{4, 2, 48000},
			},
			expected: expected,
		},
		"NonInterleaved": {
			in: &Float32NonInterleaved{
				Data: [][]uint8{
					float32ToUint8(expected[0]...),
					float32ToUint8(expected[1]...),
				},
				Size: ChunkInfo{4, 2, 48000},
			},
			expected: expected,
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
			// Data: []uint8{
			// 	1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0,
			// 	9.0, 10.0, 11.0, 12.0, 13.0, 14.0, 15.0, 16.0,
			// },
			Data: float32ToUint8(
				1.0, 2.0, 3.0, 4.0,
				5.0, 6.0, 7.0, 8.0,
			),
			Size: ChunkInfo{4, 2, 48000},
		}
		expected := &Float32Interleaved{
			Data: float32ToUint8(
				5.0, 6.0, 7.0, 8.0,
			),
			Size: ChunkInfo{2, 2, 48000},
		}
		out := in.SubAudio(2, 2)
		if !reflect.DeepEqual(expected, out) {
			t.Errorf("SubAudio differs, expected: %v, got: %v", expected, out)
		}
	})
	t.Run("NonInterleaved", func(t *testing.T) {
		in := &Float32NonInterleaved{
			Data: [][]uint8{
				float32ToUint8(1.0, 2.0, 3.0, 4.0),
				float32ToUint8(5.0, 6.0, 7.0, 8.0),
			},
			Size: ChunkInfo{4, 2, 48000},
		}
		expected := &Float32NonInterleaved{
			Data: [][]uint8{
				float32ToUint8(3.0, 4.0),
				float32ToUint8(7.0, 8.0),
			},
			Size: ChunkInfo{2, 2, 48000},
		}
		out := in.SubAudio(2, 2)
		if !reflect.DeepEqual(expected, out) {
			t.Errorf("SubAudio differs, expected: %v, got: %v", expected, out)
		}
	})
}
