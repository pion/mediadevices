package mixer

import (
	"reflect"
	"testing"

	"github.com/pion/mediadevices/pkg/wave"
)

func TestMonoMixer(t *testing.T) {
	testCases := map[string]struct {
		src      wave.Audio
		dst      wave.Audio
		expected wave.Audio
	}{
		"MultiToMono": {
			src: &wave.Int16Interleaved{
				Size: wave.ChunkInfo{
					Len:      3,
					Channels: 3,
				},
				Data: []int16{
					0, 2, 4,
					1, -2, 1,
					3, 3, 6,
				},
			},
			dst: &wave.Int16Interleaved{
				Size: wave.ChunkInfo{
					Len:      3,
					Channels: 1,
				},
				Data: make([]int16, 3),
			},
			expected: &wave.Int16Interleaved{
				Size: wave.ChunkInfo{
					Len:      3,
					Channels: 1,
				},
				Data: []int16{2, 0, 4},
			},
		},
		"MonoToStereo": {
			src: &wave.Int16Interleaved{
				Size: wave.ChunkInfo{
					Len:      3,
					Channels: 1,
				},
				Data: []int16{0, 2, 4},
			},
			dst: &wave.Int16Interleaved{
				Size: wave.ChunkInfo{
					Len:      3,
					Channels: 2,
				},
				Data: make([]int16, 6),
			},
			expected: &wave.Int16Interleaved{
				Size: wave.ChunkInfo{
					Len:      3,
					Channels: 2,
				},
				Data: []int16{0, 0, 2, 2, 4, 4},
			},
		},
	}
	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			m := &MonoMixer{}
			err := m.Mix(testCase.dst, testCase.src)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(testCase.expected, testCase.dst) {
				t.Errorf("Mix result is wrong\nexpected: %v\ngot: %v", testCase.expected, testCase.dst)
			}
		})
	}
}
