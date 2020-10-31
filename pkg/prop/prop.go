package prop

import (
	"fmt"
	"reflect"
	"strings"
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

func (m *MediaConstraints) String() string {
	return prettifyStruct(m)
}

// Media stores single set of media propaties.
type Media struct {
	DeviceID string
	Video
	Audio
}

func (m *Media) String() string {
	return prettifyStruct(m)
}

func prettifyStruct(i interface{}) string {
	var rows []string
	var addRows func(int, reflect.Value)
	addRows = func(level int, obj reflect.Value) {
		typeOf := obj.Type()
		for i := 0; i < obj.NumField(); i++ {
			field := typeOf.Field(i)
			value := obj.Field(i)

			padding := strings.Repeat("  ", level)
			switch value.Kind() {
			case reflect.Struct:
				rows = append(rows, fmt.Sprintf("%s%v:", padding, field.Name))
				addRows(level+1, value)
			case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
				if value.IsNil() {
					rows = append(rows, fmt.Sprintf("%s%v: any", padding, field.Name))
				} else {
					rows = append(rows, fmt.Sprintf("%s%v: %v", padding, field.Name, value))
				}
			default:
				rows = append(rows, fmt.Sprintf("%s%v: %v", padding, field.Name, value))
			}
		}
	}

	addRows(0, reflect.ValueOf(i).Elem())
	return strings.Join(rows, "\n")
}

// setterFn is a callback function to set value from fieldB to fieldA
type setterFn func(fieldA, fieldB reflect.Value)

// merge merges all the field values from o to p, except zero values. It's guaranteed that setterFn will be called
// when fieldA and fieldB are not struct.
func (p *Media) merge(o interface{}, set setterFn) {
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

			set(fieldA, fieldB)
		}
	}

	merge(rp, ro)
}

func (p *Media) Merge(o Media) {
	p.merge(o, func(fieldA, fieldB reflect.Value) {
		fieldA.Set(fieldB)
	})
}

func (p *Media) MergeConstraints(o MediaConstraints) {
	p.merge(o, func(fieldA, fieldB reflect.Value) {
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
		case BoolConstraint:
			fieldA.Set(reflect.ValueOf(c.Value()))
		default:
			panic("unsupported property type")
		}
	})
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
	cmps.add(p.ChannelCount, o.ChannelCount)
	cmps.add(p.IsBigEndian, o.IsBigEndian)
	cmps.add(p.IsFloat, o.IsFloat)
	cmps.add(p.IsInterleaved, o.IsInterleaved)

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
		case BoolConstraint:
			if actual, typeOK := field.actual.(bool); typeOK {
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
	ChannelCount  IntConstraint
	Latency       DurationConstraint
	SampleRate    IntConstraint
	SampleSize    IntConstraint
	IsBigEndian   BoolConstraint
	IsFloat       BoolConstraint
	IsInterleaved BoolConstraint
}

// Audio represents an audio's constraints
type Audio struct {
	ChannelCount  int
	Latency       time.Duration
	SampleRate    int
	SampleSize    int
	IsBigEndian   bool
	IsFloat       bool
	IsInterleaved bool
}
