package screen

import (
	"fmt"
	"image"
	"image/color"
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
				Label: deviceID(i),
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
	rect := s.reader.img.Bounds()
	w := rect.Max.X - rect.Min.X
	h := rect.Max.Y - rect.Min.Y
	imgI444 := image.NewYCbCr(rect, image.YCbCrSubsampleRatio444)

	if p.FrameRate == 0 {
		p.FrameRate = 10
	}
	s.tick = time.NewTicker(time.Duration(float32(time.Second) / p.FrameRate))

	r := video.ReaderFunc(func() (image.Image, error) {
		<-s.tick.C
		img := s.reader.Read()
		// Convert it to I444
		for y := 0; y < h; y++ {
			iyBase := y * imgI444.YStride
			icBase := y * imgI444.CStride
			for x := 0; x < w; x++ {
				iy := iyBase + x
				ic := icBase + x
				r, g, b, _ := img.At(x, y).RGBA()
				yy, cb, cr := color.RGBToYCbCr(uint8(r/0x100), uint8(g/0x100), uint8(b/0x100))
				imgI444.Y[iy] = yy
				imgI444.Cb[ic] = cb
				imgI444.Cr[ic] = cr
			}
		}
		return imgI444, nil
	})
	return r, nil
}

func (s *screen) Properties() []prop.Media {
	rect := s.reader.img.Bounds()
	w := rect.Max.X - rect.Min.X
	h := rect.Max.Y - rect.Min.Y
	return []prop.Media{
		{
			DeviceID: deviceID(s.num),
			Video: prop.Video{
				Width:       w,
				Height:      h,
				FrameFormat: frame.FormatI444,
			},
		},
	}
}
