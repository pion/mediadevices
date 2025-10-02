package codec

import (
	"time"
)

type BitrateTracker struct {
	windowSize time.Duration
	buffer     []int
	times      []time.Time
}

func NewBitrateTracker(windowSize time.Duration) *BitrateTracker {
	return &BitrateTracker{
		windowSize: windowSize,
	}
}

func (bt *BitrateTracker) AddFrame(sizeBytes int, timestamp time.Time) {
	bt.buffer = append(bt.buffer, sizeBytes)
	bt.times = append(bt.times, timestamp)

	// Remove old entries outside the window
	cutoff := timestamp.Add(-bt.windowSize)
	i := 0
	for ; i < len(bt.times); i++ {
		if bt.times[i].After(cutoff) {
			break
		}
	}
	bt.buffer = bt.buffer[i:]
	bt.times = bt.times[i:]
}

func (bt *BitrateTracker) GetBitrate() float64 {
	if len(bt.times) < 2 {
		return 0
	}
	totalBytes := 0
	for _, b := range bt.buffer {
		totalBytes += b
	}
	duration := bt.times[len(bt.times)-1].Sub(bt.times[0]).Seconds()
	if duration <= 0 {
		return 0
	}
	return float64(totalBytes*8) / duration // bits per second
}
