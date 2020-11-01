package mediadevices

import (
	"math"
	"time"
)

type samplerFunc func() uint32

// newVideoSampler creates a video sampler that uses the actual video frame rate and
// the codec's clock rate to come up with a duration for each sample.
func newVideoSampler(clockRate uint32) samplerFunc {
	clockRateFloat := float64(clockRate)
	lastTimestamp := time.Now()

	return samplerFunc(func() uint32 {
		now := time.Now()
		duration := now.Sub(lastTimestamp).Seconds()
		samples := uint32(math.Round(clockRateFloat * duration))
		lastTimestamp = now
		return samples
	})
}

// newAudioSampler creates a audio sampler that uses a fixed latency and
// the codec's clock rate to come up with a duration for each sample.
func newAudioSampler(clockRate uint32, latency time.Duration) samplerFunc {
	samples := uint32(math.Round(float64(clockRate) * latency.Seconds()))
	return samplerFunc(func() uint32 {
		return samples
	})
}
