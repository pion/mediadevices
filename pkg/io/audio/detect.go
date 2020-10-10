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
		return ReaderFunc(func() (wave.Audio, error) {
			var dirty bool

			chunk, err := r.Read()
			if err != nil {
				return nil, err
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

			// TODO: Also detect sample format changes?
			// TODO: Add audio detect changes. As of now, there's no useful property to track.

			if dirty {
				onChange(currentProp)
			}

			chunkCount++
			return chunk, nil
		})
	}
}
