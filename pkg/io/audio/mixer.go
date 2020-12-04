package audio

import (
	"github.com/pion/mediadevices/pkg/wave"
	"github.com/pion/mediadevices/pkg/wave/mixer"
)

// NewChannelMixer creates audio transform to mix audio channels.
func NewChannelMixer(channels int, mixer mixer.ChannelMixer) TransformFunc {
	return func(r Reader) Reader {
		return ReaderFunc(func() (wave.Audio, func(), error) {
			buff, _, err := r.Read()
			if err != nil {
				return nil, func() {}, err
			}
			ci := buff.ChunkInfo()
			if ci.Channels == channels {
				return buff, func() {}, nil
			}

			ci.Channels = channels

			var mixed wave.Audio
			switch buff.(type) {
			case *wave.Int16Interleaved:
				mixed = wave.NewInt16Interleaved(ci)
			case *wave.Int16NonInterleaved:
				mixed = wave.NewInt16NonInterleaved(ci)
			case *wave.Float32Interleaved:
				mixed = wave.NewFloat32Interleaved(ci)
			case *wave.Float32NonInterleaved:
				mixed = wave.NewFloat32NonInterleaved(ci)
			}
			if err := mixer.Mix(mixed, buff); err != nil {
				return nil, func() {}, err
			}
			return mixed, func() {}, nil
		})
	}
}
