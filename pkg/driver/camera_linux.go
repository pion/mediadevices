package driver

// #include <linux/videodev2.h>
import "C"

import (
	"github.com/blackjack/webcam"
	"github.com/pion/mediadevices/pkg/frame"
)

// Camera implementation using v4l2
// Reference: https://linuxtv.org/downloads/v4l-dvb-apis/uapi/v4l/videodev.html#videodev
type camera struct {
	path    string
	cam     *webcam.Webcam
	formats map[webcam.PixelFormat]frame.Format
	reversedFormats map[frame.Format]webcam.PixelFormat
	settings 	[]VideoSetting
}

var _ VideoAdapter = &camera{}

func init() {
	// TODO: Probably try to get more cameras
	// Get default camera
	defaultCam := newCamera("/dev/video0")

	Manager.register(defaultCam)
}

func newCamera(path string) *camera {
	formats := map[webcam.PixelFormat]frame.Format{
		webcam.PixelFormat(C.V4L2_PIX_FMT_YUYV): frame.FormatYUYV,
		webcam.PixelFormat(C.V4L2_PIX_FMT_NV12): frame.FormatNV21,
		webcam.PixelFormat(C.V4L2_PIX_FMT_MJPEG): frame.FormatMJPEG,
	}

	reversedFormats := make(map[frame.Format]webcam.PixelFormat)
	for k, v := range formats {
		reversedFormats[v] = k
	}

	return &camera{
		path: path,
		formats: formats,
		reversedFormats: reversedFormats,
	}
}

func (c *camera) Open() error {
	cam, err := webcam.Open(c.path)
	if err != nil {
		return err
	}

	settings := make([]VideoSetting, 0)
	for format := range cam.GetSupportedFormats() {
		for _, frameSize := range cam.GetSupportedFrameSizes(format) {
			settings = append(settings, VideoSetting{
				Width:  int(frameSize.MaxWidth),
				Height: int(frameSize.MaxHeight),
				FrameFormat: c.formats[format],
			})
		}
	}

	c.cam = cam
	c.settings = settings
	return nil
}

func (c *camera) Close() error {
	c.settings = nil
	if c.cam == nil {
		return nil
	}

	return c.cam.StopStreaming()
}

func (c *camera) Start(setting VideoSetting, cb DataCb) error {
	pf := c.reversedFormats[setting.FrameFormat]
	_, _, _, err := c.cam.SetImageFormat(pf, uint32(setting.Width), uint32(setting.Height))
	if err != nil {
		return err
	}

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
	return Info{
		Kind: 		Video,
		DeviceType: Camera,
	}
}

func (c *camera) Settings() []VideoSetting {
	return c.settings
}
