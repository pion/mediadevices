package vpx

import (
	"context"
	"image"
	"io"
	"math"
	"math/rand"
	"strings"
	"sync"
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

// TestEncodeErrorIncludesDetail verifies that when vpx_codec_encode returns
// a non-OK code, the Go error message includes libvpx's err_detail string
// (set inside libvpx's ERROR(...) macro). This is the field that holds the
// specific reason for VPX_CODEC_INVALID_PARAM and similar errors.
//
// We trigger the failure deterministically by reaching into the encoder
// after the first successful encode and forcing a size mismatch between
// e.raw (the vpx_image_t the encoder hands to vpx_codec_encode) and the
// encoder cfg's expected dimensions. libvpx's validate_img() rejects this
// with "Image size must match encoder init configuration size" via the
// ERROR macro and returns VPX_CODEC_INVALID_PARAM. With this change, the
// Go error wraps that detail string so oncall/operator-side logs can see
// the specific cause instead of a bare error code.
func TestEncodeErrorIncludesDetail(t *testing.T) {
	p, err := NewVP8Params()
	if err != nil {
		t.Fatal(err)
	}

	r, err := p.BuildVideoEncoder(
		video.ReaderFunc(func() (image.Image, func(), error) {
			return image.NewYCbCr(image.Rect(0, 0, 320, 240), image.YCbCrSubsampleRatio420), func() {}, nil
		}),
		prop.Media{
			Video: prop.Video{
				Width:       320,
				Height:      240,
				FrameRate:   1,
				FrameFormat: frame.FormatI420,
			},
		})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	// First read succeeds and initializes libvpx with cfg.g_w/g_h = 320/240.
	if _, rel, err := r.Read(); err != nil {
		t.Fatalf("first read should succeed: %v", err)
	} else {
		rel()
	}

	// Force a size mismatch inside the encoder by mutating e.raw's display
	// dimensions to disagree with e.cfg. The Read() path doesn't touch d_w/d_h
	// unless image dimensions change (which they won't here), so the corrupt
	// e.raw flows directly into vpx_codec_encode and triggers validate_img.
	e := r.(*encoder)
	e.mu.Lock()
	e.raw.d_w = 100
	e.raw.d_h = 100
	e.mu.Unlock()

	_, _, err = r.Read()
	if err == nil {
		t.Skip("forced mismatch unexpectedly succeeded — libvpx tolerant on this build")
	}
	t.Logf("error: %v", err)
	if !strings.Contains(err.Error(), "vpx_codec_encode failed") {
		t.Fatalf("unexpected error shape: %v", err)
	}
	// The new error shape is "vpx_codec_encode failed (N): <detail>"; we
	// require the colon-detail suffix is present. The detail itself
	// (libvpx's err_detail) should be non-empty for this size-mismatch path.
	if !strings.Contains(err.Error(), "): ") {
		t.Fatalf("error missing detail suffix; expected `(N): <detail>` shape, got: %v", err)
	}
	suffix := err.Error()[strings.Index(err.Error(), "): ")+3:]
	if suffix == "" {
		t.Fatalf("err_detail is empty; expected libvpx to set it for the size-mismatch path")
	}
	t.Logf("extracted libvpx err_detail: %q", suffix)
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

// TestEncoderHandlesLongIdleGap verifies the duration clamp in the Read()
// path. When an encoder sits idle for a long time (no frames arriving from
// the source, so vpx_codec_encode is never called and tLastFrame stays at
// encoder-creation time), the next frame would otherwise be encoded with
// `duration = t.Sub(tLastFrame).Microseconds()` — a value potentially in
// the trillions of microseconds.
//
// libvpx 1.15.0+ explicitly rejects any duration > UINT32_MAX with
// VPX_CODEC_INVALID_PARAM at vpx/src/vpx_encoder.c:206 (added in libvpx
// commit 7fb8ceccf, 2024-03-14). UINT32_MAX μs is about 71m35s, so any
// encoder idle longer than that fails its next encode. Because the failure
// path inside Read() does not refresh tLastFrame, every subsequent frame
// sees an even larger duration and fails identically — the encoder becomes
// permanently unusable until recreated.
//
// The clamp in Read() substitutes a sane synthetic frame interval when the
// computed duration exceeds maxFrameDurationUs, keeping the encoder usable
// across idle gaps of arbitrary length.
func TestEncoderHandlesLongIdleGap(t *testing.T) {
	params, err := NewVP8Params()
	if err != nil {
		t.Fatal(err)
	}

	rcloser, err := params.BuildVideoEncoder(
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
	defer rcloser.Close()

	e, ok := rcloser.(*encoder)
	if !ok {
		t.Fatalf("expected *encoder, got %T", rcloser)
	}

	// Simulate an encoder that has sat idle since creation. tStart and
	// tLastFrame are both rewound, mirroring the production sequence where
	// the encoder waited hours for a remote subscriber before its first
	// frame arrived.
	e.mu.Lock()
	const idleDuration = 6 * time.Hour
	e.tStart = e.tStart.Add(-idleDuration)
	e.tLastFrame = e.tLastFrame.Add(-idleDuration)
	tLastFrameBefore := e.tLastFrame
	e.mu.Unlock()

	// First encode after the idle gap must succeed.
	if _, rel, err := rcloser.Read(); err != nil {
		t.Fatalf("encode after long idle gap failed: %v", err)
	} else {
		rel()
	}

	// tLastFrame must be refreshed so subsequent frames see a normal-sized
	// duration. The bug's permanence stemmed from tLastFrame staying stale
	// on the failure path.
	e.mu.Lock()
	tLastFrameAfter := e.tLastFrame
	e.mu.Unlock()
	if !tLastFrameAfter.After(tLastFrameBefore) {
		t.Fatalf("tLastFrame not refreshed after encode (before=%v after=%v)",
			tLastFrameBefore, tLastFrameAfter)
	}

	// Subsequent encodes must continue working.
	for i := 0; i < 3; i++ {
		if _, rel, err := rcloser.Read(); err != nil {
			t.Fatalf("encode %d after recovery failed: %v", i, err)
		} else {
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
		totalFrames := 10
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

		var wg sync.WaitGroup
		wg.Add(1)

		counter := 0
		go func() {
			defer wg.Done()
			for {
				img, rel, err := decoder.Read()
				if err == io.EOF {
					return
				}
				if err != nil {
					t.Errorf("decoder read error: %v", err)
					return
				}
				assert.Equal(t, initialWidth, img.Bounds().Dx())
				assert.Equal(t, initialHeight, img.Bounds().Dy())
				rel()
				counter++
			}
		}()

		// --- feed encoded frames to writer
		for {
			data, rel, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("encoder error: %v", err)
			}
			_, werr := writer.Write(data)
			rel()
			if werr != nil {
				t.Fatalf("writer error: %v", werr)
			}
		}
		writer.Close()
		wg.Wait()
		assert.Equal(t, totalFrames, counter)
	})
}
