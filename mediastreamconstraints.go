package mediadevices

import (
	"fmt"
	"math"
	"strconv"

	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
)

type MediaStreamConstraints struct {
	Audio AudioTrackConstraints
	Video VideoTrackConstraints
}

type comparisons map[string]string

func (c comparisons) Add(actual, ideal interface{}) {
	c[fmt.Sprint(actual)] = fmt.Sprint(ideal)
}

// fitnessDistance is an implementation for https://w3c.github.io/mediacapture-main/#dfn-fitness-distance
func (c comparisons) fitnessDistance() float64 {
	var dist float64

	for actual, ideal := range c {
		if actual == ideal {
			continue
		}

		actual, err1 := strconv.ParseFloat(actual, 64)
		ideal, err2 := strconv.ParseFloat(ideal, 64)

		switch {
		// If both of the values are numeric, we need to normalize the values to get the distance
		case err1 == nil && err2 == nil:
			dist += math.Abs(actual-ideal) / math.Max(math.Abs(actual), math.Abs(ideal))
		// If both of the values are not numeric, the only comparison value is either 1 (matched) or 0 (not matched)
		case err1 != nil && err2 != nil:
			dist++
		// Comparing a numeric value with a non-numeric value is a an internal error, so panic.
		default:
			panic("fitnessDistance can't mix comparisons.")
		}
	}

	return dist
}

type VideoTrackConstraints struct {
	video.Property
	Enabled bool
	Codec   string
}

func (c *VideoTrackConstraints) fitnessDistance(prop video.AdvancedProperty) float64 {
	cmps := comparisons{}
	cmps.Add(prop.Width, c.Width)
	cmps.Add(prop.Height, c.Height)
	return cmps.fitnessDistance()
}

type AudioTrackConstraints struct {
	audio.Property
	Enabled bool
	Codec   string
}

func (c *AudioTrackConstraints) fitnessDistance(prop audio.AdvancedProperty) float64 {
	cmps := comparisons{}
	cmps.Add(prop.SampleRate, c.SampleRate)
	cmps.Add(prop.Latency, c.Latency)
	return cmps.fitnessDistance()
}
