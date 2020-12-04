package prop

import (
	"fmt"
	"math"
	"strings"
)

// IntConstraint is an interface to represent integer value constraint.
type IntConstraint interface {
	Compare(int) (float64, bool)
	Value() (int, bool)
}

// Int specifies ideal int value.
// Any value may be selected, but closest value takes priority.
type Int int

// Compare implements IntConstraint.
func (i Int) Compare(a int) (float64, bool) {
	return math.Abs(float64(a-int(i))) / math.Max(math.Abs(float64(a)), math.Abs(float64(i))), true
}

// Value implements IntConstraint.
func (i Int) Value() (int, bool) { return int(i), true }

// String implements Stringify
func (i Int) String() string {
	return fmt.Sprintf("%d (ideal)", i)
}

// IntExact specifies exact int value.
type IntExact int

// Compare implements IntConstraint.
func (i IntExact) Compare(a int) (float64, bool) {
	if int(i) == a {
		return 0.0, true
	}
	return 1.0, false
}

// String implements Stringify
func (i IntExact) String() string {
	return fmt.Sprintf("%d (exact)", i)
}

// Value implements IntConstraint.
func (i IntExact) Value() (int, bool) { return int(i), true }

// IntOneOf specifies list of expected float values.
type IntOneOf []int

// Compare implements IntConstraint.
func (i IntOneOf) Compare(a int) (float64, bool) {
	for _, ii := range i {
		if ii == a {
			return 0.0, true
		}
	}
	return 1.0, false
}

// Value implements IntConstraint.
func (IntOneOf) Value() (int, bool) { return 0, false }

// String implements Stringify
func (i IntOneOf) String() string {
	var opts []string
	for _, v := range i {
		opts = append(opts, fmt.Sprint(v))
	}

	return fmt.Sprintf("%s (one of values)", strings.Join(opts, ","))
}

// IntRanged specifies range of expected int value.
// If Ideal is non-zero, closest value to Ideal takes priority.
type IntRanged struct {
	Min   int
	Max   int
	Ideal int
}

// Compare implements IntConstraint.
func (i IntRanged) Compare(a int) (float64, bool) {
	if i.Min != 0 && i.Min > a {
		// Out of range
		return 1.0, false
	}
	if i.Max != 0 && i.Max < a {
		// Out of range
		return 1.0, false
	}
	if i.Ideal == 0 {
		// If the value is in the range and Ideal is not specified,
		// any value is evenly acceptable.
		return 0.0, true
	}
	switch {
	case a == i.Ideal:
		return 0.0, true
	case a < i.Ideal:
		if i.Min == 0 {
			// If Min is not specified, smaller values than Ideal are even.
			return 0.0, true
		}
		return float64(i.Ideal-a) / float64(i.Ideal-i.Min), true
	default:
		if i.Max == 0 {
			// If Max is not specified, larger values than Ideal are even.
			return 0.0, true
		}
		return float64(a-i.Ideal) / float64(i.Max-i.Ideal), true
	}
}

// Value implements IntConstraint.
func (IntRanged) Value() (int, bool) { return 0, false }

// String implements Stringify
func (i IntRanged) String() string {
	return fmt.Sprintf("%d - %d (range), %d (ideal)", i.Min, i.Max, i.Ideal)
}
