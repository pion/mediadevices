package driver

// #include <linux/videodev2.h>
import "C"

import (
	"github.com/blackjack/webcam"
	"github.com/pion/mediadevices/frame"
)

// Camera implementation using v4l2
// Reference: https://linuxtv.org/downloads/v4l-dvb-apis/uapi/v4l/videodev.html#videodev
type camera struct {
	path    string
	cam     *webcam.Webcam
	formats map[webcam.PixelFormat]frame.Format
}

var _ VideoAdapter = &camera{}

func init() {
	// TODO: Probably try to get more cameras
	// Get default camera
	defaultCam := newCamera("/dev/video0")

	Manager.register(defaultCam)
}

func newCamera(path string) *camera {
	return &camera{
		path: path,
		formats: map[webcam.PixelFormat]frame.Format{
			webcam.PixelFormat(C.V4L2_PIX_FMT_YUYV): frame.FormatYUYV,
			webcam.PixelFormat(C.V4L2_PIX_FMT_NV12): frame.FormatNV21,
			webcam.PixelFormat(C.V4L2_PIX_FMT_MJPEG): frame.FormatMJPEG,
		},
	}
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

	return c.cam.StopStreaming()
}

func (c *camera) Start(spec VideoSpec, cb DataCb) error {
	if err := c.cam.StartStreaming(); err != nil {
		return err
	}

	for {
		err := c.cam.WaitForFrame(5)
		switch err.(type) {
		case nil:
		case *webcam.Timeout:
			continue
		default:
			return err
		}

		frame, err := c.cam.ReadFrame()
		if err != nil {
			// TODO: Add a better error handling
			return err
		}

		if len(frame) == 0 {
			continue
		}

		cb(frame)
	}
}

func (c *camera) Stop() error {
	return c.cam.StopStreaming()
}

func (c *camera) Info() Info {
	return Info{}
}

func (c *camera) Specs() []VideoSpec {
	specs := make([]VideoSpec, 0)
	for format := range c.cam.GetSupportedFormats() {
		// TODO: get width and height resolutions from camera
		specs = append(specs, VideoSpec{
			Width:  640,
			Height: 480,
			FrameFormat: c.formats[format],
		})
	}

	return specs
}
