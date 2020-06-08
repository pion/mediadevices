// Package wave implements a basic audio data library.
package wave

// Audio is a finite series of audio Sample values.
type Audio interface {
	SampleFormat() SampleFormat
	ChunkInfo() ChunkInfo
	At(i, ch int) Sample
}

// EditableAudio is an editable finite series of audio Sample values.
type EditableAudio interface {
	Audio
	Set(i, ch int, s Sample)
}

// ChunkInfo contains size of the audio chunk.
type ChunkInfo struct {
	Len          int
	Channels     int
	SamplingRate int
}

// SampleFormat can convert any Sample to one from its own sample format.
type SampleFormat interface {
	Convert(c Sample) Sample
}

// SampleFormatFunc returns a SampleFormat that invokes f to implement the conversion.
func SampleFormatFunc(f func(Sample) Sample) SampleFormat {
	return &sampleFormatFunc{f}
}

type sampleFormatFunc struct {
	f func(Sample) Sample
}

func (f *sampleFormatFunc) Convert(s Sample) Sample {
	return f.f(s)
}

// SampleFormats for the standard formats.
var (
	Int16SampleFormat = SampleFormatFunc(func(s Sample) Sample {
		if _, ok := s.(Int16Sample); ok {
			return s
		}
		return Int16Sample(s.Int() >> 16)
	})
	Float32SampleFormat = SampleFormatFunc(func(s Sample) Sample {
		if _, ok := s.(Float32Sample); ok {
			return s
		}
		return Float32Sample(float32(s.Int()) / 0x100000000)
	})
)

// Sample can convert itself to 64-bits signed value.
type Sample interface {
	// Int returns the audio level value for the sample.
	// A value ranges within [0, 0xffffffff], but is represented by a int64.
	Int() int64
}
