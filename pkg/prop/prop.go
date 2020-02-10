package prop

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/pion/mediadevices/pkg/frame"
)

type Media struct {
	Video
	Audio
	Codec
}

func (p *Media) FitnessDistance(o Media) float64 {
	cmps := comparisons{}
	cmps.add(p.Width, o.Width)
	cmps.add(p.Height, o.Height)
	cmps.add(p.SampleRate, o.SampleRate)
	cmps.add(p.Latency, o.Latency)
	return cmps.fitnessDistance()
}

type comparisons map[string]string

func (c comparisons) add(actual, ideal interface{}) {
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

// Video represents a video's properties
type Video struct {
	Width, Height int
	FrameRate     float32
	FrameFormat   frame.Format
}

// Audio represents an audio's properties
type Audio struct {
	ChannelCount int
	Latency      time.Duration
	SampleRate   int
	SampleSize   int
}

// Codec represents an codec's encoding properties
type Codec struct {
	// Target bitrate in bps.
	BitRate int

	// Quolity of the encoding [0-9].
	// Larger value results higher quality and higher CPU usage.
	// It depends on the selected codec.
	Quality int

	// Expected interval of the keyframes in frames.
	KeyFrameInterval int
}
