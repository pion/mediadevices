package audio

import (
	"io"

	"github.com/faiface/beep"

	"github.com/pion/mediadevices/pkg/wave"
)

type beepStreamer struct {
	err error
	r   Reader
}

func ToBeep(r Reader) beep.Streamer {
	if r == nil {
		panic("FromReader requires a non-nil Reader")
	}

	return &beepStreamer{r: r}
}

func (b *beepStreamer) Stream(samples [][2]float64) (int, bool) {
	// Since there was an error, the stream has to be drained
	if b.err != nil {
		return 0, false
	}

	d, err := b.r.Read()
	if err != nil {
		b.err = err
		if err != io.EOF {
			return 0, false
		}
	}

	n := d.ChunkInfo().Len
	for i := 0; i < n; i++ {
		samples[i][0] = float64(wave.Float32SampleFormat.Convert(d.At(i, 0)).(wave.Float32Sample))
		samples[i][1] = float64(wave.Float32SampleFormat.Convert(d.At(i, 1)).(wave.Float32Sample))
	}

	return n, true
}

func (b *beepStreamer) Err() error {
	return b.err
}

type beepReader struct {
	s    beep.Streamer
	buff [][2]float64
	size int
}

func FromBeep(s beep.Streamer) Reader {
	if s == nil {
		panic("FromStreamer requires a non-nil beep.Streamer")
	}

	return &beepReader{
		s:    s,
		buff: make([][2]float64, 1024), // TODO: configure chunk size
	}
}

func (r *beepReader) Read() (wave.Audio, error) {
	out := wave.NewFloat32Interleaved(
		wave.ChunkInfo{Len: len(r.buff), Channels: 2, SamplingRate: 48000},
	)

	n, ok := r.s.Stream(r.buff)
	if !ok {
		err := r.s.Err()
		if err == nil {
			err = io.EOF
		}

		return nil, err
	}

	for i := 0; i < n; i++ {
		out.SetFloat32(i, 0, wave.Float32Sample(r.buff[i][0]))
		out.SetFloat32(i, 1, wave.Float32Sample(r.buff[i][1]))
	}

	return out, nil
}
