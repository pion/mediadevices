package prop

// StringConstraint is an interface to represent string constraint.
type StringConstraint interface {
	Compare(string) (float64, bool)
	Value() (string, bool)
}

// String specifies expected string.
// Any value may be selected, but matched value takes priority.
type String string

// Compare implements StringConstraint.
func (f String) Compare(a string) (float64, bool) {
	if string(f) == a {
		return 0.0, true
	}
	return 1.0, true
}

// Value implements StringConstraint.
func (f String) Value() (string, bool) { return string(f), true }

// StringExact specifies exact string.
type StringExact string

// Compare implements StringConstraint.
func (f StringExact) Compare(a string) (float64, bool) {
	if string(f) == a {
		return 0.0, true
	}
	return 1.0, false
}

// Value implements StringConstraint.
func (f StringExact) Value() (string, bool) { return string(f), true }

// StringOneOf specifies list of expected string.
type StringOneOf []string

// Compare implements StringConstraint.
func (f StringOneOf) Compare(a string) (float64, bool) {
	for _, ff := range f {
		if ff == a {
			return 0.0, true
		}
	}
	return 1.0, false
}

// Value implements StringConstraint.
func (StringOneOf) Value() (string, bool) { return "", false }
