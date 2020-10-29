package screen

import (
	"fmt"
	"image"
	"time"

	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

type screen struct {
	num    int
	reader *reader
	tick   *time.Ticker
}

func deviceID(num int) string {
	return fmt.Sprintf("X11Screen%d", num)
}

func init() {
	dp, err := openDisplay()
	if err != nil {
		// No x11 display available.
		return
	}
	defer dp.Close()
	numScreen := dp.NumScreen()
	for i := 0; i < numScreen; i++ {
		driver.GetManager().Register(
			&screen{
				num: i,
			},
			driver.Info{
				Label:      deviceID(i),
				DeviceType: driver.Screen,
			},
		)
	}
}

func (s *screen) Open() error {
	r, err := newReader(s.num)
	if err != nil {
		return err
	}
	s.reader = r
	return nil
}

func (s *screen) Close() error {
	s.reader.Close()
	if s.tick != nil {
		s.tick.Stop()
	}
	return nil
}

func (s *screen) VideoRecord(p prop.Media) (video.Reader, error) {
	if p.FrameRate == 0 {
		p.FrameRate = 10
	}
	s.tick = time.NewTicker(time.Duration(float32(time.Second) / p.FrameRate))

	var dst image.RGBA
	reader := s.reader

	r := video.ReaderFunc(func() (image.Image, func(), error) {
		<-s.tick.C
		return reader.Read().ToRGBA(&dst), func() {}, nil
	})
	return r, nil
}

func (s *screen) Properties() []prop.Media {
	rect := s.reader.img.Bounds()
	w := rect.Dx()
	h := rect.Dy()
	return []prop.Media{
		{
			DeviceID: deviceID(s.num),
			Video: prop.Video{
				Width:       w,
				Height:      h,
				FrameFormat: frame.FormatRGBA,
			},
		},
	}
}
