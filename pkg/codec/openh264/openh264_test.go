package openh264

import (
	"image"
	"testing"

	"github.com/pion/mediadevices/pkg/codec/internal/codectest"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/prop"
)

func TestEncoder(t *testing.T) {
	t.Run("SimpleRead", func(t *testing.T) {
		p, err := NewParams()
		if err != nil {
			t.Fatal(err)
		}
		codectest.VideoEncoderSimpleReadTest(t, &p,
			prop.Media{
				Video: prop.Video{
					Width:       256,
					Height:      144,
					FrameFormat: frame.FormatI420,
				},
			},
			image.NewYCbCr(
				image.Rect(0, 0, 256, 144),
				image.YCbCrSubsampleRatio420,
			),
		)
	})
	t.Run("CloseTwice", func(t *testing.T) {
		p, err := NewParams()
		if err != nil {
			t.Fatal(err)
		}
		codectest.VideoEncoderCloseTwiceTest(t, &p, prop.Media{
			Video: prop.Video{
				Width:       640,
				Height:      480,
				FrameRate:   30,
				FrameFormat: frame.FormatI420,
			},
		})
	})
}
