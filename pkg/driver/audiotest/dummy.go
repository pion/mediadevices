// Package audiotest provides dummy audio driver for testing.
package audiotest

import (
	"context"
	"io"
	"math"
	"time"

	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/mediadevices/pkg/wave"
)

func init() {
	driver.GetManager().Register(
		&dummy{}, driver.Info{Label: "AudioTest", DeviceType: driver.Microphone},
	)
}

type dummy struct {
	closed <-chan struct{}
	cancel func()
}

func (d *dummy) Open() error {
	ctx, cancel := context.WithCancel(context.Background())
	d.closed = ctx.Done()
	d.cancel = cancel
	return nil
}

func (d *dummy) Close() error {
	d.cancel()
	return nil
}

func (d *dummy) AudioRecord(p prop.Media) (audio.Reader, error) {
	var sin [100]float32
	for i := range sin {
		sin[i] = float32(math.Sin(2*math.Pi*float64(i)/100) * 0.25) // 480 Hz
	}

	if p.Latency == 0 {
		p.Latency = 20 * time.Millisecond
	}
	nSample := int(uint64(p.SampleRate) * uint64(p.Latency) / uint64(time.Second))

	nextReadTime := time.Now()
	var phase int

	closed := d.closed

	reader := audio.ReaderFunc(func() (wave.Audio, func(), error) {
		select {
		case <-closed:
			return nil, func() {}, io.EOF
		default:
		}

		time.Sleep(nextReadTime.Sub(time.Now()))
		nextReadTime = nextReadTime.Add(p.Latency)

		a := wave.NewFloat32Interleaved(
			wave.ChunkInfo{
				Channels: p.ChannelCount,
				Len:      nSample,
			},
		)

		for i := 0; i < nSample; i++ {
			phase++
			if phase >= 100 {
				phase = 0
			}
			for ch := 0; ch < p.ChannelCount; ch++ {
				a.SetFloat32(i, ch, wave.Float32Sample(sin[phase]))
			}
		}
		return a, func() {}, nil
	})
	return reader, nil
}

func (d *dummy) Properties() []prop.Media {
	return []prop.Media{
		{
			Audio: prop.Audio{
				SampleRate:   48000,
				Latency:      time.Millisecond * 20,
				ChannelCount: 1,
			},
		},
		{
			Audio: prop.Audio{
				SampleRate:   48000,
				Latency:      time.Millisecond * 20,
				ChannelCount: 2,
			},
		},
	}
}
