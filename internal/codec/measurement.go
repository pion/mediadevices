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
	for {
		n, err = r.Read(buf)
		now = time.Now()

		if err != nil {
			if e, ok := err.(*mio.InsufficientBufferError); ok {
				buf = make([]byte, 2*e.RequiredSize)
				continue
			}

			if err == io.EOF {
				dur = now.Sub(start)
				totalBytes += n
				break
			}

			return 0, err
		}

		if now.After(end) {
			break
		}
		totalBytes += n // count bytes if the data arrived within the period
	}

	avg := float64(totalBytes*8) / dur.Seconds()
	return avg, nil
}
