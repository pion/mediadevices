package mediadevices

import (
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
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
	// VideoTransform will be used to transform the video that's coming from the driver.
	// So, basically it'll look like following: driver -> VideoTransform -> codec
	VideoTransform video.TransformFunc
	// AudioTransform will be used to transform the audio that's coming from the driver.
	// So, basically it'll look like following: driver -> AudioTransform -> code
	AudioTransform audio.TransformFunc
}

type MediaOption func(*MediaTrackConstraints)
