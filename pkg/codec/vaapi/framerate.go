package vaapi

import (
	"time"
)

type framerateDetector struct {
	cnt   uint64
	stamp time.Time
	rate  uint32
}

func newFramerateDetector(initialRate uint32) *framerateDetector {
	return &framerateDetector{
		rate: initialRate,
	}
}

func (f *framerateDetector) Calc() uint32 {
	if f.cnt%16 == 0 {
		now := time.Now()
		interval := now.Sub(f.stamp)
		if !f.stamp.IsZero() {
			f.rate = uint32(interval.Nanoseconds()/(16*1000000))<<16 | 1000
			// denominator << 16 | numerator
		}
		f.stamp = now
	}
	f.cnt++
	return f.rate
}
