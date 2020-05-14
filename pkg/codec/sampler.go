package codec

import (
	"math"
	"time"
)

// SamplerFunc returns the number of samples. Each invocation may return different
// different amount of samples due to how it's calculated/measured.
type SamplerFunc func() uint32

// NewVideoSampler creates a video sampler that uses the actual video frame rate and
// the codec's clock rate to come up with a duration for each sample.
func NewVideoSampler(clockRate uint32) SamplerFunc {
	clockRateFloat := float64(clockRate)
	lastTimestamp := time.Now()

	return SamplerFunc(func() uint32 {
		now := time.Now()
		duration := now.Sub(lastTimestamp).Seconds()
		samples := uint32(math.Round(clockRateFloat * duration))
		lastTimestamp = now

		return samples
	})
}

// NewAudioSampler creates a audio sampler that uses a fixed latency and
// the codec's clock rate to come up with a duration for each sample.
func NewAudioSampler(clockRate uint32, latency time.Duration) SamplerFunc {
	samples := uint32(math.Round(float64(clockRate) * latency.Seconds()))
	return SamplerFunc(func() uint32 {
		return samples
	})
}
