// Package videotest provides vncDevice video driver for testing.
package vnc

import (
	"context"
	"fmt"
	"github.com/mitchellh/go-vnc"
	"image"
	"image/color"
	"io"
	"net"
	"time"

	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

func init() {
	driver.GetManager().Register(
		newVnc(),
		driver.Info{Label: "VNCVideo", DeviceType: driver.Camera},
	)
}

type vncDevice struct {
	closed <-chan struct{}
	cancel func()
	tick   *time.Ticker
	h,w int
	yyBase []byte
	cbBase []byte
	crBase []byte
	//mutex sync.Mutex
}

func newVnc() *vncDevice {
	return &vncDevice{}
}

func (d *vncDevice) Open() error {
	ctx, cancel := context.WithCancel(context.Background())
	//ctx.Value()
	d.closed = ctx.Done()
	d.cancel = cancel
	conn, _ := net.Dial("tcp", "10.190.50.76:5900")
	c := make(chan vnc.ServerMessage, 1)
	conf := vnc.ClientConfig{ServerMessageCh: c}
	client, _ := vnc.Client(conn, &conf)
	d.h=int(client.FrameBufferHeight)
	d.w=int(client.FrameBufferWidth)
	d.yyBase = make([]byte, d.h*d.w)
	d.cbBase = make([]byte, d.h*d.w)
	d.crBase = make([]byte, d.h*d.w)
	go func() {
		fmt.Println("Begin FramebufferUpdate")
		client.FramebufferUpdateRequest(true, 0, 0, client.FrameBufferWidth, client.FrameBufferHeight)
		for {
			select {
			case msg := <-c:
				switch t := msg.(type) {
				case *vnc.FramebufferUpdateMessage:
					for _, rect := range t.Rectangles {
						raw := rect.Enc.(*vnc.RawEncoding)
						fmt.Printf("%d\t%d\t%d\t%d\t%d\r\n",rect.X,rect.Y,rect.Width,rect.Width,len(raw.Colors))
						for y:=int(rect.Y);y<int(rect.Height+rect.Y);y++ {
							for x:=int(rect.X);x<int(rect.Width+rect.X);x++ {
								rgb:=raw.Colors[(y-int(rect.Y))*int(rect.Width)+(x-int(rect.X))]
								yy,cb,cr:=color.RGBToYCbCr(uint8(rgb.R),uint8(rgb.G),uint8(rgb.B))
								pos:=y*d.w+x
								d.yyBase[pos]=yy
								d.cbBase[pos]=cb
								d.crBase[pos]=cr
							}
						}

					}
					time.Sleep(33 * time.Millisecond)
					client.FramebufferUpdateRequest(true, 0, 0, client.FrameBufferWidth, client.FrameBufferHeight)
					break
				default:

				}
			case <-time.After(5 * time.Second):
				fmt.Println("Timeout FramebufferUpdate")
				client.FramebufferUpdateRequest(true, 0, 0, client.FrameBufferWidth, client.FrameBufferHeight)

			}
		}
	}()
	return nil
}

func (d *vncDevice) Close() error {
	d.cancel()
	if d.tick != nil {
		d.tick.Stop()
	}
	return nil
}

func (d *vncDevice) VideoRecord(p prop.Media) (video.Reader, error) {
	if p.FrameRate == 0 {
		p.FrameRate = 30
	}

	tick := time.NewTicker(time.Duration(float32(time.Second) / p.FrameRate))
	d.tick = tick
	closed := d.closed
	yy := make([]byte, d.h*d.w)
	cb := make([]byte, d.h*d.w)
	cr := make([]byte, d.h*d.w)
	r := video.ReaderFunc(func() (image.Image, func(), error) {
		select {
		case <-closed:
			return nil, func() {}, io.EOF
		default:
		}

		<-tick.C
		copy(yy, d.yyBase)
		copy(cb, d.cbBase)
		copy(cr, d.crBase)
		return &image.YCbCr{
			Y:              yy,
			YStride:        d.w,
			Cb:             cb,
			Cr:             cr,
			CStride:        d.w ,
			SubsampleRatio: image.YCbCrSubsampleRatio444,
			Rect:           image.Rect(0, 0, d.w, d.h),
		}, func() {}, nil
	})

	return r, nil
}

func (d vncDevice) Properties() []prop.Media {
	return []prop.Media{
		{
			Video: prop.Video{
				Width:       d.w,
				Height:      d.h,
				FrameFormat: frame.FormatI444,
			},
		},
	}
}
