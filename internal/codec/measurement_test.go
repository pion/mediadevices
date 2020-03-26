package codec

import (
	"io"
	"sync"
	"testing"
	"time"
)

func TestMeasureBitRateStatic(t *testing.T) {
	r, w := io.Pipe()
	dur := time.Second * 5
	dataSize := 1000
	var precision float64 = 8 // 1 byte
	var wg sync.WaitGroup
	wg.Add(1)
	done := make(chan struct{})

	go func() {
		data := make([]byte, dataSize)
		// Make sure that this goroutine is synchronized with main goroutine
		wg.Done()
		ticker := time.NewTicker(time.Second)

		for {
			select {
			case <-ticker.C:
				w.Write(data)
			case <-done:
				w.Close()
				return
			}
		}
	}()

	wg.Wait()
	bitrate, err := MeasureBitRate(r, dur)
	if err != nil {
		t.Error(err)
	}
	done <- struct{}{}

	expected := float64(dataSize * 8)
	if bitrate < expected-precision || bitrate > expected+precision {
		t.Fatalf("expected: %f (with %f precision), but got %f", expected, precision, bitrate)
	}
}

func TestMeasureBitRateDynamic(t *testing.T) {
	r, w := io.Pipe()
	dur := time.Second * 5
	dataSize := 1000
	var precision float64 = 8 // 1 byte
	var wg sync.WaitGroup
	wg.Add(1)
	done := make(chan struct{})

	go func() {
		data := make([]byte, dataSize)
		wg.Done()
		ticker := time.NewTicker(time.Millisecond * 500)
		var count int

		for {
			select {
			case <-ticker.C:
				w.Write(data)
				count++
				// Wait until 4 slow ticks, which is also equal to 2 seconds
				if count == 4 {
					ticker.Stop()
					// Speed up the tick by 2 times for the rest
					ticker = time.NewTicker(time.Millisecond * 250)
				}
			case <-done:
				w.Close()
				return
			}
		}
	}()

	wg.Wait()
	bitrate, err := MeasureBitRate(r, dur)
	if err != nil {
		t.Error(err)
	}
	done <- struct{}{}

	// Measure the expected bitrate using the number of ticks or write * data size / duration in seconds
	// For the first 2 seconds: #ticks = 2000ms / 500ms = 4
	// For the last 3 seconds: #ticks = 3000ms / 250ms = 12
	// So, in 5 seconds, there will be 16 ticks
	expected := float64(16*(dataSize*8)) / 5
	if bitrate < expected-precision || bitrate > expected+precision {
		t.Fatalf("expected: %f (with %f precision), but got %f", expected, precision, bitrate)
	}
}
