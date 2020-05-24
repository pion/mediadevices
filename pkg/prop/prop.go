package prop

import (
	"reflect"
	"time"

	"github.com/pion/mediadevices/pkg/frame"
)

// MediaConstraints represents set of media propaty constraints.
// Each field constrains property by min/ideal/max range, exact match, or oneof match.
type MediaConstraints struct {
	DeviceID StringConstraint
	VideoConstraints
	AudioConstraints
}

// Media stores single set of media propaties.
type Media struct {
	DeviceID string
	Video
	Audio
}

// Merge merges all the field values from o to p, except zero values.
func (p *Media) Merge(o MediaConstraints) {
	rp := reflect.ValueOf(p).Elem()
	ro := reflect.ValueOf(o)

	// merge b fields to a recursively
	var merge func(a, b reflect.Value)
	merge = func(a, b reflect.Value) {
		numFields := a.NumField()
		for i := 0; i < numFields; i++ {
			fieldA := a.Field(i)
			fieldB := b.Field(i)

			// if b is a struct, a is also a struct. Then,
			// we recursively merge them
			if fieldB.Kind() == reflect.Struct {
				merge(fieldA, fieldB)
				continue
			}

			// TODO: Replace this with fieldB.IsZero() when we move to go1.13
			// If non-boolean or non-discrete values are zeroes we skip them
			if fieldB.Interface() == reflect.Zero(fieldB.Type()).Interface() &&
				fieldB.Kind() != reflect.Bool {
				continue
			}

			switch c := fieldB.Interface().(type) {
			case IntConstraint:
				if v, ok := c.Value(); ok {
					fieldA.Set(reflect.ValueOf(v))
				}
			case FloatConstraint:
				if v, ok := c.Value(); ok {
					fieldA.Set(reflect.ValueOf(v))
				}
			case DurationConstraint:
				if v, ok := c.Value(); ok {
					fieldA.Set(reflect.ValueOf(v))
				}
			case FrameFormatConstraint:
				if v, ok := c.Value(); ok {
					fieldA.Set(reflect.ValueOf(v))
				}
			case StringConstraint:
				if v, ok := c.Value(); ok {
					fieldA.Set(reflect.ValueOf(v))
				}
			default:
				panic("unsupported property type")
			}
		}
	}

	merge(rp, ro)
}

// FitnessDistance calculates fitness of media property and media constraints.
// If no media satisfies given constraints, second return value will be false.
func (p *MediaConstraints) FitnessDistance(o Media) (float64, bool) {
	cmps := comparisons{}
	cmps.add(p.DeviceID, o.DeviceID)
	cmps.add(p.Width, o.Width)
	cmps.add(p.Height, o.Height)
	cmps.add(p.FrameFormat, o.FrameFormat)
	cmps.add(p.SampleRate, o.SampleRate)
	cmps.add(p.Latency, o.Latency)

	return cmps.fitnessDistance()
}

type comparisons []struct {
	desired, actual interface{}
}

func (c *comparisons) add(desired, actual interface{}) {
	if desired != nil {
		*c = append(*c,
			struct{ desired, actual interface{} }{
				desired, actual,
			},
		)
	}
}

// fitnessDistance is an implementation for https://w3c.github.io/mediacapture-main/#dfn-fitness-distance
func (c *comparisons) fitnessDistance() (float64, bool) {
	var dist float64
	for _, field := range *c {
		var d float64
		var ok bool
		switch c := field.desired.(type) {
		case IntConstraint:
			if actual, typeOK := field.actual.(int); typeOK {
				d, ok = c.Compare(actual)
			} else {
				panic("wrong type of actual value")
			}
		case FloatConstraint:
			if actual, typeOK := field.actual.(float32); typeOK {
				d, ok = c.Compare(actual)
			} else {
				panic("wrong type of actual value")
			}
		case DurationConstraint:
			if actual, typeOK := field.actual.(time.Duration); typeOK {
				d, ok = c.Compare(actual)
			} else {
				panic("wrong type of actual value")
			}
		case FrameFormatConstraint:
			if actual, typeOK := field.actual.(frame.Format); typeOK {
				d, ok = c.Compare(actual)
			} else {
				panic("wrong type of actual value")
			}
		case StringConstraint:
			if actual, typeOK := field.actual.(string); typeOK {
				d, ok = c.Compare(actual)
			} else {
				panic("wrong type of actual value")
			}
		default:
			panic("unsupported constraint type")
		}
		dist += d
		if !ok {
			return 0, false
		}
	}
	return dist, true
}

// VideoConstraints represents a video's constraints
type VideoConstraints struct {
	Width, Height IntConstraint
	FrameRate     FloatConstraint
	FrameFormat   FrameFormatConstraint
}

// Video represents a video's constraints
type Video struct {
	Width, Height int
	FrameRate     float32
	FrameFormat   frame.Format
}

// AudioConstraints represents an audio's constraints
type AudioConstraints struct {
	ChannelCount IntConstraint
	Latency      DurationConstraint
	SampleRate   IntConstraint
	SampleSize   IntConstraint
}

// Audio represents an audio's constraints
type Audio struct {
	ChannelCount int
	Latency      time.Duration
	SampleRate   int
	SampleSize   int
}
