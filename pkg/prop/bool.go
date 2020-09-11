package prop

import "fmt"

// BoolConstraint is an interface to represent bool value constraint.
type BoolConstraint interface {
	Compare(bool) (float64, bool)
	Value() bool
}

// BoolExact specifies exact bool value.
type BoolExact bool

// Compare implements BoolConstraint.
func (b BoolExact) Compare(o bool) (float64, bool) {
	if bool(b) == o {
		return 0.0, true
	}
	return 1.0, false
}

// Value implements BoolConstraint.
func (b BoolExact) Value() bool { return bool(b) }

// String implements Stringify
func (b BoolExact) String() string {
	return fmt.Sprintf("%t (exact)", b)
}

// Bool specifies ideal bool value.
type Bool BoolExact

// Compare implements BoolConstraint.
func (b Bool) Compare(o bool) (float64, bool) {
	dist, _ := BoolExact(b).Compare(o)
	return dist, true
}
