package opus

import (
	"fmt"
	"testing"
	"time"
)

func TestLatency_Validate(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		for _, l := range []Latency{
			Latency2500us, Latency5ms, Latency10ms, Latency20ms, Latency40ms, Latency60ms,
		} {
			if !l.Validate() {
				t.Errorf("Defined Latency(%v) must be valid", l)
			}
		}
	})
	t.Run("Invalid", func(t *testing.T) {
		for _, l := range []Latency{
			0, Latency(time.Second),
		} {
			if l.Validate() {
				t.Errorf("Latency(%v) must be valid", l)
			}
		}
	})
}

func TestLatency_samples(t *testing.T) {
	testCases := []struct {
		latency    Latency
		sampleRate int
		samples    int
	}{
		{Latency5ms, 48000, 240},
		{Latency20ms, 16000, 320},
		{Latency20ms, 48000, 960},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(fmt.Sprintf("%v_%d", time.Duration(testCase.latency), testCase.sampleRate), func(t *testing.T) {
			samples := testCase.latency.samples(testCase.sampleRate)
			if samples != testCase.samples {
				t.Errorf("Expected samples: %d, got: %d", testCase.samples, samples)
			}
		})
	}
}
