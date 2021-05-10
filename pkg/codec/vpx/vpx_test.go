package vpx

import (
	"image"
	"io"
	"sync/atomic"
	"testing"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

func TestImageSizeChange(t *testing.T) {
	for name, factory := range map[string]func() (codec.VideoEncoderBuilder, error){
		"VP8": func() (codec.VideoEncoderBuilder, error) {
			p, err := NewVP8Params()
			return &p, err
		},
		"VP9": func() (codec.VideoEncoderBuilder, error) {
			p, err := NewVP9Params()
			// Disable latency to ease test and begin to receive packets for each input frame
			p.LagInFrames = 0
			return &p, err
		},
	} {
		factory := factory
		t.Run(name, func(t *testing.T) {
			param, err := factory()
			if err != nil {
				t.Fatal(err)
			}

			for name, testCase := range map[string]struct {
				initialWidth, initialHeight int
				width, height               int
			}{
				"NoChange": {
					320, 240,
					320, 240,
				},
				"Enlarge": {
					320, 240,
					640, 480,
				},
				"Shrink": {
					640, 480,
					320, 240,
				},
			} {
				testCase := testCase
				t.Run(name, func(t *testing.T) {
					var cnt uint32
					r, err := param.BuildVideoEncoder(
						video.ReaderFunc(func() (image.Image, func(), error) {
							i := atomic.AddUint32(&cnt, 1)
							if i == 1 {
								return image.NewYCbCr(
									image.Rect(0, 0, testCase.width, testCase.height),
									image.YCbCrSubsampleRatio420,
								), func() {}, nil
							}
							return nil, nil, io.EOF
						}),
						prop.Media{
							Video: prop.Video{
								Width:       testCase.initialWidth,
								Height:      testCase.initialHeight,
								FrameRate:   1,
								FrameFormat: frame.FormatI420,
							},
						},
					)
					if err != nil {
						t.Fatal(err)
					}
					_, rel, err := r.Read()
					if err != nil {
						t.Fatal(err)
					}
					rel()
					_, _, err = r.Read()
					if err != io.EOF {
						t.Fatal(err)
					}
				})
			}
		})
	}
}

func TestRequestKeyFrame(t *testing.T) {
	for name, factory := range map[string]func() (codec.VideoEncoderBuilder, error){
		"VP8": func() (codec.VideoEncoderBuilder, error) {
			p, err := NewVP8Params()
			return &p, err
		},
		"VP9": func() (codec.VideoEncoderBuilder, error) {
			p, err := NewVP9Params()
			// Disable latency to ease test and begin to receive packets for each input frame
			p.LagInFrames = 0
			return &p, err
		},
	} {
		factory := factory
		t.Run(name, func(t *testing.T) {
			param, err := factory()
			if err != nil {
				t.Fatal(err)
			}

			var initialWidth, initialHeight, width, height int = 320, 240, 320, 240

			var cnt uint32
			r, err := param.BuildVideoEncoder(
				video.ReaderFunc(func() (image.Image, func(), error) {
					i := atomic.AddUint32(&cnt, 1)
					if i == 3 {
						return nil, nil, io.EOF
					}
					return image.NewYCbCr(
						image.Rect(0, 0, width, height),
						image.YCbCrSubsampleRatio420,
					), func() {}, nil
				}),
				prop.Media{
					Video: prop.Video{
						Width:       initialWidth,
						Height:      initialHeight,
						FrameRate:   1,
						FrameFormat: frame.FormatI420,
					},
				},
			)
			if err != nil {
				t.Fatal(err)
			}
			_, rel, err := r.Read()
			if err != nil {
				t.Fatal(err)
			}
			rel()
			r.ForceKeyFrame()
			_, rel, err = r.Read()
			if err != nil {
				t.Fatal(err)
			}
			if !r.(*encoder).isKeyFrame {
				t.Fatal("Not a key frame")
			}
			rel()
			_, _, err = r.Read()
			if err != io.EOF {
				t.Fatal(err)
			}
		})

	}
}
