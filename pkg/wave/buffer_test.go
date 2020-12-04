package wave

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

var (
	errIdenticalAddress = errors.New("Cloned audio has the same memory address with the original audio")
)

func TestBufferStoreCopyAndLoad(t *testing.T) {
	chunkInfo := ChunkInfo{
		Len:          4,
		Channels:     2,
		SamplingRate: 48000,
	}
	testCases := map[string]struct {
		New      func() EditableAudio
		Update   func(EditableAudio)
		Validate func(*testing.T, Audio, Audio)
	}{
		"Float32Interleaved": {
			New: func() EditableAudio {
				return NewFloat32Interleaved(chunkInfo)
			},
			Update: func(src EditableAudio) {
				src.Set(0, 0, Float32Sample(1))
			},
			Validate: func(t *testing.T, original Audio, clone Audio) {
				ok := reflect.ValueOf(original.(*Float32Interleaved).Data).Pointer() != reflect.ValueOf(clone.(*Float32Interleaved).Data).Pointer()
				if !ok {
					t.Error(errIdenticalAddress)
				}
			},
		},
		"Float32NonInterleaved": {
			New: func() EditableAudio {
				return NewFloat32NonInterleaved(chunkInfo)
			},
			Update: func(src EditableAudio) {
				src.Set(0, 0, Float32Sample(1))
			},
			Validate: func(t *testing.T, original Audio, clone Audio) {
				originalReal := original.(*Float32NonInterleaved)
				cloneReal := clone.(*Float32NonInterleaved)
				if reflect.ValueOf(originalReal.Data).Pointer() == reflect.ValueOf(cloneReal.Data).Pointer() {
					t.Error(errIdenticalAddress)
				}

				for i := range cloneReal.Data {
					if reflect.ValueOf(originalReal.Data[i]).Pointer() == reflect.ValueOf(cloneReal.Data[i]).Pointer() {
						err := fmt.Errorf("Channel %d memory address should be different", i)
						t.Errorf("%v: %w", errIdenticalAddress, err)
					}
				}
			},
		},
		"Int16Interleaved": {
			New: func() EditableAudio {
				return NewInt16Interleaved(chunkInfo)
			},
			Update: func(src EditableAudio) {
				src.Set(1, 1, Int16Sample(2))
			},
			Validate: func(t *testing.T, original Audio, clone Audio) {
				ok := reflect.ValueOf(original.(*Int16Interleaved).Data).Pointer() != reflect.ValueOf(clone.(*Int16Interleaved).Data).Pointer()
				if !ok {
					t.Error(errIdenticalAddress)
				}
			},
		},
		"Int16NonInterleaved": {
			New: func() EditableAudio {
				return NewInt16NonInterleaved(chunkInfo)
			},
			Update: func(src EditableAudio) {
				src.Set(1, 1, Int16Sample(2))
			},
			Validate: func(t *testing.T, original Audio, clone Audio) {
				originalReal := original.(*Int16NonInterleaved)
				cloneReal := clone.(*Int16NonInterleaved)
				if reflect.ValueOf(originalReal.Data).Pointer() == reflect.ValueOf(cloneReal.Data).Pointer() {
					t.Error(errIdenticalAddress)
				}

				for i := range cloneReal.Data {
					if reflect.ValueOf(originalReal.Data[i]).Pointer() == reflect.ValueOf(cloneReal.Data[i]).Pointer() {
						err := fmt.Errorf("Channel %d memory address should be different", i)
						t.Errorf("%v: %w", errIdenticalAddress, err)
					}
				}
			},
		},
	}

	buffer := NewBuffer()

	for name, testCase := range testCases {
		// Since the test also wants to make sure that Copier can convert from 1 type to another,
		// t.Run is not ideal since it'll run the tests separately
		t.Log("Testing", name)

		src := testCase.New()
		src.Set(0, 0, Int16Sample(1))
		buffer.StoreCopy(src)

		testCase.Validate(t, src, buffer.Load())

		if !reflect.DeepEqual(buffer.Load(), src) {
			t.Fatalf(`Expected the copied audio chunk to be identical with the source

Expected:
%v

Actual:
%v
			`, src, buffer.Load())
		}

		testCase.Update(src)
		buffer.StoreCopy(src)
		if !reflect.DeepEqual(buffer.Load(), src) {
			t.Fatal("Expected the copied audio chunk to be identical with the source after an update in source")
		}
	}
}
