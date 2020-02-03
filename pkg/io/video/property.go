package video

import "github.com/pion/mediadevices/pkg/frame"

// Property represents a video's basic properties
type Property struct {
	Width, Height int
	FrameRate     float32
}

// AdvancedProperty represents a video's advanced properties.
type AdvancedProperty struct {
	Property
	FrameFormat frame.Format
	BitRate     int
}
