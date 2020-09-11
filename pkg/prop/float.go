package prop

import (
	"fmt"
	"math"
	"strings"
)

// FloatConstraint is an interface to represent float value constraint.
type FloatConstraint interface {
	Compare(float32) (float64, bool)
	Value() (float32, bool)
}

// Float specifies ideal float value.
// Any value may be selected, but closest value takes priority.
type Float float32

// Compare implements FloatConstraint.
func (f Float) Compare(a float32) (float64, bool) {
	return math.Abs(float64(a-float32(f))) / math.Max(math.Abs(float64(a)), math.Abs(float64(f))), true
}

// Value implements FloatConstraint.
func (f Float) Value() (float32, bool) { return float32(f), true }

// String implements Stringify
func (f Float) String() string {
	return fmt.Sprintf("%.2f (ideal)", f)
}

// FloatExact specifies exact float value.
type FloatExact float32

// Compare implements FloatConstraint.
func (f FloatExact) Compare(a float32) (float64, bool) {
	if float32(f) == a {
		return 0.0, true
	}
	return 1.0, false
}

// Value implements FloatConstraint.
func (f FloatExact) Value() (float32, bool) { return float32(f), true }

// String implements Stringify
func (f FloatExact) String() string {
	return fmt.Sprintf("%.2f (exact)", f)
}

// FloatOneOf specifies list of expected float values.
type FloatOneOf []float32

// Compare implements FloatConstraint.
func (f FloatOneOf) Compare(a float32) (float64, bool) {
	for _, ff := range f {
		if ff == a {
			return 0.0, true
		}
	}
	return 1.0, false
}

// Value implements FloatConstraint.
func (FloatOneOf) Value() (float32, bool) { return 0, false }

// String implements Stringify
func (f FloatOneOf) String() string {
	var opts []string
	for _, v := range f {
		opts = append(opts, fmt.Sprintf("%.2f", v))
	}

	return fmt.Sprintf("%s (one of values)", strings.Join(opts, ","))
}

// FloatRanged specifies range of expected float value.
// If Ideal is non-zero, closest value to Ideal takes priority.
type FloatRanged struct {
	Min   float32
	Max   float32
	Ideal float32
}

// Compare implements FloatConstraint.
func (f FloatRanged) Compare(a float32) (float64, bool) {
	if f.Min != 0 && f.Min > a {
		// Out of range
		return 1.0, false
	}
	if f.Max != 0 && f.Max < a {
		// Out of range
		return 1.0, false
	}
	if f.Ideal == 0 {
		// If the value is in the range and Ideal is not specified,
		// any value is evenly acceptable.
		return 0.0, true
	}
	switch {
	case a == f.Ideal:
		return 0.0, true
	case a < f.Ideal:
		if f.Min == 0 {
			// If Min is not specified, smaller values than Ideal are even.
			return 0.0, true
		}
		return float64(f.Ideal-a) / float64(f.Ideal-f.Min), true
	default:
		if f.Max == 0 {
			// If Max is not specified, larger values than Ideal are even.
			return 0.0, true
		}
		return float64(a-f.Ideal) / float64(f.Max-f.Ideal), true
	}
}

// Value implements FloatConstraint.
func (FloatRanged) Value() (float32, bool) { return 0, false }

// String implements Stringify
func (f FloatRanged) String() string {
	return fmt.Sprintf("%.2f - %.2f (range), %.2f (ideal)", f.Min, f.Max, f.Ideal)
}
