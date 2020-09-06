package wave

// Int16Sample is a 16-bits signed integer audio sample.
type Int16Sample int16

func (s Int16Sample) Int() int64 {
	return int64(s) << 16
}

// Int16Interleaved multi-channel interlaced Audio.
type Int16Interleaved struct {
	Data []uint8
	Size ChunkInfo
}

// ChunkInfo returns audio chunk size.
func (a *Int16Interleaved) ChunkInfo() ChunkInfo {
	return a.Size
}

func (a *Int16Interleaved) SampleFormat() SampleFormat {
	return Int16SampleFormat
}

func (a *Int16Interleaved) At(i, ch int) Sample {
	loc := 2 * (i*a.Size.Channels + ch)

	var s Int16Sample
	s |= Int16Sample(a.Data[loc]) << 8
	s |= Int16Sample(a.Data[loc+1])
	return s
}

func (a *Int16Interleaved) Set(i, ch int, s Sample) {
	a.SetInt16(i, ch, Int16SampleFormat.Convert(s).(Int16Sample))
}

func (a *Int16Interleaved) SetInt16(i, ch int, s Int16Sample) {
	loc := 2 * (i*a.Size.Channels + ch)
	a.Data[loc] = uint8(s >> 8)
	a.Data[loc+1] = uint8(s)
}

// SubAudio returns part of the original audio sharing the buffer.
func (a *Int16Interleaved) SubAudio(offsetSamples, nSamples int) *Int16Interleaved {
	ret := *a
	offset := 2 * offsetSamples * a.Size.Channels
	n := 2 * nSamples * a.Size.Channels
	ret.Data = ret.Data[offset : offset+n]
	ret.Size.Len = nSamples
	return &ret
}

func NewInt16Interleaved(size ChunkInfo) *Int16Interleaved {
	return &Int16Interleaved{
		Data: make([]uint8, size.Channels*size.Len*2),
		Size: size,
	}
}

// Int16NonInterleaved multi-channel interlaced Audio.
type Int16NonInterleaved struct {
	Data [][]uint8
	Size ChunkInfo
}

// ChunkInfo returns audio chunk size.
func (a *Int16NonInterleaved) ChunkInfo() ChunkInfo {
	return a.Size
}

func (a *Int16NonInterleaved) SampleFormat() SampleFormat {
	return Int16SampleFormat
}

func (a *Int16NonInterleaved) At(i, ch int) Sample {
	loc := i * 2

	var s Int16Sample
	s |= Int16Sample(a.Data[ch][loc]) << 8
	s |= Int16Sample(a.Data[ch][loc+1])
	return s
}

func (a *Int16NonInterleaved) Set(i, ch int, s Sample) {
	a.SetInt16(i, ch, Int16SampleFormat.Convert(s).(Int16Sample))
}

func (a *Int16NonInterleaved) SetInt16(i, ch int, s Int16Sample) {
	loc := i * 2
	a.Data[ch][loc] = uint8(s >> 8)
	a.Data[ch][loc+1] = uint8(s)
}

// SubAudio returns part of the original audio sharing the buffer.
func (a *Int16NonInterleaved) SubAudio(offsetSamples, nSamples int) *Int16NonInterleaved {
	ret := *a
	ret.Size.Len = nSamples

	nSamples *= 2
	offsetSamples *= 2
	for i := range a.Data {
		ret.Data[i] = ret.Data[i][offsetSamples : offsetSamples+nSamples]
	}
	return &ret
}

func NewInt16NonInterleaved(size ChunkInfo) *Int16NonInterleaved {
	d := make([][]uint8, size.Channels)
	for i := 0; i < size.Channels; i++ {
		d[i] = make([]uint8, size.Len*2)
	}
	return &Int16NonInterleaved{
		Data: d,
		Size: size,
	}
}
