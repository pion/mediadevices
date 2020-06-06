package wave

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Format represents how audio is formatted in memory
type Format string

const (
	FormatInt16Interleaved      Format = "Int16Interleaved"
	FormatInt16NonInterleaved          = "Int16NonInterleaved"
	FormatFloat32Interleaved           = "Float32Interleaved"
	FormatFloat32NonInterleaved        = "Float32NonInterleaved"
)

// Decoder decodes raw chunk to Audio
type Decoder interface {
	// Decode decodes raw chunk in endian byte order
	Decode(endian binary.ByteOrder, chunk []byte, channels int) (Audio, error)
}

// DecoderFunc is a proxy type for Decoder
type DecoderFunc func(endian binary.ByteOrder, chunk []byte, channels int) (Audio, error)

func (f DecoderFunc) Decode(endian binary.ByteOrder, chunk []byte, channels int) (Audio, error) {
	return f(endian, chunk, channels)
}

// NewDecoder creates a decoder to decode raw audio data in the given format
func NewDecoder(f Format) (Decoder, error) {
	var decoder DecoderFunc

	switch f {
	case FormatInt16Interleaved:
		decoder = decodeInt16Interleaved
	case FormatInt16NonInterleaved:
		decoder = decodeInt16NonInterleaved
	case FormatFloat32Interleaved:
		decoder = decodeFloat32Interleaved
	case FormatFloat32NonInterleaved:
		decoder = decodeFloat32NonInterleaved
	default:
		return nil, fmt.Errorf("%s is not supported", f)
	}

	return decoder, nil
}

func calculateChunkInfo(chunk []byte, channels int, sampleSize int) (ChunkInfo, error) {
	if channels <= 0 {
		return ChunkInfo{}, fmt.Errorf("channels has to be greater than 0")
	}

	if sampleSize <= 0 {
		return ChunkInfo{}, fmt.Errorf("sample size has to be greater than 0")
	}

	sampleLen := channels * sampleSize
	if len(chunk)%sampleLen != 0 {
		expectedLen := len(chunk) + (sampleLen - len(chunk)%sampleLen)
		return ChunkInfo{}, fmt.Errorf("expected chunk to have a length of %d, but got %d", expectedLen, len(chunk))
	}

	return ChunkInfo{
		Channels: channels,
		Len:      len(chunk) / (channels * sampleSize),
	}, nil
}

func decodeInt16Interleaved(endian binary.ByteOrder, chunk []byte, channels int) (Audio, error) {
	sampleSize := 2
	chunkInfo, err := calculateChunkInfo(chunk, channels, sampleSize)
	if err != nil {
		return nil, err
	}

	container := NewInt16Interleaved(chunkInfo)
	sampleLen := sampleSize * channels
	var i int
	for offset := 0; offset+sampleLen <= len(chunk); offset += sampleLen {
		for ch := 0; ch < channels; ch++ {
			flatOffset := offset + ch*sampleSize
			sample := endian.Uint16(chunk[flatOffset : flatOffset+sampleSize])
			container.SetInt16(i, ch, Int16Sample(sample))
		}
		i++
	}

	return container, nil
}

func decodeInt16NonInterleaved(endian binary.ByteOrder, chunk []byte, channels int) (Audio, error) {
	sampleSize := 2
	chunkInfo, err := calculateChunkInfo(chunk, channels, sampleSize)
	if err != nil {
		return nil, err
	}

	container := NewInt16NonInterleaved(chunkInfo)
	chunkLen := len(chunk) / channels
	for ch := 0; ch < channels; ch++ {
		offset := ch * chunkLen
		for i := 0; i < chunkInfo.Len; i++ {
			flatOffset := offset + i*sampleSize
			sample := endian.Uint16(chunk[flatOffset : flatOffset+sampleSize])
			container.SetInt16(i, ch, Int16Sample(sample))
		}
	}

	return container, nil
}

func decodeFloat32Interleaved(endian binary.ByteOrder, chunk []byte, channels int) (Audio, error) {
	sampleSize := 4
	chunkInfo, err := calculateChunkInfo(chunk, channels, sampleSize)
	if err != nil {
		return nil, err
	}

	container := NewFloat32Interleaved(chunkInfo)
	sampleLen := sampleSize * channels
	var i int
	for offset := 0; offset+sampleLen <= len(chunk); offset += sampleLen {
		for ch := 0; ch < channels; ch++ {
			flatOffset := offset + ch*sampleSize
			sample := endian.Uint32(chunk[flatOffset : flatOffset+sampleSize])
			sampleF := math.Float32frombits(sample)
			container.SetFloat32(i, ch, Float32Sample(sampleF))
		}
		i++
	}

	return container, nil
}

func decodeFloat32NonInterleaved(endian binary.ByteOrder, chunk []byte, channels int) (Audio, error) {
	sampleSize := 4
	chunkInfo, err := calculateChunkInfo(chunk, channels, sampleSize)
	if err != nil {
		return nil, err
	}

	container := NewFloat32NonInterleaved(chunkInfo)
	chunkLen := len(chunk) / channels
	for ch := 0; ch < channels; ch++ {
		offset := ch * chunkLen
		for i := 0; i < chunkInfo.Len; i++ {
			flatOffset := offset + i*sampleSize
			sample := endian.Uint32(chunk[flatOffset : flatOffset+sampleSize])
			sampleF := math.Float32frombits(sample)
			container.SetFloat32(i, ch, Float32Sample(sampleF))
		}
	}

	return container, nil
}
