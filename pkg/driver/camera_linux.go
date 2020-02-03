package driver

// #include <linux/videodev2.h>
import "C"

import (
	"image"
	"io"

	"github.com/blackjack/webcam"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/io/video"
)

// Camera implementation using v4l2
// Reference: https://linuxtv.org/downloads/v4l-dvb-apis/uapi/v4l/videodev.html#videodev
type camera struct {
	path            string
	cam             *webcam.Webcam
	formats         map[webcam.PixelFormat]frame.Format
	reversedFormats map[frame.Format]webcam.PixelFormat
	properties      []video.AdvancedProperty
}

var _ VideoAdapter = &camera{}

func init() {
	// TODO: Probably try to get more cameras
	// Get default camera
	defaultCam := newCamera("/dev/video0")

	GetManager().Register(defaultCam)
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

	return &camera{
		path:            path,
		formats:         formats,
		reversedFormats: reversedFormats,
	}
}

func (c *camera) Open() error {
	cam, err := webcam.Open(c.path)
	if err != nil {
		return err
	}

	properties := make([]video.AdvancedProperty, 0)
	for format := range cam.GetSupportedFormats() {
		for _, frameSize := range cam.GetSupportedFrameSizes(format) {
			properties = append(properties, video.AdvancedProperty{
				Property: video.Property{
					Width:  int(frameSize.MaxWidth),
					Height: int(frameSize.MaxHeight),
				},
				FrameFormat: c.formats[format],
			})
		}
	}

	c.cam = cam
	c.properties = properties
	return nil
}

func (c *camera) Close() error {
	c.properties = nil
	if c.cam == nil {
		return nil
	}

	return c.cam.StopStreaming()
}

func (c *camera) Start(prop video.AdvancedProperty) (video.Reader, error) {
	decoder, err := frame.NewDecoder(prop.FrameFormat)
	if err != nil {
		return nil, err
	}

	pf := c.reversedFormats[prop.FrameFormat]
	_, _, _, err = c.cam.SetImageFormat(pf, uint32(prop.Width), uint32(prop.Height))
	if err != nil {
		return nil, err
	}

	if err := c.cam.StartStreaming(); err != nil {
		return nil, err
	}

	r := video.ReaderFunc(func() (img image.Image, err error) {
		// Wait until a frame is ready
		for {
			err := c.cam.WaitForFrame(5)
			switch err.(type) {
			case nil:
			case *webcam.Timeout:
				continue
			default:
				// Camera has been stopped.
				return nil, io.EOF
			}

			b, err := c.cam.ReadFrame()
			if err != nil {
				// Camera has been stopped.
				return nil, io.EOF
			}

			// Frame is not ready.
			if len(b) == 0 {
				continue
			}

			return decoder.Decode(b, prop.Width, prop.Height)
		}
	})

	return r, nil
}

func (c *camera) Stop() error {
	return c.cam.StopStreaming()
}

func (c *camera) Info() Info {
	return Info{
		DeviceType: Camera,
	}
}

func (c *camera) Properties() []video.AdvancedProperty {
	return c.properties
}
