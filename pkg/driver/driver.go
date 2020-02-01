package driver

import (
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
	"time"
)

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
	Start(setting VideoSetting) (video.Reader, error)
	Stop() error
	Settings() []VideoSetting
}

type VideoSetting struct {
	Width, Height int
	FrameFormat   frame.Format
}

type AudioCapable interface {
	Start(setting AudioSetting) (audio.Reader, error)
	Stop() error
	Settings() []AudioSetting
}

type AudioSetting struct {
	SampleRate int
	Latency    time.Duration
	Mono       bool
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
