// +build cpuusage

// This is not an actual benchmark test.
// Please manually check the CPU usage during the test.
// $ go test -bench . -tags cpuusage -benchtime 10s -benchmem

package camera

import (
	"testing"
	"time"

	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/prop"
)

func BenchmarkRead(b *testing.B) {
	c := newCamera("/dev/video0")

	props := map[string]prop.Media{
		"480p": prop.Media{
			Video: prop.Video{
				Width:       640,
				Height:      480,
				FrameFormat: frame.FormatYUYV,
			},
		},
		"720p": prop.Media{
			Video: prop.Video{
				Width:       1280,
				Height:      720,
				FrameFormat: frame.FormatYUYV,
			},
		},
		"1080p": prop.Media{
			Video: prop.Video{
				Width:       1920,
				Height:      1080,
				FrameFormat: frame.FormatYUYV,
			},
		},
	}
	for name, p := range props {
		time.Sleep(500 * time.Millisecond)
		p := p
		b.Run(name, func(b *testing.B) {
			if err := c.Open(); err != nil {
				b.Skip("You don't have camera.")
			}
			defer c.Close()

			r, err := c.VideoRecord(p)
			if err != nil {
				b.Skipf("Failed to capture image: %v", err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := r.Read()
				if err != nil {
					b.Fatalf("Failed to read: %v", err)
				}
			}
			b.StopTimer()
		})
	}
}
