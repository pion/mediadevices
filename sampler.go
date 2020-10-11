package mediadevices

import (
	"math"
	"time"

	"github.com/pion/webrtc/v2/pkg/media"
)

type samplerFunc func(b []byte) error

// newSampler creates a sampler that estimates duration per sample
func newSampler(t LocalTrack) samplerFunc {
	var last time.Time
	first := true
	return samplerFunc(func(b []byte) error {
		now := time.Now()
		var samples uint32
		if first {
			first = false
		} else {
			duration := now.Sub(last)
			samples = uint32(math.Round(float64(t.Codec().ClockRate) * duration.Seconds()))
		}

		err := t.WriteSample(media.Sample{Data: b, Samples: samples})
		last = now
		return err
	})
}
