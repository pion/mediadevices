package opus

import (
	"github.com/pion/mediadevices/pkg/codec/internal/codectest"
	"github.com/pion/mediadevices/pkg/prop"
	"testing"
)

func TestEncoder(t *testing.T) {
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
