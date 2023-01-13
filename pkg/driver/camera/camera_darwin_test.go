// +build darwin

// $ go test -v . -tags darwin -run="^TestCameraFrameFormatSupport$"

package camera

import (
	"testing"

	"github.com/pion/mediadevices/pkg/avfoundation"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/prop"
)

func TestCameraFrameFormatSupport(t *testing.T) {
	devices, err := avfoundation.Devices(avfoundation.Video)
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) > 0 {
		c := newCamera(devices[0])
		if err := c.Open(); err != nil {
			t.Fatal(err)
		}
		defer c.Close()

		supportedFormats := make(map[frame.Format]struct{})
		for _, p := range c.Properties() {
			supportedFormats[p.FrameFormat] = struct{}{}
		}

		for _, format := range []frame.Format{
			frame.FormatI420,
			frame.FormatNV12,
			frame.FormatNV21,
			frame.FormatYUY2,
			frame.FormatUYVY,
		} {
			if _, ok := supportedFormats[format]; !ok {
				t.Logf("[%v] UNSUPPORTED", format)
				continue
			}
			r, err := c.VideoRecord(prop.Media{
				Video: prop.Video{
					Width:       640,
					Height:      480,
					FrameFormat: format,
				}})
			if err != nil {
				t.Logf("[%v] Failed to capture image: %v", format, err)
				continue
			}
			for i := 0; i < 10; i++ {
				_, _, err := r.Read()
				if err != nil {
					t.Logf("[%v] Failed to read: %v", format, err)
					continue
				}
			}
			t.Logf("[%v] OK", format)
		}
	}
}
