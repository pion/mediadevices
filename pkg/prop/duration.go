package prop

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// DurationConstraint is an interface to represent time.Duration constraint.
type DurationConstraint interface {
	Compare(time.Duration) (float64, bool)
	Value() (time.Duration, bool)
}

// Duration specifies ideal duration value.
// Any value may be selected, but closest value takes priority.
type Duration time.Duration

// Compare implements DurationConstraint.
func (d Duration) Compare(a time.Duration) (float64, bool) {
	return math.Abs(float64(a-time.Duration(d))) / math.Max(math.Abs(float64(a)), math.Abs(float64(d))), true
}

// Value implements DurationConstraint.
func (d Duration) Value() (time.Duration, bool) { return time.Duration(d), true }

// String implements Stringify
func (d Duration) String() string {
	return fmt.Sprintf("%v (ideal)", time.Duration(d))
}

// DurationExact specifies exact duration value.
type DurationExact time.Duration

// Compare implements DurationConstraint.
func (d DurationExact) Compare(a time.Duration) (float64, bool) {
	if time.Duration(d) == a {
		return 0.0, true
	}
	return 1.0, false
}

// Value implements DurationConstraint.
func (d DurationExact) Value() (time.Duration, bool) { return time.Duration(d), true }

// String implements Stringify
func (d DurationExact) String() string {
	return fmt.Sprintf("%v (exact)", time.Duration(d))
}

// DurationOneOf specifies list of expected duration values.
type DurationOneOf []time.Duration

// Compare implements DurationConstraint.
func (d DurationOneOf) Compare(a time.Duration) (float64, bool) {
	for _, ii := range d {
		if ii == a {
			return 0.0, true
		}
	}
	return 1.0, false
}

// Value implements DurationConstraint.
func (DurationOneOf) Value() (time.Duration, bool) { return 0, false }

// String implements Stringify
func (d DurationOneOf) String() string {
	var opts []string
	for _, v := range d {
		opts = append(opts, fmt.Sprint(v))
	}

	return fmt.Sprintf("%s (one of values)", strings.Join(opts, ","))
}

// DurationRanged specifies range of expected duration value.
// If Ideal is non-zero, closest value to Ideal takes priority.
type DurationRanged struct {
	Min   time.Duration
	Max   time.Duration
	Ideal time.Duration
}

// Compare implements DurationConstraint.
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

// Value implements DurationConstraint.
func (DurationRanged) Value() (time.Duration, bool) { return 0, false }

// String implements Stringify
func (d DurationRanged) String() string {
	return fmt.Sprintf("%s - %s (range), %s (ideal)", d.Min, d.Max, d.Ideal)
}
