package vaapi

import (
	"errors"
	"os"
	"testing"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/codec/internal/codectest"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/prop"
)

func TestEncoder(t *testing.T) {
	if _, err := os.Stat("/dev/dri/card0"); errors.Is(err, os.ErrNotExist) {
		t.Skip("/dev/dri/card0 not found")
	}

	for name, factory := range map[string]func() (codec.VideoEncoderBuilder, error){
		"VP8": func() (codec.VideoEncoderBuilder, error) {
			p, err := NewVP8Params()
			return &p, err
		},
		"VP9": func() (codec.VideoEncoderBuilder, error) {
			p, err := NewVP9Params()
			return &p, err
		},
	} {
		factory := factory
		t.Run(name, func(t *testing.T) {
			t.Run("CloseTwice", func(t *testing.T) {
				p, err := factory()
				if err != nil {
					t.Fatal(err)
				}
				codectest.VideoEncoderCloseTwiceTest(t, p, prop.Media{
					Video: prop.Video{
						Width:       640,
						Height:      480,
						FrameRate:   30,
						FrameFormat: frame.FormatI420,
					},
				})
			})
		})
	}
}
