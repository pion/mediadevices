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
				Data: []uint8{
					0x00, 0x01, 0x00, 0x02, 0x00, 0x04,
					0x00, 0x01, 0x00, 0x02, 0x00, 0x01,
					0x00, 0x03, 0x00, 0x03, 0x00, 0x06,
				},
			},
			dst: &wave.Int16Interleaved{
				Size: wave.ChunkInfo{
					Len:      3,
					Channels: 1,
				},
				Data: make([]uint8, 3*2),
			},
			expected: &wave.Int16Interleaved{
				Size: wave.ChunkInfo{
					Len:      3,
					Channels: 1,
				},
				Data: []uint8{0x00, 0x02, 0x00, 0x01, 0x00, 0x04},
			},
		},
		"MonoToStereo": {
			src: &wave.Int16Interleaved{
				Size: wave.ChunkInfo{
					Len:      3,
					Channels: 1,
				},
				Data: []uint8{0x00, 0x00, 0x00, 0x02, 0x00, 0x04},
			},
			dst: &wave.Int16Interleaved{
				Size: wave.ChunkInfo{
					Len:      3,
					Channels: 2,
				},
				Data: make([]uint8, 6*2),
			},
			expected: &wave.Int16Interleaved{
				Size: wave.ChunkInfo{
					Len:      3,
					Channels: 2,
				},
				Data: []uint8{0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x02, 0x00, 0x04, 0x00, 0x04},
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
