package camera

// #include <linux/videodev2.h>
import "C"

import (
	"context"
	"errors"
	"image"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/pion/mediadevices/pkg/driver/availability"

	"github.com/blackjack/webcam"
	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

const (
	maxEmptyFrameCount = 5
	prioritizedDevice  = "video0"
)

var (
	errReadTimeout = errors.New("read timeout")
	errEmptyFrame  = errors.New("empty frame")
	// Reference: https://commons.wikimedia.org/wiki/File:Vector_Video_Standards2.svg
	supportedResolutions = [][2]int{
		{320, 240},
		{640, 480},
		{768, 576},
		{800, 600},
		{1024, 768},
		{1280, 854},
		{1280, 960},
		{1280, 1024},
		{1400, 1050},
		{1600, 1200},
		{2048, 1536},
		{320, 200},
		{800, 480},
		{854, 480},
		{1024, 600},
		{1152, 768},
		{1280, 720},
		{1280, 768},
		{1366, 768},
		{1280, 800},
		{1440, 900},
		{1440, 960},
		{1680, 1050},
		{1920, 1080},
		{2048, 1080},
		{1920, 1200},
		{2560, 1600},
	}
)

const bufCount = 2

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
	prevFrameTime   time.Time
}

func init() {
	Initialize()
}

// Initialize finds and registers camera devices. This is part of an experimental API.
func Initialize() {
	// Clear all registered camera devices to prevent duplicates.
	// If first initalize call, this will be a noop.
	manager := driver.GetManager()
	for _, d := range manager.Query(driver.FilterVideoRecorder()) {
		manager.Delete(d.ID())
	}
	discovered := make(map[string]struct{})
	discover(discovered, "/dev/v4l/by-id/*")
	discover(discovered, "/dev/v4l/by-path/*")
	discover(discovered, "/dev/video*")
}

func discover(discovered map[string]struct{}, pattern string) {
	devices, err := filepath.Glob(pattern)
	if err != nil {
		// No v4l device.
		return
	}
	for _, device := range devices {
		label := filepath.Base(device)
		reallink, err := os.Readlink(device)
		if err != nil {
			reallink = label
		} else {
			reallink = filepath.Base(reallink)
		}
		if _, ok := discovered[reallink]; ok {
			continue
		}

		discovered[reallink] = struct{}{}
		cam := newCamera(device)
		priority := driver.PriorityNormal
		if reallink == prioritizedDevice {
			priority = driver.PriorityHigh
		}

		var name, busInfo string
		if webcamCam, err := webcam.Open(cam.path); err == nil {
			name, _ = webcamCam.GetName()
			busInfo, _ = webcamCam.GetBusInfo()
		}

		driver.GetManager().Register(cam, driver.Info{
			// 	Source: https://www.kernel.org/doc/html/v4.9/media/uapi/v4l/vidioc-querycap.html
			//	Name of the device, a NUL-terminated UTF-8 string. For example: “Yoyodyne TV/FM”. One driver may support
			//	different brands or models of video hardware. This information is intended for users, for example in a
			//	menu of available devices. Since multiple TV cards of the same brand may be installed which are
			//	supported by the same driver, this name should be combined with the character device file name
			//	(e.g. /dev/video2) or the bus_info string to avoid ambiguities.
			Name:       name + LabelSeparator + busInfo,
			Label:      label + LabelSeparator + reallink,
			DeviceType: driver.Camera,
			Priority:   priority,
		})
	}
}

