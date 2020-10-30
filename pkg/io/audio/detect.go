package audio

import (
	"time"

	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/mediadevices/pkg/wave"
)

// DetectChanges will detect chunk and audio property changes. For audio property detection,
// since it's time related, interval will be used to determine the sample rate.
func DetectChanges(interval time.Duration, onChange func(prop.Media)) TransformFunc {
	return func(r Reader) Reader {
		var currentProp prop.Media
		var chunkCount uint
		return ReaderFunc(func() (wave.Audio, func(), error) {
			var dirty bool

			chunk, _, err := r.Read()
			if err != nil {
				return nil, func() {}, err
			}

			info := chunk.ChunkInfo()
			if currentProp.ChannelCount != info.Channels {
				currentProp.ChannelCount = info.Channels
				dirty = true
			}

			if currentProp.SampleRate != info.SamplingRate {
				currentProp.SampleRate = info.SamplingRate
				dirty = true
			}

			var latency time.Duration
			if currentProp.SampleRate != 0 {
				latency = time.Duration(chunk.ChunkInfo().Len) * time.Second / time.Nanosecond / time.Duration(currentProp.SampleRate)
			}
			if currentProp.Latency != latency {
				currentProp.Latency = latency
				dirty = true
			}

			// TODO: Also detect sample format changes?
			// TODO: Add audio detect changes. As of now, there's no useful property to track.

			if dirty {
				onChange(currentProp)
			}

			chunkCount++
			return chunk, func() {}, nil
		})
	}
}
