package prop

import (
	"math"
)

type IntConstraint interface {
	Compare(int) (float64, bool)
	Value() (int, bool)
}

type Int int

func (i Int) Compare(a int) (float64, bool) {
	return math.Abs(float64(a-int(i))) / math.Max(math.Abs(float64(a)), math.Abs(float64(i))), true
}

func (i Int) Value() (int, bool) { return int(i), true }

type IntExact int

func (i IntExact) Compare(a int) (float64, bool) {
	if int(i) == a {
		return 0.0, true
	}
	return 1.0, false
}

func (i IntExact) Value() (int, bool) { return int(i), true }

type IntOneOf []int

func (i IntOneOf) Compare(a int) (float64, bool) {
	for _, ii := range i {
		if ii == a {
			return 0.0, true
		}
	}
	return 1.0, false
}

func (IntOneOf) Value() (int, bool) { return 0, false }

type IntRanged struct {
	Min   int
	Max   int
	Ideal int
}

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

func (IntRanged) Value() (int, bool) { return 0, false }
