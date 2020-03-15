package opus

import (
	"fmt"
	"io"
	"math"
	"reflect"
	"unsafe"

	"github.com/lherman-cs/opus"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/prop"
)

type encoder struct {
	engine *opus.Encoder
	inBuff [][2]float32
	reader audio.Reader
}

var latencies = []float64{5, 10, 20, 40, 60}

var _ io.ReadCloser = &encoder{}

func newEncoder(r audio.Reader, p prop.Media, params Params) (io.ReadCloser, error) {
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

	inBuffSize := targetLatency * float64(p.SampleRate) / 1000
	inBuff := make([][2]float32, int(inBuffSize))
	e := encoder{engine, inBuff, r}
	return &e, nil
}

func flatten(samples [][2]float32) []float32 {
	if len(samples) == 0 {
		return nil
	}

	data := uintptr(unsafe.Pointer(&samples[0]))
	l := len(samples) * 2
	return *(*[]float32)(unsafe.Pointer(&reflect.SliceHeader{Data: data, Len: l, Cap: l}))
}

func (e *encoder) Read(p []byte) (n int, err error) {
	var curN int

	// While the buffer is not full, keep reading so that we meet the latency requirement
	for curN < len(e.inBuff) {
		n, err = e.reader.Read(e.inBuff[curN:])
		if err != nil {
			return 0, err
		}

		curN += n
	}

	n, err = e.engine.EncodeFloat32(flatten(e.inBuff), p)
	if err != nil {
		return n, err
	}

	return n, nil
}

func (e *encoder) Close() error {
	return nil
}
