// Package videotest provides dummy video driver for testing.
package videotest

import (
	"context"
	"image"
	"io"
	"math/rand"
	"time"

	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

func init() {
	driver.GetManager().Register(
		newVideoTest(),
		driver.Info{Label: "VideoTest", DeviceType: driver.Camera},
	)
}

type dummy struct {
	closed <-chan struct{}
	cancel func()
	tick   *time.Ticker
}

func newVideoTest() *dummy {
	return &dummy{}
}

func (d *dummy) Open() error {
	ctx, cancel := context.WithCancel(context.Background())
	d.closed = ctx.Done()
	d.cancel = cancel
	return nil
}

func (d *dummy) Close() error {
	d.cancel()
	if d.tick != nil {
		d.tick.Stop()
	}
	return nil
}

func (d *dummy) VideoRecord(p prop.Media) (video.Reader, error) {
	yi := p.Width * p.Height
	ci := yi / 2

	yy := make([]byte, yi)
	cb := make([]byte, ci)
	cr := make([]byte, ci)
	colors := [][3]byte{
		{235, 128, 128},
		{210, 16, 146},
		{170, 166, 16},
		{145, 54, 34},
		{107, 202, 222},
		{82, 90, 240},
		{41, 240, 110},
	}

	if p.FrameRate == 0 {
		p.FrameRate = 30
	}
	d.tick = time.NewTicker(time.Duration(float32(time.Second) / p.FrameRate))

	r := video.ReaderFunc(func() (image.Image, error) {
		select {
		case <-d.closed:
			return nil, io.EOF
		default:
		}

		<-d.tick.C

		for y := 0; y < p.Height; y++ {
			yi := p.Width * y
			ci := p.Width * y / 2
			if y > p.Height*3/4 {
				for x := 0; x < p.Width; x++ {
					c := x * 7 / p.Width
					if c > 4 {
						// Noise
						yy[yi+x] = uint8(rand.Int31n(2) * 255)
						cb[ci+x/2] = 128
						cr[ci+x/2] = 128
					} else {
						// Gray
						yy[yi+x] = uint8(x * 255 * 7 / (5 * p.Width))
						cb[ci+x/2] = 128
						cr[ci+x/2] = 128
					}
				}
			} else {
				// Color bar
				for x := 0; x < p.Width; x++ {
					c := x * 7 / p.Width
					yy[yi+x] = uint8(uint16(colors[c][0]) * 75 / 100)
					cb[ci+x/2] = colors[c][1]
					cr[ci+x/2] = colors[c][2]
				}
			}
		}
		return &image.YCbCr{
			Y:              yy,
			YStride:        p.Width,
			Cb:             cb,
			Cr:             cr,
			CStride:        p.Width / 2,
			SubsampleRatio: image.YCbCrSubsampleRatio422,
			Rect:           image.Rect(0, 0, p.Width, p.Height),
		}, nil
	})

	return r, nil
}

func (d dummy) Properties() []prop.Media {
	return []prop.Media{
		{
			Video: prop.Video{
				Width:       640,
				Height:      480,
				FrameFormat: frame.FormatYUYV,
			},
		},
	}
}
