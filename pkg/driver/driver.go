package driver

import (
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

// VideoRecorder is the interface implemented by video device drivers that outputs images.
type VideoRecorder interface {
	// VideoRecord starts recording video according to the selected property `p`
	// which is assumed to be the best match with user requested property `req`.
	VideoRecord(p, req prop.Media) (r video.Reader, err error)
}

// AudioRecorder is the interface implemented by audio device drivers that outputs PCM data.
type AudioRecorder interface {
	// AudioRecord starts recording audio according to the selected property `p`
	// which is assumed to be the best match with user requested property `req`.
	AudioRecord(p, req prop.Media) (r audio.Reader, err error)
}

// Priority represents device selection priority level
type Priority float32

const (
	// PriorityHigh is a value for system default devices
	PriorityHigh Priority = 0.1
	// PriorityNormal is a value for normal devices
	PriorityNormal Priority = 0.0
	// PriorityLow is a value for unrecommended devices
	PriorityLow Priority = -0.1
)

type Info struct {
	Label      string
	DeviceType DeviceType
	Priority   Priority
}

type Adapter interface {
	Open() error
	Close() error
	Properties() []prop.Media
}

type Driver interface {
	Adapter
	ID() string
	Info() Info
	Status() State
}
