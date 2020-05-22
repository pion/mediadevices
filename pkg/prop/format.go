package prop

import (
	"github.com/pion/mediadevices/pkg/frame"
)

// FrameFormatConstraint is an interface to represent frame format constraint.
type FrameFormatConstraint interface {
	Compare(frame.Format) (float64, bool)
	Value() (frame.Format, bool)
}

// FrameFormat specifies expected frame format.
// Any value may be selected, but matched value takes priority.
type FrameFormat frame.Format

// Compare implements FrameFormatConstraint.
func (f FrameFormat) Compare(a frame.Format) (float64, bool) {
	if frame.Format(f) == a {
		return 0.0, true
	}
	return 1.0, true
}

// Value implements FrameFormatConstraint.
func (f FrameFormat) Value() (frame.Format, bool) { return frame.Format(f), true }

// FrameFormatExact specifies exact frame format.
type FrameFormatExact frame.Format

// Compare implements FrameFormatConstraint.
func (f FrameFormatExact) Compare(a frame.Format) (float64, bool) {
	if frame.Format(f) == a {
		return 0.0, true
	}
	return 1.0, false
}

// Value implements FrameFormatConstraint.
func (f FrameFormatExact) Value() (frame.Format, bool) { return frame.Format(f), true }

// FrameFormatOneOf specifies list of expected frame format.
type FrameFormatOneOf []frame.Format

// Compare implements FrameFormatConstraint.
func (f FrameFormatOneOf) Compare(a frame.Format) (float64, bool) {
	for _, ff := range f {
		if ff == a {
			return 0.0, true
		}
	}
	return 1.0, false
}

// Value implements FrameFormatConstraint.
func (FrameFormatOneOf) Value() (frame.Format, bool) { return "", false }
