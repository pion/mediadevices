package camera

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/blackjack/webcam"
	codecEngine "github.com/pion/codec"
	"github.com/pion/codec/h264"
	"github.com/pion/mediadevices/yuv"
	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
)

// Camera implementation using v4l2
// Reference: https://linuxtv.org/downloads/v4l-dvb-apis/uapi/v4l/videodev.html#videodev
type Camera struct {
	cam     *webcam.Webcam
	track   *webrtc.Track
	encoder codecEngine.Encoder
	opts    Options
}

func New(opts Options) (*Camera, error) {
	cam, err := webcam.Open("/dev/video0")
	if err != nil {
		return nil, err
	}

	width := opts.Width
	height := opts.Height

	var selectedFormat webcam.PixelFormat
	for v, k := range cam.GetSupportedFormats() {
		if strings.HasPrefix(k, "YUYV") {
			selectedFormat = v
			break
		}
	}

	if selectedFormat == 0 {
		return nil, fmt.Errorf("YUYV")
	}

	if _, _, _, err = cam.SetImageFormat(selectedFormat, uint32(width), uint32(height)); err != nil {
		return nil, err
	}

	var c *Camera
	switch opts.Codec {
	case webrtc.H264:
		// TODO: Replace "pion1" with device id instead
		track, err := opts.PC.NewTrack(webrtc.DefaultPayloadTypeH264, rand.Uint32(), "video", "pion1")
		if err != nil {
			return nil, err
		}
		// TODO: Remove hardcoded values
		encoder, err := h264.NewEncoder(h264.Options{
			Width:        width,
			Height:       height,
			MaxFrameRate: 30,
			Bitrate:      1000000,
		})
		if err != nil {
			return nil, err
		}
		c = &Camera{
			cam:     cam,
			track:   track,
			encoder: encoder,
			opts:    opts,
		}
	default:
		return nil, fmt.Errorf("%s is not currently supported", opts.Codec)
	}

	return c, nil
}

func (c *Camera) Start() error {
	if err := c.cam.StartStreaming(); err != nil {
		return err
	}

	decoder, err := frame.NewDecoder(frame.FormatYUY2)
	if err != nil {
		return err
	}

	lastTimestamp := time.Now()
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

		img, err := decoder.Decode(frame, c.opts.Width, c.opts.Height)
		if err != nil {
			continue
		}

		encoded, err := c.encoder.Encode(img)
		if err != nil {
			// TODO: Add a better error handling
			return err
		}

		now := time.Now()
		duration := now.Sub(lastTimestamp).Seconds()
		samples := uint32(clockRate * duration)
		lastTimestamp = now

		if err := c.track.WriteSample(media.Sample{Data: encoded, Samples: samples}); err != nil {
			// TODO: Add a better error handling
			continue
		}
	}
}

func (c *Camera) Track() *webrtc.Track {
	return c.track
}

func (c *Camera) Stop() {
	if c.cam == nil {
		return
	}

	c.cam.StopStreaming()
}
