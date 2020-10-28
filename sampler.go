package mediadevices

import (
	"math"
	"time"

	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
)

type samplerFunc func(b []byte) error

// newVideoSampler creates a video sampler that uses the actual video frame rate and
// the codec's clock rate to come up with a duration for each sample.
func newVideoSampler(t *webrtc.Track) samplerFunc {
	clockRate := float64(t.Codec().ClockRate)
	lastTimestamp := time.Now()

	return samplerFunc(func(b []byte) error {
		now := time.Now()
		duration := now.Sub(lastTimestamp).Seconds()
		samples := uint32(math.Round(clockRate * duration))
		lastTimestamp = now

		return t.WriteSample(media.Sample{Data: b, Samples: samples})
	})
}

// newAudioSampler creates a audio sampler that uses a fixed latency and
// the codec's clock rate to come up with a duration for each sample.
func newAudioSampler(t *webrtc.Track, latency time.Duration) samplerFunc {
	samples := uint32(math.Round(float64(t.Codec().ClockRate) * latency.Seconds()))
	return samplerFunc(func(b []byte) error {
		return t.WriteSample(media.Sample{Data: b, Samples: samples})
	})
}
