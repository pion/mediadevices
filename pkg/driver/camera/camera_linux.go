package camera

// #include <linux/videodev2.h>
import "C"

import (
	"context"
	"errors"
	"image"
	"io"
	"io/ioutil"
	"sync"

	"github.com/blackjack/webcam"
	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

const (
	maxEmptyFrameCount = 5
)

var (
	errReadTimeout = errors.New("read timeout")
	errEmptyFrame  = errors.New("empty frame")
)

// Camera implementation using v4l2
// Reference: https://linuxtv.org/downloads/v4l-dvb-apis/uapi/v4l/videodev.html#videodev
type camera struct {
	path            string
	cam             *webcam.Webcam
	formats         map[webcam.PixelFormat]frame.Format
	reversedFormats map[frame.Format]webcam.PixelFormat
	started         bool
	mutex           sync.Mutex
	cancel          func()
}

func init() {
	searchPath := "/dev/v4l/by-path/"
	devices, err := ioutil.ReadDir(searchPath)
	if err != nil {
		// No v4l device.
		return
	}
	for _, device := range devices {
		cam := newCamera(searchPath + device.Name())
		driver.GetManager().Register(cam, driver.Info{
			Label:      device.Name(),
			DeviceType: driver.Camera,
		})
	}
}

func newCamera(path string) *camera {
	formats := map[webcam.PixelFormat]frame.Format{
		webcam.PixelFormat(C.V4L2_PIX_FMT_YUYV):  frame.FormatYUYV,
		webcam.PixelFormat(C.V4L2_PIX_FMT_NV12):  frame.FormatNV21,
		webcam.PixelFormat(C.V4L2_PIX_FMT_MJPEG): frame.FormatMJPEG,
	}

	reversedFormats := make(map[frame.Format]webcam.PixelFormat)
	for k, v := range formats {
		reversedFormats[v] = k
	}

	c := &camera{
		path:            path,
		formats:         formats,
		reversedFormats: reversedFormats,
	}
	return c
}

func (c *camera) Open() error {
	cam, err := webcam.Open(c.path)
	if err != nil {
		return err
	}

	c.cam = cam
	return nil
}

func (c *camera) Close() error {
	if c.cam == nil {
		return nil
	}

	if c.cancel != nil {
		// Let the reader knows that the caller has closed the camera
		c.cancel()
		// Wait until the reader unref the buffer
		c.mutex.Lock()
		defer c.mutex.Unlock()

		// Note: StopStreaming frees frame buffers even if they are still used in Go code.
		//       There is currently no convenient way to do this safely.
		//       So, consumer of this stream must close camera after unusing all images.
		c.cam.StopStreaming()
		c.cancel = nil
	}
	c.cam.Close()
	return nil
}

func (c *camera) VideoRecord(p prop.Media) (video.Reader, error) {
	decoder, err := frame.NewDecoder(p.FrameFormat)
	if err != nil {
		return nil, err
	}

	pf := c.reversedFormats[p.FrameFormat]
	_, _, _, err = c.cam.SetImageFormat(pf, uint32(p.Width), uint32(p.Height))
	if err != nil {
		return nil, err
	}

	if err := c.cam.StartStreaming(); err != nil {
		return nil, err
	}

	cam := c.cam

	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	var buf []byte
	r := video.ReaderFunc(func() (img image.Image, err error) {
		// Lock to avoid accessing the buffer after StopStreaming()
		c.mutex.Lock()
		defer c.mutex.Unlock()

		// Wait until a frame is ready
		for i := 0; i < maxEmptyFrameCount; i++ {
			if ctx.Err() != nil {
				// Return EOF if the camera is already closed.
				return nil, io.EOF
			}

			err := cam.WaitForFrame(5) // 5 seconds
			switch err.(type) {
			case nil:
			case *webcam.Timeout:
				return nil, errReadTimeout
			default:
				// Camera has been stopped.
				return nil, err
			}

			b, err := cam.ReadFrame()
			if err != nil {
				// Camera has been stopped.
				return nil, err
			}

			// Frame is empty.
			// Retry reading and return errEmptyFrame if it exceeds maxEmptyFrameCount.
			if len(b) == 0 {
				continue
			}

			if len(b) > len(buf) {
				// Grow the intermediate buffer
				buf = make([]byte, len(b))
			}

			// move the memory from mmap to Go. This will guarantee that any data that's going out
			// from this reader will be Go safe. Otherwise, it's possible that outside of this reader
			// that this memory is still being used even after we close it.
			n := copy(buf, b)
			return decoder.Decode(buf[:n], p.Width, p.Height)
		}
		return nil, errEmptyFrame
	})

	return r, nil
}

func (c *camera) Properties() []prop.Media {
	properties := make([]prop.Media, 0)
	for format := range c.cam.GetSupportedFormats() {
		for _, frameSize := range c.cam.GetSupportedFrameSizes(format) {
			properties = append(properties, prop.Media{
				Video: prop.Video{
					Width:       int(frameSize.MaxWidth),
					Height:      int(frameSize.MaxHeight),
					FrameFormat: c.formats[format],
				},
			})
		}
	}
	return properties
}
