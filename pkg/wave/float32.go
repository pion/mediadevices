package wave

import "math"

// Float32Sample is a 32-bits float audio sample.
type Float32Sample float32

func (s Float32Sample) Int() int64 {
	return int64(s * 0x100000000)
}

// Float32Interleaved multi-channel interlaced Audio.
type Float32Interleaved struct {
	Data []uint8
	Size ChunkInfo
}

// ChunkInfo returns audio chunk size.
func (a *Float32Interleaved) ChunkInfo() ChunkInfo {
	return a.Size
}

func (a *Float32Interleaved) SampleFormat() SampleFormat {
	return Float32SampleFormat
}

func (a *Float32Interleaved) At(i, ch int) Sample {
	loc := 4 * (a.Size.Channels*i + ch)

	var v uint32
	v |= uint32(a.Data[loc]) << 24
	v |= uint32(a.Data[loc+1]) << 16
	v |= uint32(a.Data[loc+2]) << 8
	v |= uint32(a.Data[loc+3])

	return Float32Sample(math.Float32frombits(v))
}

func (a *Float32Interleaved) Set(i, ch int, s Sample) {
	a.SetFloat32(i, ch, Float32SampleFormat.Convert(s).(Float32Sample))
}

func (a *Float32Interleaved) SetFloat32(i, ch int, s Float32Sample) {
	loc := 4 * (a.Size.Channels*i + ch)

	v := math.Float32bits(float32(s))
	a.Data[loc] = uint8(v >> 24)
	a.Data[loc+1] = uint8(v >> 16)
	a.Data[loc+2] = uint8(v >> 8)
	a.Data[loc+3] = uint8(v)
}

// SubAudio returns part of the original audio sharing the buffer.
func (a *Float32Interleaved) SubAudio(offsetSamples, nSamples int) *Float32Interleaved {
	ret := *a
	offset := 4 * offsetSamples * a.Size.Channels
	n := 4 * nSamples * a.Size.Channels
	ret.Data = ret.Data[offset : offset+n]
	ret.Size.Len = nSamples
	return &ret
}

func NewFloat32Interleaved(size ChunkInfo) *Float32Interleaved {
	return &Float32Interleaved{
		Data: make([]uint8, size.Channels*size.Len*4),
		Size: size,
	}
}

// Float32NonInterleaved multi-channel interlaced Audio.
type Float32NonInterleaved struct {
	Data [][]uint8
	Size ChunkInfo
}

// ChunkInfo returns audio chunk size.
func (a *Float32NonInterleaved) ChunkInfo() ChunkInfo {
	return a.Size
}

func (a *Float32NonInterleaved) SampleFormat() SampleFormat {
	return Float32SampleFormat
}

func (a *Float32NonInterleaved) At(i, ch int) Sample {
	loc := i * 4

	var v uint32
	v |= uint32(a.Data[ch][loc]) << 24
	v |= uint32(a.Data[ch][loc+1]) << 16
	v |= uint32(a.Data[ch][loc+2]) << 8
	v |= uint32(a.Data[ch][loc+3])

	return Float32Sample(math.Float32frombits(v))
}

func (a *Float32NonInterleaved) Set(i, ch int, s Sample) {
	a.SetFloat32(i, ch, Float32SampleFormat.Convert(s).(Float32Sample))
}

func (a *Float32NonInterleaved) SetFloat32(i, ch int, s Float32Sample) {
	loc := i * 4

	v := math.Float32bits(float32(s))
	a.Data[ch][loc] = uint8(v >> 24)
	a.Data[ch][loc+1] = uint8(v >> 16)
	a.Data[ch][loc+2] = uint8(v >> 8)
	a.Data[ch][loc+3] = uint8(v)
}

// SubAudio returns part of the original audio sharing the buffer.
func (a *Float32NonInterleaved) SubAudio(offsetSamples, nSamples int) *Float32NonInterleaved {
	ret := *a
	ret.Size.Len = nSamples

	offsetSamples *= 4
	nSamples *= 4
	for i := range a.Data {
		ret.Data[i] = ret.Data[i][offsetSamples : offsetSamples+nSamples]
	}
	return &ret
}

func NewFloat32NonInterleaved(size ChunkInfo) *Float32NonInterleaved {
	d := make([][]uint8, size.Channels)
	for i := 0; i < size.Channels; i++ {
		d[i] = make([]uint8, size.Len*4)
	}
	return &Float32NonInterleaved{
		Data: d,
		Size: size,
	}
}
