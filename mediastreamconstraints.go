package mediadevices

import "github.com/pion/mediadevices/pkg/driver"

import "math"

type MediaStreamConstraints struct {
	Audio MediaTrackConstraints
	Video VideoTrackConstraints
}

type MediaTrackConstraints bool

type VideoTrackConstraints struct {
	Enabled       bool
	Width, Height int
	Codec         Codec
}

// fitnessDistance is an implementation for https://w3c.github.io/mediacapture-main/#dfn-fitness-distance
func (c *VideoTrackConstraints) fitnessDistance(s driver.VideoSpec) float64 {
	var dist float64

	if s.Width != c.Width {
		actualWidth := float64(s.Width)
		idealWidth := float64(c.Width)
		dist += math.Abs(actualWidth-idealWidth) / math.Max(math.Abs(actualWidth), math.Abs(idealWidth))
	}

	if s.Height != c.Height {
		actualHeight := float64(s.Height)
		idealHeight := float64(c.Height)
		dist += math.Abs(actualHeight-idealHeight) / math.Max(math.Abs(actualHeight), math.Abs(idealHeight))
	}

	return dist
}
