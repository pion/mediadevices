package opus

import (
	"errors"
	"fmt"
	"math"

	"github.com/lherman-cs/opus"
	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/mediadevices/pkg/wave"
	"github.com/pion/mediadevices/pkg/wave/mixer"
)

type encoder struct {
	engine *opus.Encoder
	inBuff wave.Audio
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

	if params.ChannelMixer == nil {
		params.ChannelMixer = &mixer.MonoMixer{}
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

	channels := p.ChannelCount

	engine, err := opus.NewEncoder(p.SampleRate, channels, opus.AppVoIP)
	if err != nil {
		return nil, err
	}
	if err := engine.SetBitrate(params.BitRate); err != nil {
		return nil, err
	}

	rMix := audio.NewChannelMixer(channels, params.ChannelMixer)
	rBuf := audio.NewBuffer(int(targetLatency * float64(p.SampleRate) / 1000))
	e := encoder{
		engine: engine,
		reader: rMix(rBuf(r)),
	}
	return &e, nil
}

func (e *encoder) Read() ([]byte, func(), error) {
	buff, _, err := e.reader.Read()
	if err != nil {
		return nil, func() {}, err
	}

	encoded := make([]byte, 1024)
	switch b := buff.(type) {
	case *wave.Int16Interleaved:
		n, err := e.engine.Encode(b.Data, encoded)
		return encoded[:n:n], func() {}, err
	case *wave.Float32Interleaved:
		n, err := e.engine.EncodeFloat32(b.Data, encoded)
		return encoded[:n:n], func() {}, err
	default:
		return nil, func() {}, errors.New("unknown type of audio buffer")
	}
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
