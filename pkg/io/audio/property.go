package audio

import "time"

// Property represents an audio's basic properties
type Property struct {
	ChannelCount int
	Latency      time.Duration
	SampleRate   int
	SampleSize   int
}

// AdvancedProperty represents an audio's advanced properties.
type AdvancedProperty struct {
	Property
}
