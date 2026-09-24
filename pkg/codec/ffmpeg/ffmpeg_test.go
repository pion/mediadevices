package ffmpeg

import (
	"context"
	"image"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/codec/internal/codectest"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

func TestEncoder(t *testing.T) {
	for name, factory := range map[string]func() (codec.VideoEncoderBuilder, error){
		"x264": func() (codec.VideoEncoderBuilder, error) {
			p, err := NewH264X264Params()
			p.FrameRate = 30
			p.BitRate = 1000000
			p.KeyFrameInterval = 60
			return &p, err
		},
	} {
		factory := factory
		t.Run(name, func(t *testing.T) {
			t.Run("SimpleRead", func(t *testing.T) {
				p, err := factory()
				if err != nil {
					t.Fatal(err)
				}
				codectest.VideoEncoderSimpleReadTest(t, p,
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
			t.Run("ReadAfterClose", func(t *testing.T) {
				p, err := factory()
				if err != nil {
					t.Fatal(err)
				}
				codectest.VideoEncoderReadAfterCloseTest(t, p,
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
		})
	}
}

func TestImageSizeChange(t *testing.T) {
	t.Skip("Changing image size on the fly is currently not supported")

	for name, factory := range map[string]func() (codec.VideoEncoderBuilder, error){
		"x264": func() (codec.VideoEncoderBuilder, error) {
			p, err := NewH264X264Params()
			p.FrameRate = 30
			p.BitRate = 1000000
			p.KeyFrameInterval = 60
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
		"x264": func() (codec.VideoEncoderBuilder, error) {
			p, err := NewH264X264Params()
			p.FrameRate = 30
			p.BitRate = 1000000
			p.KeyFrameInterval = 60
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
			r.Controller().(codec.KeyFrameController).ForceKeyFrame()
			_, rel, err = r.Read()
			if err != nil {
				t.Fatal(err)
			}
			// TODO: check if this is a key frame
			// if !r.(*encoder).isKeyFrame {
			// 	t.Fatal("Not a key frame")
			// }
			rel()
			_, _, err = r.Read()
			if err != io.EOF {
				t.Fatal(err)
			}
		})

	}
}

func TestSetBitrate(t *testing.T) {
	for name, factory := range map[string]func() (codec.VideoEncoderBuilder, error){
		"x264": func() (codec.VideoEncoderBuilder, error) {
			p, err := NewH264X264Params()
			p.FrameRate = 30
			p.BitRate = 1000000
			p.KeyFrameInterval = 60
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
			err = r.Controller().(codec.BitRateController).SetBitRate(1000) // 1000 bit/second is ridiculously low, but this is a testcase.
			if err != nil {
				t.Fatal(err)
			}
			_, rel, err = r.Read()
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
}

func TestShouldImplementBitRateControl(t *testing.T) {
	e := &softwareEncoder{}
	if _, ok := e.Controller().(codec.BitRateController); !ok {
		t.Error()
	}
}

func TestShouldImplementKeyFrameControl(t *testing.T) {
	e := &softwareEncoder{}
	if _, ok := e.Controller().(codec.KeyFrameController); !ok {
		t.Error()
	}
}

func TestEncoderFrameMonotonic(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	params, err := NewH264X264Params()
	params.FrameRate = 30
	params.BitRate = 1000000
	params.KeyFrameInterval = 60
	if err != nil {
		t.Fatal(err)
	}

	encoder, err := params.BuildVideoEncoder(
		video.ReaderFunc(func() (image.Image, func(), error) {
			return image.NewYCbCr(
				image.Rect(0, 0, 320, 240),
				image.YCbCrSubsampleRatio420,
			), func() {}, nil
		},
		), prop.Media{
			Video: prop.Video{
				Width:       320,
				Height:      240,
				FrameRate:   30,
				FrameFormat: frame.FormatI420,
			},
		})
	if err != nil {
		t.Fatal(err)
	}

	ticker := time.NewTicker(33 * time.Millisecond)
	defer ticker.Stop()
	ctxx, cancell := context.WithCancel(ctx)
	defer cancell()
	for {
		select {
		case <-ctxx.Done():
			return
		case <-ticker.C:
			_, rel, err := encoder.Read()
			if err != nil {
				t.Fatal(err)
			}
			rel()
		}
	}
}
