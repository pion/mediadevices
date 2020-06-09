package wave

import (
	"encoding/binary"
	"math"
	"reflect"
	"testing"
)

func TestCalculateChunkInfo(t *testing.T) {
	testCases := map[string]struct {
		chunk      []byte
		channels   int
		sampleSize int
		expected   ChunkInfo
		expectErr  bool
	}{
		"InvalidChunkSize1": {
			chunk:      make([]byte, 3),
			channels:   2,
			sampleSize: 2,
			expected:   ChunkInfo{},
			expectErr:  true,
		},
		"InvalidChunkSize2": {
			chunk:      make([]byte, 4),
			channels:   2,
			sampleSize: 4,
			expected:   ChunkInfo{},
			expectErr:  true,
		},
		"InvalidChannels": {
			chunk:      nil,
			channels:   0,
			sampleSize: 2,
			expected:   ChunkInfo{},
			expectErr:  true,
		},
		"InvalidSampleSize": {
			chunk:      nil,
			channels:   2,
			sampleSize: 0,
			expected:   ChunkInfo{},
			expectErr:  true,
		},
		"Valid1": {
			chunk:      nil,
			channels:   2,
			sampleSize: 2,
			expected: ChunkInfo{
				Len:          0,
				Channels:     2,
				SamplingRate: 0,
			},
			expectErr: false,
		},
		"Valid2": {
			chunk:      make([]byte, 8),
			channels:   2,
			sampleSize: 4,
			expected: ChunkInfo{
				Len:          1,
				Channels:     2,
				SamplingRate: 0,
			},
			expectErr: false,
		},
		"Valid3": {
			chunk:      make([]byte, 4),
			channels:   1,
			sampleSize: 2,
			expected: ChunkInfo{
				Len:          2,
				Channels:     1,
				SamplingRate: 0,
			},
			expectErr: false,
		},
	}

	for testCaseName, testCase := range testCases {
		testCase := testCase
		t.Run(testCaseName, func(t *testing.T) {
			actual, err := calculateChunkInfo(testCase.chunk, testCase.channels, testCase.sampleSize)
			if testCase.expectErr && err == nil {
				t.Fatal("expected an error, but got nil")
			} else if !testCase.expectErr && err != nil {
				t.Fatalf("expected no error, but got %s", err)
			} else if !testCase.expectErr && !reflect.DeepEqual(actual, testCase.expected) {
				t.Errorf("Wrong chunk info calculation result,\nexpected:\n%+v\ngot:\n%+v", testCase.expected, actual)
			}
		})
	}
}

func TestNewDecoder(t *testing.T) {
	rawFormats := []RawFormat{
		{
			SampleSize:  2,
			IsFloat:     false,
			Interleaved: false,
		},
		{
			SampleSize:  4,
			IsFloat:     true,
			Interleaved: false,
		},
		{
			SampleSize:  2,
			IsFloat:     false,
			Interleaved: true,
		},
		{
			SampleSize:  4,
			IsFloat:     true,
			Interleaved: true,
		},
	}

	for _, rawFormat := range rawFormats {
		_, err := NewDecoder(&rawFormat)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestDecodeInt16Interleaved(t *testing.T) {
	raw := []byte{
		// 16 bits per channel
		0x01, 0x02, 0x03, 0x04,
		0x05, 0x06, 0x07, 0x08,
	}
	decoder, _ := newInt16InterleavedDecoder()

	t.Run("BigEndian", func(t *testing.T) {
		expected := &Int16Interleaved{
			Data: []int16{
				int16(binary.BigEndian.Uint16([]byte{0x01, 0x02})),
				int16(binary.BigEndian.Uint16([]byte{0x03, 0x04})),
				int16(binary.BigEndian.Uint16([]byte{0x05, 0x06})),
				int16(binary.BigEndian.Uint16([]byte{0x07, 0x08})),
			},
			Size: ChunkInfo{
				Len:      2,
				Channels: 2,
			},
		}
		actual, err := decoder.Decode(binary.BigEndian, raw, 2)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Wrong decode result,\nexpected:\n%+v\ngot:\n%+v", expected, actual)
		}
	})

	t.Run("LittleEndian", func(t *testing.T) {
		expected := &Int16Interleaved{
			Data: []int16{
				int16(binary.LittleEndian.Uint16([]byte{0x01, 0x02})),
				int16(binary.LittleEndian.Uint16([]byte{0x03, 0x04})),
				int16(binary.LittleEndian.Uint16([]byte{0x05, 0x06})),
				int16(binary.LittleEndian.Uint16([]byte{0x07, 0x08})),
			},
			Size: ChunkInfo{
				Len:      2,
				Channels: 2,
			},
		}
		actual, err := decoder.Decode(binary.LittleEndian, raw, 2)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Wrong decode result,\nexpected:\n%+v\ngot:\n%+v", expected, actual)
		}
	})
}

