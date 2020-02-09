package microphone

import (
	"io"
	"time"

	"github.com/jfreymuth/pulse"
	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/prop"
)

type microphone struct {
	c           *pulse.Client
	samplesChan chan<- []float32
}

func init() {
	driver.GetManager().Register(&microphone{})
}

func (m *microphone) Open() error {
	var err error
	m.c, err = pulse.NewClient()
	if err != nil {
		return err
	}

	return nil
}

func (m *microphone) Close() error {
	if m.samplesChan != nil {
		close(m.samplesChan)
		m.samplesChan = nil
	}

	m.c.Close()
	return nil
}

func (m *microphone) AudioRecord(p prop.Media) (audio.Reader, error) {
	var options []pulse.RecordOption
	if p.ChannelCount == 1 {
		options = append(options, pulse.RecordMono)
	} else {
		options = append(options, pulse.RecordStereo)
	}
	latency := p.Latency.Seconds()
	options = append(options, pulse.RecordSampleRate(p.SampleRate), pulse.RecordLatency(latency))

	samplesChan := make(chan []float32, 1)
	var buff []float32
	var bi int
	var more bool

	handler := func(b []float32) {
		samplesChan <- b
	}

	stream, err := m.c.NewRecord(handler, options...)
	if err != nil {
		return nil, err
	}

	reader := audio.ReaderFunc(func(samples [][2]float32) (n int, err error) {
		for i := range samples {
			// if we don't have anything left in buff, we'll wait until we receive
			// more samples
			if bi == len(buff) {
				buff, more = <-samplesChan
				if !more {
					stream.Close()
					return i, io.EOF
				}
				bi = 0
			}

			samples[i][0] = buff[bi]
			if p.ChannelCount == 2 {
				samples[i][1] = buff[bi+1]
				bi++
			}
			bi++
		}

		return len(samples), nil
	})

	stream.Start()
	m.samplesChan = samplesChan
	return reader, nil
}

func (m *microphone) Properties() []prop.Media {
	// TODO: Get actual properties
	monoProp := prop.Media{
		Audio: prop.Audio{
			SampleRate:   48000,
			Latency:      time.Millisecond * 20,
			ChannelCount: 1,
		},
	}

	stereoProp := monoProp
	stereoProp.ChannelCount = 2

	return []prop.Media{monoProp, stereoProp}
}
