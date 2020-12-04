package mediadevices

import (
	"github.com/pion/mediadevices/pkg/prop"
)

type MediaStreamConstraints struct {
	Audio MediaOption
	Video MediaOption
	Codec *CodecSelector
}

// MediaTrackConstraints represents https://w3c.github.io/mediacapture-main/#dom-mediatrackconstraints
type MediaTrackConstraints struct {
	prop.MediaConstraints
	selectedMedia prop.Media
}

type MediaOption func(*MediaTrackConstraints)
