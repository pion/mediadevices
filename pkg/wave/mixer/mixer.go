package mixer

import (
	"errors"

	"github.com/pion/mediadevices/pkg/wave"
)

// ChannelMixer mixes audio into specifix channels.
type ChannelMixer interface {
	Mix(dst wave.Audio, src wave.Audio) error
}

// MonoMixer mixes channels into monaural audio.
type MonoMixer struct {
}

func (m *MonoMixer) Mix(dst wave.Audio, src wave.Audio) error {
	if dst.ChunkInfo().Len != src.ChunkInfo().Len {
		return errors.New("buffer size mismatch")
	}
	dstSetter, ok := dst.(wave.EditableAudio)
	if !ok {
		return errors.New("destination buffer is not settable")
	}

	n := src.ChunkInfo().Len
	channels := src.ChunkInfo().Channels
	dstChannels := dst.ChunkInfo().Channels
	for i := 0; i < n; i++ {
		var mean int64
		for ch := 0; ch < channels; ch++ {
			mean += src.At(i, ch).Int()
		}
		mean /= int64(channels)

		for ch := 0; ch < dstChannels; ch++ {
			dstSetter.Set(i, ch, wave.Int64Sample(mean))
		}
	}
	return nil
}
