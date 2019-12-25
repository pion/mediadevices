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

type OpenCloser interface {
	Open() error
	Close() error
}

type Infoer interface {
	Info() Info
}

type Info struct {
	Kind Kind
}

type VideoCapable interface {
	Start(spec VideoSpec, cb DataCb) error
	Stop() error
	Specs() []VideoSpec
}

type VideoSpec struct {
	Width, Height int
	FrameFormat   frame.Format
}

type AudioCapable interface {
	Start(spec AudioSpec, cb DataCb) error
	Stop() error
	Specs() []AudioSpec
}

type AudioSpec struct {
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
