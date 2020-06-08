package audio

import (
	"github.com/pion/mediadevices/pkg/wave"
	"github.com/pion/mediadevices/pkg/wave/mixer"
)

// NewChannelMixer creates audio transform to mix audio channels.
func NewChannelMixer(channels int, mixer mixer.ChannelMixer) TransformFunc {
	return func(r Reader) Reader {
		return ReaderFunc(func() (wave.Audio, error) {
			buff, err := r.Read()
			if err != nil {
				return nil, err
			}
			ci := buff.ChunkInfo()
			if ci.Channels == channels {
				return buff, nil
			}
			var mixed wave.Audio
			switch buff.(type) {
			case *wave.Int16Interleaved:
				mixed = wave.NewInt16Interleaved(
					wave.ChunkInfo{Channels: channels, Len: ci.Len},
				)
			case *wave.Int16NonInterleaved:
				mixed = wave.NewInt16NonInterleaved(
					wave.ChunkInfo{Channels: channels, Len: ci.Len},
				)
			case *wave.Float32Interleaved:
				mixed = wave.NewFloat32Interleaved(
					wave.ChunkInfo{Channels: channels, Len: ci.Len},
				)
			case *wave.Float32NonInterleaved:
				mixed = wave.NewFloat32NonInterleaved(
					wave.ChunkInfo{Channels: channels, Len: ci.Len},
				)
			}
			if err := mixer.Mix(mixed, buff); err != nil {
				return nil, err
			}
			return mixed, nil
		})
	}
}
