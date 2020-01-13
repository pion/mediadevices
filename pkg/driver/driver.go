package driver

import (
	"github.com/pion/mediadevices/pkg/frame"
)

type State uint

const (
	StateClosed State = iota
	StateOpened
	StateStarted
	StateStopped
)

type DataCb func(b []byte)
type AudioDataCb func(b []int16)

type OpenCloser interface {
	Open() error
	Close() error
}

type Infoer interface {
	Info() Info
}

type Info struct {
	Kind       Kind
	DeviceType DeviceType
}

type VideoCapable interface {
	Start(setting VideoSetting, cb DataCb) error
	Stop() error
	Settings() []VideoSetting
}

type VideoSetting struct {
	Width, Height int
	FrameFormat   frame.Format
}

type AudioCapable interface {
	Start(setting AudioSetting, cb AudioDataCb) error
	Stop() error
	Settings() []AudioSetting
}

type AudioSetting struct {
	SampleRate int
}

type Adapter interface {
	OpenCloser
	Infoer
}

type VideoAdapter interface {
	Adapter
	VideoCapable
}

type AudioAdapter interface {
	Adapter
	AudioCapable
}

type Driver interface {
	Adapter
	ID() string
	Status() State
}

type VideoDriver interface {
	Driver
	VideoCapable
}

type AudioDriver interface {
	Driver
	AudioCapable
}
