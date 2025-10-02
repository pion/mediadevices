package codec

import (
	"math"
	"testing"
	"time"
)

func TestBitrateTracker(t *testing.T) {
	packetSize := 1000
	bt := NewBitrateTracker(time.Second)
	bt.AddFrame(packetSize, time.Now())
	bt.AddFrame(packetSize, time.Now().Add(time.Millisecond*100))
	bt.AddFrame(packetSize, time.Now().Add(time.Millisecond*999))
	eps := float64(packetSize*8) / 10
	if got, want := bt.GetBitrate(), float64(packetSize*8)*3; math.Abs(got-want) > eps {
		t.Fatalf("GetBitrate() = %v, want %v (|diff| <= %v)", got, want, eps)
	}
}
