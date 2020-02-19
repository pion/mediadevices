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

	nextReadTime := time.Now()
	var phase int

	closed := d.closed

	reader := audio.ReaderFunc(func(samples [][2]float32) (int, error) {
		select {
		case <-closed:
			return 0, io.EOF
		default:
		}

		time.Sleep(nextReadTime.Sub(time.Now()))
		dur := time.Second * time.Duration(len(samples)) / 48000
		nextReadTime = nextReadTime.Add(dur)

		for i := range samples {
			phase++
			if phase >= 100 {
				phase = 0
			}
			samples[i][0] = sin[phase]
			if p.ChannelCount == 2 {
				samples[i][1] = sin[phase]
			}
		}
		return len(samples), nil
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
