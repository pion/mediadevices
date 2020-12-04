package wave

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"unsafe"
)

// Format represents how audio is formatted in memory
type Format fmt.Stringer

type RawFormat struct {
	SampleSize  int
	IsFloat     bool
	Interleaved bool
}

func (f *RawFormat) String() string {
	sampleSizeInBits := f.SampleSize * 8
	dataTypeStr := "Int"
	if f.IsFloat {
		dataTypeStr = "Float"
	}
	interleavedStr := "NonInterleaved"
	if f.Interleaved {
		interleavedStr = "Interleaved"
	}

	return fmt.Sprintf("%s%d%s", dataTypeStr, sampleSizeInBits, interleavedStr)
}

var hostEndian binary.ByteOrder
var registeredDecoders = map[string]Decoder{}

func init() {
	switch v := *(*uint16)(unsafe.Pointer(&([]byte{0x12, 0x34}[0]))); v {
	case 0x1234:
		hostEndian = binary.BigEndian
	case 0x3412:
		hostEndian = binary.LittleEndian
	default:
		panic(fmt.Sprintf("failed to determine host endianness: %x", v))
	}

	decoderBuilders := []DecoderBuilderFunc{
		newInt16InterleavedDecoder,
		newInt16NonInterleavedDecoder,
		newFloat32InterleavedDecoder,
		newFloat32NonInterleavedDecoder,
	}

	for _, decoderBuilder := range decoderBuilders {
		err := RegisterDecoder(decoderBuilder)
		if err != nil {
			panic(err)
		}
	}
}

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

// DecoderBuilder builds raw audio decoder
type DecoderBuilder interface {
	// NewDecoder creates a new decoder for specified format
	NewDecoder() (Decoder, Format)
}

// DecoderBuilderFunc is a proxy type for DecoderBuilder
type DecoderBuilderFunc func() (Decoder, Format)

func (builderFunc DecoderBuilderFunc) NewDecoder() (Decoder, Format) {
	return builderFunc()
}

func RegisterDecoder(builder DecoderBuilder) error {
	decoder, format := builder.NewDecoder()
	formatStr := format.String()
	if _, ok := registeredDecoders[formatStr]; ok {
		return fmt.Errorf("%v has already been registered", format)
	}

	registeredDecoders[formatStr] = decoder
	return nil
}

// NewDecoder creates a decoder to decode raw audio data in the given format
func NewDecoder(format Format) (Decoder, error) {
	decoder, ok := registeredDecoders[format.String()]
	if !ok {
		return nil, fmt.Errorf("%s format is not supported", format)
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

func newInt16InterleavedDecoder() (Decoder, Format) {
	format := &RawFormat{
		SampleSize:  2,
		IsFloat:     false,
		Interleaved: true,
	}

	decoder := DecoderFunc(func(endian binary.ByteOrder, chunk []byte, channels int) (Audio, error) {
		sampleSize := format.SampleSize
		chunkInfo, err := calculateChunkInfo(chunk, channels, sampleSize)
		if err != nil {
			return nil, err
		}

		container := NewInt16Interleaved(chunkInfo)

		if endian == hostEndian {
			n := len(chunk)
			h := reflect.SliceHeader{Data: uintptr(unsafe.Pointer(&container.Data[0])), Len: n, Cap: n}
			dst := *(*[]byte)(unsafe.Pointer(&h))
			copy(dst, chunk)
			return container, nil
		}

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

	})

	return decoder, format
}

func newInt16NonInterleavedDecoder() (Decoder, Format) {
	format := &RawFormat{
		SampleSize:  2,
		IsFloat:     false,
		Interleaved: false,
	}

	decoder := DecoderFunc(func(endian binary.ByteOrder, chunk []byte, channels int) (Audio, error) {
		sampleSize := format.SampleSize
		chunkInfo, err := calculateChunkInfo(chunk, channels, sampleSize)
		if err != nil {
			return nil, err
		}

		container := NewInt16NonInterleaved(chunkInfo)
		chunkLen := len(chunk) / channels

		if endian == hostEndian {
			for ch := 0; ch < channels; ch++ {
				offset := ch * chunkLen
				h := reflect.SliceHeader{Data: uintptr(unsafe.Pointer(&container.Data[ch][0])), Len: chunkLen, Cap: chunkLen}
				dst := *(*[]byte)(unsafe.Pointer(&h))
				copy(dst, chunk[offset:offset+chunkLen])
			}
			return container, nil
		}

		for ch := 0; ch < channels; ch++ {
			offset := ch * chunkLen
			for i := 0; i < chunkInfo.Len; i++ {
				flatOffset := offset + i*sampleSize
				sample := endian.Uint16(chunk[flatOffset : flatOffset+sampleSize])
				container.SetInt16(i, ch, Int16Sample(sample))
			}
		}

		return container, nil
	})

	return decoder, format
}

func newFloat32InterleavedDecoder() (Decoder, Format) {
	format := &RawFormat{
		SampleSize:  4,
		IsFloat:     true,
		Interleaved: true,
	}

	decoder := DecoderFunc(func(endian binary.ByteOrder, chunk []byte, channels int) (Audio, error) {
		sampleSize := format.SampleSize
		chunkInfo, err := calculateChunkInfo(chunk, channels, sampleSize)
		if err != nil {
			return nil, err
		}

		container := NewFloat32Interleaved(chunkInfo)

		if endian == hostEndian {
			n := len(chunk)
			h := reflect.SliceHeader{Data: uintptr(unsafe.Pointer(&container.Data[0])), Len: n, Cap: n}
			dst := *(*[]byte)(unsafe.Pointer(&h))
			copy(dst, chunk)
			return container, nil
		}

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
	})

	return decoder, format
}

func newFloat32NonInterleavedDecoder() (Decoder, Format) {
	format := &RawFormat{
		SampleSize:  4,
		IsFloat:     true,
		Interleaved: false,
	}

	decoder := DecoderFunc(func(endian binary.ByteOrder, chunk []byte, channels int) (Audio, error) {
		sampleSize := format.SampleSize
		chunkInfo, err := calculateChunkInfo(chunk, channels, sampleSize)
		if err != nil {
			return nil, err
		}

		container := NewFloat32NonInterleaved(chunkInfo)
		chunkLen := len(chunk) / channels

		if endian == hostEndian {
			for ch := 0; ch < channels; ch++ {
				offset := ch * chunkLen
				h := reflect.SliceHeader{Data: uintptr(unsafe.Pointer(&container.Data[ch][0])), Len: chunkLen, Cap: chunkLen}
				dst := *(*[]byte)(unsafe.Pointer(&h))
				copy(dst, chunk[offset:offset+chunkLen])
			}
			return container, nil
		}

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
	})

	return decoder, format
}
