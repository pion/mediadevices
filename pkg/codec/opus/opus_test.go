package opus

import (
	"testing"

	"github.com/pion/mediadevices/pkg/codec/internal/codectest"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/mediadevices/pkg/wave"
)

func TestEncoder(t *testing.T) {
	t.Run("SimpleRead", func(t *testing.T) {
		p, err := NewParams()
		if err != nil {
			t.Fatal(err)
		}
		codectest.AudioEncoderSimpleReadTest(t, &p,
			prop.Media{
				Audio: prop.Audio{
					SampleRate:   48000,
					ChannelCount: 2,
				},
			},
			wave.NewInt16Interleaved(wave.ChunkInfo{
				Len:          960,
				SamplingRate: 48000,
				Channels:     2,
			}),
		)
	})
	t.Run("CloseTwice", func(t *testing.T) {
		p, err := NewParams()
		if err != nil {
			t.Fatal(err)
		}
		codectest.AudioEncoderCloseTwiceTest(t, &p, prop.Media{
			Audio: prop.Audio{
				SampleRate:   48000,
				ChannelCount: 2,
			},
		})
	})
}
