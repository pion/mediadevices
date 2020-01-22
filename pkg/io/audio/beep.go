package audio

import (
	"io"

	"github.com/faiface/beep"
)

type beepStreamer struct {
	err  error
	r    Reader
	buff [][2]float32
}

func ToBeep(r Reader) beep.Streamer {
	if r == nil {
		panic("FromReader requires a non-nil Reader")
	}

	return &beepStreamer{r: r}
}

func (b *beepStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	// Since there was an error, the stream has to be drained
	if b.err != nil {
		return 0, false
	}

	if len(b.buff) < len(samples) {
		b.buff = append(b.buff, make([][2]float32, len(samples)-len(b.buff))...)
	}

	n, err := b.r.Read(b.buff[:len(samples)])
	if err != nil {
		b.err = err
		if err != io.EOF {
			return 0, false
		}
	}

	for i := 0; i < n; i++ {
		samples[i][0] = float64(b.buff[i][0])
		samples[i][1] = float64(b.buff[i][1])
	}

	return n, true
}

func (b *beepStreamer) Err() error {
	return b.err
}

type beepReader struct {
	s    beep.Streamer
	buff [][2]float64
}

func FromBeep(s beep.Streamer) Reader {
	if s == nil {
		panic("FromStreamer requires a non-nil beep.Streamer")
	}

	return &beepReader{s: s}
}

func (r *beepReader) Read(samples [][2]float32) (n int, err error) {
	if len(r.buff) < len(samples) {
		r.buff = append(r.buff, make([][2]float64, len(samples)-len(r.buff))...)
	}

	n, ok := r.s.Stream(r.buff[:len(samples)])
	if !ok {
		err := r.s.Err()
		if err == nil {
			err = io.EOF
		}

		return n, err
	}

	for i := 0; i < n; i++ {
		samples[i][0] = float32(r.buff[i][0])
		samples[i][1] = float32(r.buff[i][1])
	}

	return n, nil
}
