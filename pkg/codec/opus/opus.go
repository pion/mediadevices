package opus

import (
	"fmt"
	"math"

	"github.com/lherman-cs/opus"
	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/mediadevices/pkg/wave"
)

type encoder struct {
	engine *opus.Encoder
	inBuff *wave.Float32Interleaved
	reader audio.Reader
}

var latencies = []float64{5, 10, 20, 40, 60}

func newEncoder(r audio.Reader, p prop.Media, params Params) (codec.ReadCloser, error) {
	if p.SampleRate == 0 {
		return nil, fmt.Errorf("opus: inProp.SampleRate is required")
	}

	if p.Latency == 0 {
		p.Latency = 20
	}

	if params.BitRate == 0 {
		params.BitRate = 32000
	}

	// Select the nearest supported latency
	var targetLatency float64
	// TODO: use p.Latency.Milliseconds() after Go 1.12 EOL
	latencyInMS := float64(p.Latency.Nanoseconds() / 1000000)
	nearestDist := math.Inf(+1)
	for _, latency := range latencies {
		dist := math.Abs(latency - latencyInMS)
		if dist >= nearestDist {
			break
		}

		nearestDist = dist
		targetLatency = latency
	}

	// Since audio.Reader only supports stereo mode, channels is always 2
	channels := 2

	engine, err := opus.NewEncoder(p.SampleRate, channels, opus.AppVoIP)
	if err != nil {
		return nil, err
	}
	if err := engine.SetBitrate(params.BitRate); err != nil {
		return nil, err
	}

	inBuffSize := int(targetLatency * float64(p.SampleRate) / 1000)
	inBuff := wave.NewFloat32Interleaved(
		wave.ChunkInfo{Channels: channels, Len: inBuffSize},
	)
	inBuff.Data = inBuff.Data[:0]
	e := encoder{engine, inBuff, r}
	return &e, nil
}

func (e *encoder) Read(p []byte) (n int, err error) {
	// While the buffer is not full, keep reading so that we meet the latency requirement
	nLatency := e.inBuff.ChunkInfo().Len * e.inBuff.ChunkInfo().Channels
	for len(e.inBuff.Data) < nLatency {
		buff, err := e.reader.Read()
		if err != nil {
			return 0, err
		}
		// TODO: convert audio format
		b, ok := buff.(*wave.Float32Interleaved)
		if !ok {
			panic("unsupported audio format")
		}
		switch {
		case b.Size.Channels == 1 && e.inBuff.ChunkInfo().Channels != 1:
			for _, d := range b.Data {
				for ch := 0; ch < e.inBuff.ChunkInfo().Channels; ch++ {
					e.inBuff.Data = append(e.inBuff.Data, d)
				}
			}
		case b.Size.Channels == e.inBuff.ChunkInfo().Channels:
			e.inBuff.Data = append(e.inBuff.Data, b.Data...)
		}
	}

	n, err = e.engine.EncodeFloat32(e.inBuff.Data[:nLatency], p)
	if err != nil {
		return n, err
	}
	e.inBuff.Data = e.inBuff.Data[nLatency:]

	return n, nil
}

func (e *encoder) SetBitRate(b int) error {
	panic("SetBitRate is not implemented")
}

func (e *encoder) ForceKeyFrame() error {
	panic("ForceKeyFrame is not implemented")
}

func (e *encoder) Close() error {
	return nil
}
