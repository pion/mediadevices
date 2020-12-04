package wave

// Float32Sample is a 32-bits float audio sample.
type Float32Sample float32

func (s Float32Sample) Int() int64 {
	return int64(s * 0x100000000)
}

// Float32Interleaved multi-channel interlaced Audio.
type Float32Interleaved struct {
	Data []float32
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
	return Float32Sample(a.Data[i*a.Size.Channels+ch])
}

func (a *Float32Interleaved) Set(i, ch int, s Sample) {
	a.Data[i*a.Size.Channels+ch] = float32(Float32SampleFormat.Convert(s).(Float32Sample))
}

func (a *Float32Interleaved) SetFloat32(i, ch int, s Float32Sample) {
	a.Data[i*a.Size.Channels+ch] = float32(s)
}

// SubAudio returns part of the original audio sharing the buffer.
func (a *Float32Interleaved) SubAudio(offsetSamples, nSamples int) *Float32Interleaved {
	ret := *a
	offset := offsetSamples * a.Size.Channels
	n := nSamples * a.Size.Channels
	ret.Data = ret.Data[offset : offset+n]
	ret.Size.Len = nSamples
	return &ret
}

func NewFloat32Interleaved(size ChunkInfo) *Float32Interleaved {
	return &Float32Interleaved{
		Data: make([]float32, size.Channels*size.Len),
		Size: size,
	}
}

// Float32NonInterleaved multi-channel interlaced Audio.
type Float32NonInterleaved struct {
	Data [][]float32
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
	return Float32Sample(a.Data[ch][i])
}

func (a *Float32NonInterleaved) Set(i, ch int, s Sample) {
	a.Data[ch][i] = float32(Float32SampleFormat.Convert(s).(Float32Sample))
}

func (a *Float32NonInterleaved) SetFloat32(i, ch int, s Float32Sample) {
	a.Data[ch][i] = float32(s)
}

// SubAudio returns part of the original audio sharing the buffer.
func (a *Float32NonInterleaved) SubAudio(offsetSamples, nSamples int) *Float32NonInterleaved {
	ret := *a
	for i := range a.Data {
		ret.Data[i] = ret.Data[i][offsetSamples : offsetSamples+nSamples]
	}
	ret.Size.Len = nSamples
	return &ret
}

func NewFloat32NonInterleaved(size ChunkInfo) *Float32NonInterleaved {
	d := make([][]float32, size.Channels)
	for i := 0; i < size.Channels; i++ {
		d[i] = make([]float32, size.Len)
	}
	return &Float32NonInterleaved{
		Data: d,
		Size: size,
	}
}
