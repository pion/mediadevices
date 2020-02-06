package mediadevices

import (
	"github.com/pion/mediadevices/pkg/prop"
)

type MediaStreamConstraints struct {
	Audio MediaOption
	Video MediaOption
}

// MediaTrackConstraints represents https://w3c.github.io/mediacapture-main/#dom-mediatrackconstraints
type MediaTrackConstraints struct {
	prop.Media
	Enabled bool
	Codec   string
}

type MediaOption func(*MediaTrackConstraints)
