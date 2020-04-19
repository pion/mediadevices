package mediadevices

import (
	"github.com/pion/mediadevices/pkg/prop"
)

type MediaStreamConstraints struct {
	Audio MediaTrackConstraints
	Video MediaTrackConstraints
}

// MediaTrackConstraints represents https://w3c.github.io/mediacapture-main/#dom-mediatrackconstraints
type MediaTrackConstraints func(*prop.Media)
