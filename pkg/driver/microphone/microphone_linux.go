package microphone

import (
	"io"
	"time"

	"github.com/jfreymuth/pulse"
	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/mediadevices/pkg/wave"
)

type microphone struct {
	c           *pulse.Client
	id          string
	samplesChan chan<- []int16
}

func init() {
	pa, err := pulse.NewClient()
	if err != nil {
		// No pulseaudio
		return
	}
	defer pa.Close()
	sources, err := pa.ListSources()
	if err != nil {
		panic(err)
	}
	defaultSource, err := pa.DefaultSource()
	if err != nil {
		panic(err)
	}
	for _, source := range sources {
		priority := driver.PriorityNormal
		if defaultSource.ID() == source.ID() {
			priority = driver.PriorityHigh
		}
		driver.GetManager().Register(&microphone{id: source.ID()}, driver.Info{
			Label:      source.ID(),
			DeviceType: driver.Microphone,
			Priority:   priority,
		})
	}
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

	src, err := m.c.SourceByID(m.id)
	if err != nil {
		return nil, err
	}

	options = append(options,
		pulse.RecordSampleRate(p.SampleRate),
		pulse.RecordLatency(latency),
		pulse.RecordSource(src),
	)

	samplesChan := make(chan []int16, 1)

	handler := func(b []int16) (int, error) {
		samplesChan <- b
		return len(b), nil
	}

	stream, err := m.c.NewRecord(pulse.Int16Writer(handler), options...)
	if err != nil {
		return nil, err
	}

	reader := audio.ReaderFunc(func() (wave.Audio, error) {
		buff, ok := <-samplesChan
		if !ok {
			stream.Close()
			return nil, io.EOF
		}

		a := wave.NewInt16Interleaved(
			wave.ChunkInfo{
				Channels: p.ChannelCount,
				Len:      len(buff) / p.ChannelCount,
			},
		)
		copy(a.Data, buff)

		return a, nil
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
