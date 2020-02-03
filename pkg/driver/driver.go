package driver

import (
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
)

type OpenCloser interface {
	Open() error
	Close() error
}

type Infoer interface {
	Info() Info
}

type Info struct {
	DeviceType DeviceType
}

type VideoCapable interface {
	Start(prop video.AdvancedProperty) (video.Reader, error)
	Stop() error
	Properties() []video.AdvancedProperty
}

type AudioCapable interface {
	Start(prop audio.AdvancedProperty) (audio.Reader, error)
	Stop() error
	Properties() []audio.AdvancedProperty
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
