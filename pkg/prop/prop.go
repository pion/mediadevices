package prop

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"

	"github.com/pion/mediadevices/pkg/frame"
)

type Media struct {
	DeviceID string
	Video
	Audio
}

// Merge merges all the field values from o to p, except zero values.
func (p *Media) Merge(o Media) {
	rp := reflect.ValueOf(p).Elem()
	ro := reflect.ValueOf(o)

	// merge b fields to a recursively
	var merge func(a, b reflect.Value)
	merge = func(a, b reflect.Value) {
		numFields := a.NumField()
		for i := 0; i < numFields; i++ {
			fieldA := a.Field(i)
			fieldB := b.Field(i)

			// if a is a struct, b is also a struct. Then,
			// we recursively merge them
			if fieldA.Kind() == reflect.Struct {
				merge(fieldA, fieldB)
				continue
			}

			// TODO: Replace this with fieldB.IsZero() when we move to go1.13
			// If non-boolean or non-discrete values are zeroes we skip them
			if fieldB.Interface() == reflect.Zero(fieldB.Type()).Interface() &&
				fieldB.Kind() != reflect.Bool {
				continue
			}

			fieldA.Set(fieldB)
		}
	}

	merge(rp, ro)
}

func (p *Media) FitnessDistance(o Media) float64 {
	cmps := comparisons{}
	cmps.add(p.Width, o.Width)
	cmps.add(p.Height, o.Height)
	cmps.add(p.FrameFormat, o.FrameFormat)
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

		actualF, err1 := strconv.ParseFloat(actual, 64)
		idealF, err2 := strconv.ParseFloat(ideal, 64)

		switch {
		// If both of the values are numeric, we need to normalize the values to get the distance
		case err1 == nil && err2 == nil:
			dist += math.Abs(actualF-idealF) / math.Max(math.Abs(actualF), math.Abs(idealF))
		// If both of the values are not numeric, the only comparison value is either 0 (matched) or 1 (not matched)
		case err1 != nil && err2 != nil:
			if actual != ideal {
				dist++
			}
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
