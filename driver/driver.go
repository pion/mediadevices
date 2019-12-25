package driver

import (
	"github.com/pion/mediadevices/frame"
)

type State uint

const (
	StateClosed State = iota
	StateOpened
	StateStarted
	StateStopped
)

type DataCb func(b []byte)

type Driver interface {
	Open() error
	Stop() error
	Close() error
	Info() Info
}

type Info struct {
	Kind Kind
}

type VideoDriver interface {
	Driver
	Start(spec VideoSpec, cb DataCb) error
	Specs() []VideoSpec
}

type VideoSpec struct {
	Width, Height int
	FrameFormat   frame.Format
}

type AudioDriver interface {
	Driver
	Start(spec AudioSpec, cb DataCb) error
	Specs() []AudioSpec
}

type AudioSpec struct {
}

type QueryResult struct {
	ID     string
	Driver Driver
}
