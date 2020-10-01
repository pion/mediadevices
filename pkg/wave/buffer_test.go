package wave

import (
	"reflect"
	"testing"
)

func TestBufferStoreCopyAndLoad(t *testing.T) {
	chunkInfo := ChunkInfo{
		Len:          4,
		Channels:     2,
		SamplingRate: 48000,
	}
	testCases := map[string]struct {
		New    func() EditableAudio
		Update func(EditableAudio)
	}{
		"Float32Interleaved": {
			New: func() EditableAudio {
				return NewFloat32Interleaved(chunkInfo)
			},
			Update: func(src EditableAudio) {
				src.Set(0, 0, Float32Sample(1))
			},
		},
		"Float32NonInterleaved": {
			New: func() EditableAudio {
				return NewFloat32NonInterleaved(chunkInfo)
			},
			Update: func(src EditableAudio) {
				src.Set(0, 0, Float32Sample(1))
			},
		},
		"Int16Interleaved": {
			New: func() EditableAudio {
				return NewInt16Interleaved(chunkInfo)
			},
			Update: func(src EditableAudio) {
				src.Set(1, 1, Int16Sample(2))
			},
		},
		"Int16NonInterleaved": {
			New: func() EditableAudio {
				return NewInt16NonInterleaved(chunkInfo)
			},
			Update: func(src EditableAudio) {
				src.Set(1, 1, Int16Sample(2))
			},
		},
	}

	buffer := NewBuffer()

	for name, testCase := range testCases {
		// Since the test also wants to make sure that Copier can convert from 1 type to another,
		// t.Run is not ideal since it'll run the tests separately
		t.Log("Testing", name)

		src := testCase.New()
		buffer.StoreCopy(src)
		if !reflect.DeepEqual(buffer.Load(), src) {
			t.Fatal("Expected the copied audio chunk to be identical with the source")
		}

		testCase.Update(src)
		buffer.StoreCopy(src)
		if !reflect.DeepEqual(buffer.Load(), src) {
			t.Fatal("Expected the copied audio chunk to be identical with the source after an update in source")
		}
	}
}
