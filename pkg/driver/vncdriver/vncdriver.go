// Package videotest provides vncDevice video driver for testing.
package vncdriver

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/pion/mediadevices/pkg/driver/vncdriver/vnc"
	"image"
	"io"
	"net"
	"sync"
	"time"

	//"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

type vncDevice struct {
	closed   <-chan struct{}
	cancel   func()
	tick     *time.Ticker
	h, w     int
	rawPixel []byte
	mutex    sync.Mutex
	vClient  *vnc.ClientConn
	vncAddr  string
}

func NewVnc(vncAddr string) *vncDevice {
	return &vncDevice{vncAddr: vncAddr}
}
func (d *vncDevice) PointerEvent(mask uint8, x, y uint16) {
	d.vClient.PointerEvent(vnc.ButtonMask(mask), x, y)
}
func (d *vncDevice) KeyEvent(keysym uint32, down bool) {
	d.vClient.KeyEvent(keysym, down)
}
func (d *vncDevice) Open() error {
	ctx, cancel := context.WithCancel(context.Background())
	d.closed = ctx.Done()
	d.cancel = cancel
	if d.vClient != nil {
		return nil
	}
	c := make(chan vnc.ServerMessage, 1)
	conf := vnc.ClientConfig{
		ServerMessageCh: c,
		Exclusive:       false,
	}
	d.mutex.Lock()
	defer d.mutex.Unlock()
	conn, err := net.Dial("tcp", d.vncAddr)
	if err != nil {
		return err
	}
	d.vClient, err = vnc.Client(conn, &conf)
	if err != nil {
		return err
	}
	d.vClient.SetEncodings([]vnc.Encoding{
		&vnc.ZlibEncoding{},
		&vnc.RawEncoding{},
	})
	d.w = int(d.vClient.FrameBufferWidth)
	d.h = int(d.vClient.FrameBufferHeight)

	d.rawPixel = make([]byte, d.h*d.w*4)

	capCtx, _ := context.WithCancel(ctx)
	go func(ctx context.Context) {
		fmt.Println("Begin FramebufferUpdate")
		if d.vClient == nil {
			return
		}
		d.vClient.FramebufferUpdateRequest(true, 0, 0, uint16(d.w), uint16(d.h))
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Stop Record Video By FBUpdate")
				return
			case msg := <-c:
				switch t := msg.(type) {
				case *vnc.FramebufferUpdateMessage:
					for _, rect := range t.Rectangles {
						var pix []uint32
						switch t := rect.Enc.(type) {
						case *vnc.RawEncoding:
							pix = t.RawPixel
						case *vnc.ZlibEncoding:
							pix = t.RawPixel
						}
						for y := int(rect.Y); y < int(rect.Height+rect.Y); y++ {
							for x := int(rect.X); x < int(rect.Width+rect.X); x++ {
								binary.LittleEndian.PutUint32(d.rawPixel[(y*d.w+x)*4:], pix[(y-int(rect.Y))*int(rect.Width)+(x-int(rect.X))])
								//BigEndian
							}
						}

					}
					//time.Sleep(33 * time.Millisecond)
					d.vClient.FramebufferUpdateRequest(true, 0, 0, uint16(d.w), uint16(d.h))
					break
				default:

				}
			case <-time.After(10 * time.Second):
				//fmt.Println("Timeout FramebufferUpdate")
				if d.vClient.FramebufferUpdateRequest(true, 0, 0, uint16(d.w), uint16(d.h)) != nil {
					d.cancel()
					return
				}

			}
		}
	}(capCtx)
	return nil
}

func (d *vncDevice) Close() error {
	d.cancel()
	if d.tick != nil {
		d.tick.Stop()
	}
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if d.vClient != nil {
		d.vClient.Close()
		d.vClient = nil
	}
	return nil
}

func (d *vncDevice) VideoRecord(p prop.Media) (video.Reader, error) {
	if p.FrameRate == 0 {
		p.FrameRate = 15
	}

	tick := time.NewTicker(time.Duration(float32(time.Second) / p.FrameRate))
	d.tick = tick
	closed := d.closed
	pixs := make([]byte, d.h*d.w*4)
	r := video.ReaderFunc(func() (image.Image, func(), error) {
		select {
		case <-closed:
			fmt.Println("Stop Record Video By VideoRecord")
			return nil, func() {}, io.EOF
		default:
		}

		<-tick.C
		copy(pixs, d.rawPixel)
		return &image.RGBA{
			Pix:    pixs,
			Stride: 4,
			Rect:   image.Rect(0, 0, d.w, d.h),
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
