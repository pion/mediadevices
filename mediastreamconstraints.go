package mediadevices

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/pion/mediadevices/pkg/driver"
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
	Enabled       bool
	Width, Height int
	Codec         string
}

func (c *VideoTrackConstraints) fitnessDistance(s driver.VideoSetting) float64 {
	cmps := comparisons{}
	cmps.Add(s.Width, c.Width)
	cmps.Add(s.Height, c.Height)
	return cmps.fitnessDistance()
}

type AudioTrackConstraints struct {
	Enabled    bool
	Codec      string
	SampleRate int
	Latency    time.Duration
}

func (c *AudioTrackConstraints) fitnessDistance(s driver.AudioSetting) float64 {
	cmps := comparisons{}
	cmps.Add(s.SampleRate, c.SampleRate)
	cmps.Add(s.Latency, c.Latency)
	return cmps.fitnessDistance()
}
