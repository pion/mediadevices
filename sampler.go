package mediadevices

import (
	"time"

	"github.com/pion/webrtc/v2/pkg/media"
)

type samplerFunc func(b []byte) error

func newVideoSampler(t LocalTrack) samplerFunc {
	clockRate := float64(t.Codec().ClockRate)
	lastTimestamp := time.Now()

	return samplerFunc(func(b []byte) error {
		now := time.Now()
		duration := now.Sub(lastTimestamp).Seconds()
		samples := uint32(clockRate * duration)
		lastTimestamp = now

		return t.WriteSample(media.Sample{Data: b, Samples: samples})
	})
}

func newAudioSampler(t LocalTrack, latency time.Duration) samplerFunc {
	samples := uint32(float64(t.Codec().ClockRate) * latency.Seconds())
	return samplerFunc(func(b []byte) error {
		return t.WriteSample(media.Sample{Data: b, Samples: samples})
	})
}
