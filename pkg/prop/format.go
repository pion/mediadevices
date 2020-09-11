package prop

import (
	"fmt"
	"github.com/pion/mediadevices/pkg/frame"
	"strings"
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

// String implements Stringify
func (f FrameFormat) String() string {
	return fmt.Sprintf("%s (ideal)", frame.Format(f))
}

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

// String implements Stringify
func (f FrameFormatExact) String() string {
	return fmt.Sprintf("%s (exact)", frame.Format(f))
}

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

// String implements Stringify
func (f FrameFormatOneOf) String() string {
	var opts []string
	for _, v := range f {
		opts = append(opts, fmt.Sprint(v))
	}

	return fmt.Sprintf("%s (one of values)", strings.Join(opts, ","))
}
