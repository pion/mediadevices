package wave

// Int16Sample is a 16-bits signed integer audio sample.
type Int16Sample int16

func (s Int16Sample) Int() int64 {
	return int64(s) << 16
}

// Int16Interleaved multi-channel interlaced Audio.
type Int16Interleaved struct {
	Data []int16
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
	return Int16Sample(a.Data[i*a.Size.Channels+ch])
}

func (a *Int16Interleaved) Set(i, ch int, s Sample) {
	a.Data[i*a.Size.Channels+ch] = int16(Int16SampleFormat.Convert(s).(Int16Sample))
}

func (a *Int16Interleaved) SetInt16(i, ch int, s Int16Sample) {
	a.Data[i*a.Size.Channels+ch] = int16(s)
}

// SubAudio returns part of the original audio sharing the buffer.
func (a *Int16Interleaved) SubAudio(offsetSamples, nSamples int) *Int16Interleaved {
	ret := *a
	offset := offsetSamples * a.Size.Channels
	n := nSamples * a.Size.Channels
	ret.Data = ret.Data[offset : offset+n]
	ret.Size.Len = nSamples
	return &ret
}

func NewInt16Interleaved(size ChunkInfo) *Int16Interleaved {
	return &Int16Interleaved{
		Data: make([]int16, size.Channels*size.Len),
		Size: size,
	}
}

// Int16NonInterleaved multi-channel interlaced Audio.
type Int16NonInterleaved struct {
	Data [][]int16
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
	return Int16Sample(a.Data[ch][i])
}

func (a *Int16NonInterleaved) Set(i, ch int, s Sample) {
	a.Data[ch][i] = int16(Int16SampleFormat.Convert(s).(Int16Sample))
}

func (a *Int16NonInterleaved) SetInt16(i, ch int, s Int16Sample) {
	a.Data[ch][i] = int16(s)
}

// SubAudio returns part of the original audio sharing the buffer.
func (a *Int16NonInterleaved) SubAudio(offsetSamples, nSamples int) *Int16NonInterleaved {
	ret := *a
	for i := range a.Data {
		ret.Data[i] = ret.Data[i][offsetSamples : offsetSamples+nSamples]
	}
	ret.Size.Len = nSamples
	return &ret
}

func NewInt16NonInterleaved(size ChunkInfo) *Int16NonInterleaved {
	d := make([][]int16, size.Channels)
	for i := 0; i < size.Channels; i++ {
		d[i] = make([]int16, size.Len)
	}
	return &Int16NonInterleaved{
		Data: d,
		Size: size,
	}
}
