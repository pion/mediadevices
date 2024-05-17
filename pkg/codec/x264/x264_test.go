package x264

import (
	"image"
	"testing"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/codec/internal/codectest"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

func getTestVideoEncoder() (codec.ReadCloser, error) {
	p, err := NewParams()
	if err != nil {
		return nil, err
	}
	p.BitRate = 200000
	enc, err := p.BuildVideoEncoder(video.ReaderFunc(func() (image.Image, func(), error) {
		return image.NewYCbCr(
			image.Rect(0, 0, 256, 144),
			image.YCbCrSubsampleRatio420,
		), nil, nil
	}), prop.Media{
		Video: prop.Video{
			Width:       256,
			Height:      144,
			FrameFormat: frame.FormatI420,
		},
	})
	if err != nil {
		return nil, err
	}
	return enc, nil
}

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
	e := &encoder{}
	if _, ok := e.Controller().(codec.KeyFrameController); !ok {
		t.Error()
	}
}

func TestNoErrorOnForceKeyFrame(t *testing.T) {
	enc, err := getTestVideoEncoder()
	if err != nil {
		t.Error(err)
	}
	kfc, ok := enc.Controller().(codec.KeyFrameController)
	if !ok {
		t.Error()
	}
	if err := kfc.ForceKeyFrame(); err != nil {
		t.Error(err)
	}
	_, rel, err := enc.Read() // try to read the encoded frame
	rel()
	if err != nil {
		t.Fatal(err)
	}
}

func TestShouldImplementBitRateControl(t *testing.T) {
	e := &encoder{}
	if _, ok := e.Controller().(codec.BitRateController); !ok {
		t.Error()
	}
}

func TestNoErrorOnSetBitRate(t *testing.T) {
	enc, err := getTestVideoEncoder()
	if err != nil {
		t.Error(err)
	}
	brc, ok := enc.Controller().(codec.BitRateController)
	if !ok {
		t.Error()
	}
	if err := brc.SetBitRate(1000); err != nil { // 1000 bit/second is ridiculously low, but this is a testcase.
		t.Error(err)
	}
	_, rel, err := enc.Read() // try to read the encoded frame
	rel()
	if err != nil {
		t.Fatal(err)
	}
}
