package mediadevices

import (
	"time"

	"github.com/pion/webrtc/v2/pkg/media"
)

type sampler struct {
	track         LocalTrack
	clockRate     float64
	lastTimestamp time.Time
}

func newSampler(track LocalTrack) *sampler {
	return &sampler{
		track:         track,
		clockRate:     float64(track.Codec().ClockRate),
		lastTimestamp: time.Now(),
	}
}

func (s *sampler) sample(b []byte) error {
	now := time.Now()
	duration := now.Sub(s.lastTimestamp).Seconds()
	samples := uint32(s.clockRate * duration)
	s.lastTimestamp = now

	sample := media.Sample{Data: b, Samples: samples}
	return s.track.WriteSample(sample)
}