func newCamera(path string) *camera {
	formats := map[webcam.PixelFormat]frame.Format{
		webcam.PixelFormat(C.V4L2_PIX_FMT_YUV420): frame.FormatI420,
		webcam.PixelFormat(C.V4L2_PIX_FMT_NV21):   frame.FormatNV21,
		webcam.PixelFormat(C.V4L2_PIX_FMT_NV12):   frame.FormatNV12,
		webcam.PixelFormat(C.V4L2_PIX_FMT_YUYV):   frame.FormatYUYV,
		webcam.PixelFormat(C.V4L2_PIX_FMT_UYVY):   frame.FormatUYVY,
		webcam.PixelFormat(C.V4L2_PIX_FMT_MJPEG):  frame.FormatMJPEG,
		webcam.PixelFormat(C.V4L2_PIX_FMT_Z16):    frame.FormatZ16,
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

func getCameraReadTimeout() uint32 {
	// default to 5 seconds
	var readTimeoutSec uint32 = 5
	if val, ok := os.LookupEnv("PION_MEDIADEVICES_CAMERA_READ_TIMEOUT"); ok {
		if valInt, err := strconv.Atoi(val); err == nil {
			if valInt > 0 {
				readTimeoutSec = uint32(valInt)
			}
		}
	}
	return readTimeoutSec
}

func (c *camera) Open() error {
	cam, err := webcam.Open(c.path)
	if err != nil {
		return err
	}

	// Buffering should be handled in higher level.
	err = cam.SetBufferCount(bufCount)
	if err != nil {
		return err
	}

	c.prevFrameTime = time.Now()
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

	if p.FrameRate > 0 {
		err = c.cam.SetFramerate(float32(p.FrameRate))
		if err != nil {
			return nil, err
		}
	}

	if err := c.cam.StartStreaming(); err != nil {
		return nil, err
	}

	cam := c.cam

	readTimeoutSec := getCameraReadTimeout()

	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	var buf []byte
	r := video.ReaderFunc(func() (img image.Image, release func(), err error) {
		// Lock to avoid accessing the buffer after StopStreaming()
		c.mutex.Lock()
		defer c.mutex.Unlock()

		// Wait until a frame is ready
		for i := 0; i < maxEmptyFrameCount; i++ {
			if ctx.Err() != nil {
				// Return EOF if the camera is already closed.
				return nil, func() {}, io.EOF
			}

			if p.DiscardFramesOlderThan != 0 && time.Now().Sub(c.prevFrameTime) >= p.DiscardFramesOlderThan {
				for i := 0; i < bufCount; i++ {
					_ = cam.WaitForFrame(readTimeoutSec)
					_, _ = cam.ReadFrame()
				}
			}

			err := cam.WaitForFrame(readTimeoutSec)
			switch err.(type) {
			case nil:
			case *webcam.Timeout:
				return nil, func() {}, errReadTimeout
			default:
				// Camera has been stopped.
				return nil, func() {}, err
			}

			b, err := cam.ReadFrame()
			if err != nil {
				// Camera has been stopped.
				return nil, func() {}, err
			}

			if p.DiscardFramesOlderThan != 0 {
				c.prevFrameTime = time.Now()
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
		return nil, func() {}, errEmptyFrame
	})

	return r, nil
}

func (c *camera) Properties() []prop.Media {
	properties := make([]prop.Media, 0)
	for format := range c.cam.GetSupportedFormats() {
		for _, frameSize := range c.cam.GetSupportedFrameSizes(format) {
			supportedFormat, ok := c.formats[format]
			if !ok {
				continue
			}

			if frameSize.StepWidth == 0 || frameSize.StepHeight == 0 {
				framerates := c.cam.GetSupportedFramerates(format, uint32(frameSize.MaxWidth), uint32(frameSize.MaxHeight))
				// If the camera doesn't support framerate, we just add the resolution and format
				if len(framerates) == 0 {
					properties = append(properties, prop.Media{
						Video: prop.Video{
							Width:       int(frameSize.MaxWidth),
							Height:      int(frameSize.MaxHeight),
							FrameFormat: supportedFormat,
						},
					})
					continue
				}

				for _, framerate := range framerates {
					for _, fps := range enumFramerate(framerate) {
						properties = append(properties, prop.Media{
							Video: prop.Video{
								Width:       int(frameSize.MaxWidth),
								Height:      int(frameSize.MaxHeight),
								FrameFormat: supportedFormat,
								FrameRate:   fps,
							},
						})
					}
				}
			} else {
				// FIXME: we should probably use a custom data structure to capture all of the supported resolutions
				for _, supportedResolution := range supportedResolutions {
					minWidth, minHeight := int(frameSize.MinWidth), int(frameSize.MinHeight)
					maxWidth, maxHeight := int(frameSize.MaxWidth), int(frameSize.MaxHeight)
					stepWidth, stepHeight := int(frameSize.StepWidth), int(frameSize.StepHeight)
					width, height := supportedResolution[0], supportedResolution[1]

					if width < minWidth || width > maxWidth ||
						height < minHeight || height > maxHeight {
						continue
					}

					if (width-minWidth)%stepWidth != 0 ||
						(height-minHeight)%stepHeight != 0 {
						continue
					}

					framerates := c.cam.GetSupportedFramerates(format, uint32(width), uint32(height))
					if len(framerates) == 0 {
						properties = append(properties, prop.Media{
							Video: prop.Video{
								Width:       width,
								Height:      height,
								FrameFormat: supportedFormat,
							},
						})
						continue
					}

					for _, framerate := range framerates {
						for _, fps := range enumFramerate(framerate) {
							properties = append(properties, prop.Media{
								Video: prop.Video{
									Width:       width,
									Height:      height,
									FrameFormat: supportedFormat,
									FrameRate:   fps,
								},
							})
						}
					}
				}
			}
		}
	}
	return properties
}

func (c *camera) IsAvailable() (bool, error) {
	var err error

	// close the opened file descriptor as quickly as possible and in all cases, including panics
	func() {
		var cam *webcam.Webcam
		if cam, err = webcam.Open(c.path); err == nil {
			defer cam.Close()
			var index int32
			// "Drivers must implement all the input ioctls when the device has one or more inputs..."
			// Source: https://www.kernel.org/doc/html/latest/userspace-api/media/v4l/video.html?highlight=vidioc_enuminput
			if index, err = cam.GetInput(); err == nil {
				err = cam.SelectInput(uint32(index))
			}
		}
	}()

	var errno syscall.Errno
	errors.As(err, &errno)

	// See https://man7.org/linux/man-pages/man3/errno.3.html
	switch {
	case err == nil:
		return true, nil
	case errno == syscall.EBUSY:
		return false, availability.ErrBusy
	case errno == syscall.ENODEV || errno == syscall.ENOENT:
		return false, availability.ErrNoDevice
	default:
		return false, availability.NewError(errno.Error())
	}
}

// enumFramerate returns a list of fps options from a FrameRate struct.
// discrete framerates will return a list of 1 fps element.
// stepwise framerates will return a list of all possible fps options.
func enumFramerate(framerate webcam.FrameRate) []float32 {
	var framerates []float32
	if framerate.StepNumerator == 0 && framerate.StepDenominator == 0 {
		fr, err := calcFramerate(framerate.MaxNumerator, framerate.MaxDenominator)
		if err != nil {
			return framerates
		}
		framerates = append(framerates, fr)
	} else {
		for n := framerate.MinNumerator; n <= framerate.MaxNumerator; n += framerate.StepNumerator {
			for d := framerate.MinDenominator; d <= framerate.MaxDenominator; d += framerate.StepDenominator {
				fr, err := calcFramerate(n, d)
				if err != nil {
					continue
				}
				framerates = append(framerates, fr)
			}
		}
	}
	return framerates
}

// calcFramerate turns fraction into a float32 fps value.
func calcFramerate(numerator uint32, denominator uint32) (float32, error) {
	if denominator == 0 {
		return 0, errors.New("framerate denominator is zero")
	}
	// round to three decimal places to avoid floating point precision issues
	return float32(math.Round(1000.0/((float64(numerator))/float64(denominator))) / 1000), nil
}
