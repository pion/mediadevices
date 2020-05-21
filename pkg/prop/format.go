package prop

import (
	"github.com/pion/mediadevices/pkg/frame"
)

type FrameFormatConstraint interface {
	Compare(frame.Format) (float64, bool)
	Value() (frame.Format, bool)
}

type FrameFormat frame.Format

func (f FrameFormat) Compare(a frame.Format) (float64, bool) {
	if frame.Format(f) == a {
		return 0.0, true
	}
	return 1.0, true
}

func (f FrameFormat) Value() (frame.Format, bool) { return frame.Format(f), true }

type FrameFormatExact frame.Format

func (f FrameFormatExact) Compare(a frame.Format) (float64, bool) {
	if frame.Format(f) == a {
		return 0.0, true
	}
	return 1.0, false
}

func (f FrameFormatExact) Value() (frame.Format, bool) { return frame.Format(f), true }

type FrameFormatOneOf []frame.Format

func (f FrameFormatOneOf) Compare(a frame.Format) (float64, bool) {
	for _, ff := range f {
		if ff == a {
			return 0.0, true
		}
	}
	return 1.0, false
}

func (FrameFormatOneOf) Value() (frame.Format, bool) { return "", false }
