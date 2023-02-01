package x264

import (
	"image"
	"testing"

	"github.com/pion/mediadevices/pkg/codec"
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
		p.BitRate = 200000
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
		p.BitRate = 200000
		codectest.VideoEncoderCloseTwiceTest(t, &p, prop.Media{
			Video: prop.Video{
				Width:       640,
				Height:      480,
				FrameRate:   30,
				FrameFormat: frame.FormatI420,
			},
		})
	})
	t.Run("ReadAfterClose", func(t *testing.T) {
		p, err := NewParams()
		if err != nil {
			t.Fatal(err)
		}
		p.BitRate = 200000
		codectest.VideoEncoderReadAfterCloseTest(t, &p,
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
}

func TestShouldImplementKeyFrameControl(t *testing.T) {
	t.SkipNow() // TODO: Implement key frame control

	e := &encoder{}
	if _, ok := e.Controller().(codec.KeyFrameController); !ok {
		t.Error()
	}
}

func TestShouldImplementBitRateControl(t *testing.T) {
	t.SkipNow() // TODO: Implement bit rate control

	e := &encoder{}
	if _, ok := e.Controller().(codec.BitRateController); !ok {
		t.Error()
	}
}
