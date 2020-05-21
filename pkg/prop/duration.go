package prop

import (
	"math"
	"time"
)

type DurationConstraint interface {
	Compare(time.Duration) (float64, bool)
	Value() (time.Duration, bool)
}

type Duration time.Duration

func (d Duration) Compare(a time.Duration) (float64, bool) {
	return math.Abs(float64(a-time.Duration(d))) / math.Max(math.Abs(float64(a)), math.Abs(float64(d))), true
}

func (d Duration) Value() (time.Duration, bool) { return time.Duration(d), true }

type DurationExact time.Duration

func (d DurationExact) Compare(a time.Duration) (float64, bool) {
	if time.Duration(d) == a {
		return 0.0, true
	}
	return 1.0, false
}

func (d DurationExact) Value() (time.Duration, bool) { return time.Duration(d), true }

type DurationOneOf []time.Duration

func (d DurationOneOf) Compare(a time.Duration) (float64, bool) {
	for _, ii := range d {
		if ii == a {
			return 0.0, true
		}
	}
	return 1.0, false
}

func (DurationOneOf) Value() (time.Duration, bool) { return 0, false }

type DurationRanged struct {
	Min   time.Duration
	Max   time.Duration
	Ideal time.Duration
}

func (d DurationRanged) Compare(a time.Duration) (float64, bool) {
	if d.Min != 0 && d.Min > a {
		// Out of range
		return 1.0, false
	}
	if d.Max != 0 && d.Max < a {
		// Out of range
		return 1.0, false
	}
	if d.Ideal == 0 {
		// If the value is in the range and Ideal is not specified,
		// any value is evenly acceptable.
		return 0.0, true
	}
	switch {
	case a == d.Ideal:
		return 0.0, true
	case a < d.Ideal:
		if d.Min == 0 {
			// If Min is not specified, smaller values than Ideal are even.
			return 0.0, true
		}
		return float64(d.Ideal-a) / float64(d.Ideal-d.Min), true
	default:
		if d.Max == 0 {
			// If Max is not specified, larger values than Ideal are even.
			return 0.0, true
		}
		return float64(a-d.Ideal) / float64(d.Max-d.Ideal), true
	}
}

func (DurationRanged) Value() (time.Duration, bool) { return 0, false }
