package mediadevices

import "github.com/pion/mediadevices/pkg/driver"

import "math"

import "time"

type MediaStreamConstraints struct {
	Audio AudioTrackConstraints
	Video VideoTrackConstraints
}

type VideoTrackConstraints struct {
	Enabled       bool
	Width, Height int
	Codec         string
}

// fitnessDistance is an implementation for https://w3c.github.io/mediacapture-main/#dfn-fitness-distance
func (c *VideoTrackConstraints) fitnessDistance(s driver.VideoSetting) float64 {
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

type AudioTrackConstraints struct {
	Enabled    bool
	Codec      string
	SampleRate int
	Latency    time.Duration
}

// fitnessDistance is an implementation for https://w3c.github.io/mediacapture-main/#dfn-fitness-distance
func (c *AudioTrackConstraints) fitnessDistance(s driver.AudioSetting) float64 {
	var dist float64

	if s.SampleRate != c.SampleRate {
		actualSampleRate := float64(s.SampleRate)
		idealSampleRate := float64(c.SampleRate)
		max := math.Max(math.Abs(actualSampleRate), math.Abs(idealSampleRate))
		dist += math.Abs(actualSampleRate-idealSampleRate) / max
	}
	if s.Latency != c.Latency {
		actualLatency := float64(s.Latency)
		idealLatency := float64(c.Latency)
		max := math.Max(math.Abs(actualLatency), math.Abs(idealLatency))
		dist += math.Abs(actualLatency-idealLatency) / max
	}

	return dist
}
