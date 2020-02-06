package driver

import (
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

type VideoRecorder interface {
	VideoRecord(p prop.Media) (r video.Reader, err error)
}

type AudioRecorder interface {
	AudioRecord(p prop.Media) (r audio.Reader, err error)
}

type Adapter interface {
	Open() error
	Close() error
	Properties() []prop.Media
}

type Driver interface {
	Adapter
	ID() string
	Status() State
}