func TestDecodeInt16NonInterleaved(t *testing.T) {
	raw := []byte{
		// 16 bits per channel
		0x01, 0x02, 0x03, 0x04,
		0x05, 0x06, 0x07, 0x08,
	}

	decoder, _ := newInt16NonInterleavedDecoder()

	t.Run("BigEndian", func(t *testing.T) {
		expected := &Int16NonInterleaved{
			Data: [][]int16{
				{int16(binary.BigEndian.Uint16([]byte{0x01, 0x02})), int16(binary.BigEndian.Uint16([]byte{0x03, 0x04}))},
				{int16(binary.BigEndian.Uint16([]byte{0x05, 0x06})), int16(binary.BigEndian.Uint16([]byte{0x07, 0x08}))},
			},
			Size: ChunkInfo{
				Len:      2,
				Channels: 2,
			},
		}
		actual, err := decoder.Decode(binary.BigEndian, raw, 2)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Wrong decode result,\nexpected:\n%+v\ngot:\n%+v", expected, actual)
		}
	})

	t.Run("LittleEndian", func(t *testing.T) {
		expected := &Int16NonInterleaved{
			Data: [][]int16{
				{int16(binary.LittleEndian.Uint16([]byte{0x01, 0x02})), int16(binary.LittleEndian.Uint16([]byte{0x03, 0x04}))},
				{int16(binary.LittleEndian.Uint16([]byte{0x05, 0x06})), int16(binary.LittleEndian.Uint16([]byte{0x07, 0x08}))},
			},
			Size: ChunkInfo{
				Len:      2,
				Channels: 2,
			},
		}
		actual, err := decoder.Decode(binary.LittleEndian, raw, 2)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Wrong decode result,\nexpected:\n%+v\ngot:\n%+v", expected, actual)
		}
	})
}

func TestDecodeFloat32Interleaved(t *testing.T) {
	raw := []byte{
		// 32 bits per channel
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
	}

	decoder, _ := newFloat32InterleavedDecoder()

	t.Run("BigEndian", func(t *testing.T) {
		expected := &Float32Interleaved{
			Data: []float32{
				math.Float32frombits(binary.BigEndian.Uint32([]byte{0x01, 0x02, 0x03, 0x04})),
				math.Float32frombits(binary.BigEndian.Uint32([]byte{0x05, 0x06, 0x07, 0x08})),
				math.Float32frombits(binary.BigEndian.Uint32([]byte{0x09, 0x0a, 0x0b, 0x0c})),
				math.Float32frombits(binary.BigEndian.Uint32([]byte{0x0d, 0x0e, 0x0f, 0x10})),
			},
			Size: ChunkInfo{
				Len:      2,
				Channels: 2,
			},
		}
		actual, err := decoder.Decode(binary.BigEndian, raw, 2)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Wrong decode result,\nexpected:\n%+v\ngot:\n%+v", expected, actual)
		}
	})

	t.Run("LittleEndian", func(t *testing.T) {
		expected := &Float32Interleaved{
			Data: []float32{
				math.Float32frombits(binary.LittleEndian.Uint32([]byte{0x01, 0x02, 0x03, 0x04})),
				math.Float32frombits(binary.LittleEndian.Uint32([]byte{0x05, 0x06, 0x07, 0x08})),
				math.Float32frombits(binary.LittleEndian.Uint32([]byte{0x09, 0x0a, 0x0b, 0x0c})),
				math.Float32frombits(binary.LittleEndian.Uint32([]byte{0x0d, 0x0e, 0x0f, 0x10})),
			},
			Size: ChunkInfo{
				Len:      2,
				Channels: 2,
			},
		}
		actual, err := decoder.Decode(binary.LittleEndian, raw, 2)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Wrong decode result,\nexpected:\n%+v\ngot:\n%+v", expected, actual)
		}
	})
}

func TestDecodeFloat32NonInterleaved(t *testing.T) {
	raw := []byte{
		// 32 bits per channel
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
	}

	decoder, _ := newFloat32NonInterleavedDecoder()

	t.Run("BigEndian", func(t *testing.T) {
		expected := &Float32NonInterleaved{
			Data: [][]float32{
				{
					math.Float32frombits(binary.BigEndian.Uint32([]byte{0x01, 0x02, 0x03, 0x04})),
					math.Float32frombits(binary.BigEndian.Uint32([]byte{0x05, 0x06, 0x07, 0x08})),
				},
				{
					math.Float32frombits(binary.BigEndian.Uint32([]byte{0x09, 0x0a, 0x0b, 0x0c})),
					math.Float32frombits(binary.BigEndian.Uint32([]byte{0x0d, 0x0e, 0x0f, 0x10})),
				},
			},
			Size: ChunkInfo{
				Len:      2,
				Channels: 2,
			},
		}
		actual, err := decoder.Decode(binary.BigEndian, raw, 2)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Wrong decode result,\nexpected:\n%+v\ngot:\n%+v", expected, actual)
		}
	})

	t.Run("LittleEndian", func(t *testing.T) {
		expected := &Float32NonInterleaved{
			Data: [][]float32{
				{
					math.Float32frombits(binary.LittleEndian.Uint32([]byte{0x01, 0x02, 0x03, 0x04})),
					math.Float32frombits(binary.LittleEndian.Uint32([]byte{0x05, 0x06, 0x07, 0x08})),
				},
				{
					math.Float32frombits(binary.LittleEndian.Uint32([]byte{0x09, 0x0a, 0x0b, 0x0c})),
					math.Float32frombits(binary.LittleEndian.Uint32([]byte{0x0d, 0x0e, 0x0f, 0x10})),
				},
			},
			Size: ChunkInfo{
				Len:      2,
				Channels: 2,
			},
		}
		actual, err := decoder.Decode(binary.LittleEndian, raw, 2)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Wrong decode result,\nexpected:\n%+v\ngot:\n%+v", expected, actual)
		}
	})
}
