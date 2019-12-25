package mediadevices

import (
	"time"

	"github.com/pion/webrtc/v2/pkg/media"
)

type sampler struct {
	clockRate     float64
	lastTimestamp time.Time
}

func newSampler(clockRate uint32) *sampler {
	return &sampler{
		clockRate:     float64(clockRate),
		lastTimestamp: time.Now(),
	}
}

func (s *sampler) sample(b []byte) media.Sample {
	now := time.Now()
	duration := now.Sub(s.lastTimestamp).Seconds()
	samples := uint32(s.clockRate * duration)
	s.lastTimestamp = now

	return media.Sample{Data: b, Samples: samples}
}
