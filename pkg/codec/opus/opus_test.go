package opus

import (
	"testing"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/codec/internal/codectest"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/mediadevices/pkg/wave"
)

func TestShouldImplementBitRateControl(t *testing.T) {
	e := &encoder{}
	if _, ok := e.Controller().(codec.BitRateController); !ok {
		t.Error()
	}
}

func TestShouldImplementKeyFrameControl(t *testing.T) {
	t.SkipNow() // TODO: Implement key frame control

	e := &encoder{}
	if _, ok := e.Controller().(codec.KeyFrameController); !ok {
		t.Error()
	}
}

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
	t.Run("ReadAfterClose", func(t *testing.T) {
		p, err := NewParams()
		if err != nil {
			t.Fatal(err)
		}
		codectest.AudioEncoderReadAfterCloseTest(t, &p,
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
}
