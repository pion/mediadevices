package audio

import (
	"reflect"
	"testing"
	"time"

	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/mediadevices/pkg/wave"
)

func TestDetectChanges(t *testing.T) {
	buildSource := func(p prop.Media) (Reader, func(prop.Media)) {
		return ReaderFunc(func() (wave.Audio, func(), error) {
				return wave.NewFloat32Interleaved(wave.ChunkInfo{
					Len:          960,
					Channels:     p.ChannelCount,
					SamplingRate: p.SampleRate,
				}), func() {}, nil
			}), func(newProp prop.Media) {
				p = newProp
			}
	}

	t.Run("OnChangeCalledBeforeFirstFrame", func(t *testing.T) {
		var detectBeforeFirstChunk bool
		var expected prop.Media
		var actual prop.Media
		expected.ChannelCount = 2
		expected.SampleRate = 48000
		expected.Latency = time.Millisecond * 20
		src, _ := buildSource(expected)
		src = DetectChanges(time.Second, func(p prop.Media) {
			actual = p
			detectBeforeFirstChunk = true
		})(src)

		_, _, err := src.Read()
		if err != nil {
			t.Fatal(err)
		}

		if !detectBeforeFirstChunk {
			t.Fatal("on change callback should have called before first chunk")
		}

		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("Received an unexpected prop\nExpected:\n%v\nActual:\n%v\n", expected, actual)
		}
	})

	t.Run("DetectChangesOnEveryUpdate", func(t *testing.T) {
		var expected prop.Media
		var actual prop.Media
		expected.ChannelCount = 2
		expected.SampleRate = 48000
		expected.Latency = 20 * time.Millisecond
		src, update := buildSource(expected)
		src = DetectChanges(time.Second, func(p prop.Media) {
			actual = p
		})(src)

		for channelCount := 1; channelCount < 8; channelCount++ {
			expected.ChannelCount = channelCount
			update(expected)
			_, _, err := src.Read()
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("Received an unexpected prop\nExpected:\n%v\nActual:\n%v\n", expected, actual)
			}
		}
	})
}
