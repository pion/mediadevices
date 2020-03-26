package codec

import (
	"io"
	"time"

	mio "github.com/pion/mediadevices/pkg/io"
)

// MeasureBitRate measures average bitrate after dur by reading r as fast as possible
func MeasureBitRate(r io.Reader, dur time.Duration) (float64, error) {
	var n, totalBytes int
	var err error
	buf := make([]byte, 1024)
	start := time.Now()
	now := start
	end := now.Add(dur)
	for now.Before(end) {
		n, err = r.Read(buf)
		if err != nil {
			if e, ok := err.(*mio.InsufficientBufferError); ok {
				buf = make([]byte, 2*e.RequiredSize)
				continue
			}

			if err == io.EOF {
				totalBytes += n
				break
			}

			return 0, err
		}

		totalBytes += n
		now = time.Now()
	}

	elapsed := time.Now().Sub(start).Seconds()
	avg := float64(totalBytes*8) / elapsed
	return avg, nil
}
