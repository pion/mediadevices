package vpx

import (
	"context"
	"image"
	"io"
	"math"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/codec/internal/codectest"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/stretchr/testify/assert"
)

func TestEncoder(t *testing.T) {
	for name, factory := range map[string]func() (codec.VideoEncoderBuilder, error){
		"VP8": func() (codec.VideoEncoderBuilder, error) {
			p, err := NewVP8Params()
			return &p, err
		},
		"VP9": func() (codec.VideoEncoderBuilder, error) {
			p, err := NewVP9Params()
			p.LagInFrames = 0
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
			r.Controller().(codec.KeyFrameController).ForceKeyFrame()
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

func TestSetBitrate(t *testing.T) {
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
	e := &encoder{}
	if _, ok := e.Controller().(codec.BitRateController); !ok {
		t.Error()
	}
}

func TestShouldImplementKeyFrameControl(t *testing.T) {
	e := &encoder{}
	if _, ok := e.Controller().(codec.KeyFrameController); !ok {
		t.Error()
	}
}

func TestEncoderFrameMonotonic(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	params, err := NewVP8Params()
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

func TestVP8DynamicQPControl(t *testing.T) {
	t.Run("VP8", func(t *testing.T) {
		p, err := NewVP8Params()
		if err != nil {
			t.Fatal(err)
		}
		p.LagInFrames = 0 // Disable frame lag buffering for real-time encoding
		p.RateControlEndUsage = RateControlCBR
		totalFrames := 100
		frameRate := 10
		initialWidth, initialHeight := 800, 600
		var cnt uint32

		r, err := p.BuildVideoEncoder(
			video.ReaderFunc(func() (image.Image, func(), error) {
				i := atomic.AddUint32(&cnt, 1)
				if i == uint32(totalFrames+1) {
					return nil, nil, io.EOF
				}
				img := image.NewYCbCr(image.Rect(0, 0, initialWidth, initialHeight), image.YCbCrSubsampleRatio420)
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				for i := range img.Y {
					img.Y[i] = uint8(r.Intn(256))
				}
				for i := range img.Cb {
					img.Cb[i] = uint8(r.Intn(256))
				}
				for i := range img.Cr {
					img.Cr[i] = uint8(r.Intn(256))
				}
				return img, func() {}, nil
			}),
			prop.Media{
				Video: prop.Video{
					Width:       initialWidth,
					Height:      initialHeight,
					FrameRate:   float32(frameRate),
					FrameFormat: frame.FormatI420,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		initialBitrate := 100
		currentBitrate := initialBitrate
		targetBitrate := 300
		for i := 0; i < totalFrames; i++ {
			r.Controller().(codec.KeyFrameController).ForceKeyFrame()
			r.Controller().(codec.QPController).DynamicQPControl(currentBitrate, targetBitrate)
			data, rel, err := r.Read()
			if err != nil {
				t.Fatal(err)
			}
			rel()
			encodedSize := len(data)
			currentBitrate = encodedSize * 8 / 1000 / frameRate
		}
		assert.Less(t, math.Abs(float64(targetBitrate-currentBitrate)), math.Abs(float64(initialBitrate-currentBitrate)))
	})
}

func TestVP8EncodeDecode(t *testing.T) {
	t.Run("VP8", func(t *testing.T) {
		initialWidth, initialHeight := 800, 600
		reader, writer := io.Pipe()
		decoder, err := BuildVideoDecoder(reader, prop.Media{
			Video: prop.Video{
				Width:       initialWidth,
				Height:      initialHeight,
				FrameFormat: frame.FormatI420,
			},
		})
		if err != nil {
			t.Fatalf("Error creating VP8 decoder: %v", err)
		}
		defer decoder.Close()

		// [... encoder setup code ...]
		p, err := NewVP8Params()
		if err != nil {
			t.Fatal(err)
		}
		p.LagInFrames = 0 // Disable frame lag buffering for real-time encoding
		p.RateControlEndUsage = RateControlCBR
		totalFrames := 100
		var cnt uint32
		r, err := p.BuildVideoEncoder(
			video.ReaderFunc(func() (image.Image, func(), error) {
				i := atomic.AddUint32(&cnt, 1)
				if i == uint32(totalFrames+1) {
					return nil, nil, io.EOF
				}
				img := image.NewYCbCr(image.Rect(0, 0, initialWidth, initialHeight), image.YCbCrSubsampleRatio420)
				return img, func() {}, nil
			}),
			prop.Media{
				Video: prop.Video{
					Width:       initialWidth,
					Height:      initialHeight,
					FrameRate:   30,
					FrameFormat: frame.FormatI420,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		data, rel, err := r.Read()
		if err != nil {
			t.Fatal(err)
		}
		defer rel()

		// Decode the frame
		writer.Write(data)
		writer.Close()

		// Poll for frame with timeout
		timeout := time.After(2 * time.Second)
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				t.Fatal("Timeout: No frame received within 2 seconds")
			case <-ticker.C:
				frame, rel, err := decoder.Read()
				if err != nil {
					t.Fatal(err)
				}
				defer rel()
				if frame != nil {
					t.Log("Successfully received and decoded frame")
					return // Test passes
				}
			}
		}
	})
}
