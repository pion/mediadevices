package driver

import (
	"github.com/jfreymuth/pulse"
)

type microphone struct {
	c *pulse.Client
	s *pulse.RecordStream
}

var _ AudioAdapter = &microphone{}

func init() {
	GetManager().Register(&microphone{})
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
	m.c.Close()
	if m.s != nil {
		m.s.Close()
	}

	return nil
}

func (m *microphone) Start(setting AudioSetting, cb AudioDataCb) error {
	buff := make([]int16, 960)
	n := 0
	handler := func(b []int16) {
		for n+len(b) >= 960 {
			nCopied := copy(buff[n:], b)
			cb(buff)
			n = 0
			b = b[nCopied:]
		}
		nCopied := copy(buff[n:], b)
		n += nCopied
	}

	stream, err := m.c.NewRecord(handler, pulse.RecordSampleRate(48000), pulse.RecordLatency(0.005))
	if err != nil {
		return err
	}

	stream.Start()
	m.s = stream
	return nil
}

func (m *microphone) Stop() error {
	m.s.Stop()
	return nil
}

func (m *microphone) Info() Info {
	return Info{
		Kind:       Audio,
		DeviceType: Microphone,
	}
}

func (m *microphone) Settings() []AudioSetting {
	src, err := m.c.DefaultSource()
	if err != nil {
		return nil
	}

	return []AudioSetting{AudioSetting{src.SampleRate()}}
}
