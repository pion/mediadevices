package wave

import "fmt"

var (
	errUnsupportedFormat = fmt.Errorf("Unsupported format")
)

// Buffer is a buffer that can store any audio format.
type Buffer struct {
	// TODO: Probably standardize the audio formats so that we don't need to have the following different types
	//       and duplicated codes for each type
	bufferFloat32Interleaved    []float32
	bufferFloat32NonInterleaved [][]float32
	bufferInt16Interleaved      []int16
	bufferInt16NonInterleaved   [][]int16
	tmp                         Audio
}

// NewBuffer creates a new Buffer instance
func NewBuffer() *Buffer {
	return &Buffer{}
}

// Load loads the current owned Audio
func (buff *Buffer) Load() Audio {
	return buff.tmp
}

// StoreCopy makes a copy of src and store its copy. StoreCopy will reuse as much memory as it can
// from the previous copies. For example, if StoreCopy is given an audio that has the format from the previous call,
// StoreCopy will not allocate extra memory and only copy the content from src to the previous buffer.
func (buff *Buffer) StoreCopy(src Audio) {
	switch src := src.(type) {
	case *Float32Interleaved:
		clone, ok := buff.tmp.(*Float32Interleaved)
		if ok {
			*clone = *src
		} else {
			copied := *src
			clone = &copied
		}

		neededSize := len(src.Data)
		if len(buff.bufferFloat32Interleaved) < neededSize {
			if cap(buff.bufferFloat32Interleaved) >= neededSize {
				buff.bufferFloat32Interleaved = buff.bufferFloat32Interleaved[:neededSize]
			} else {
				buff.bufferFloat32Interleaved = make([]float32, neededSize)
			}
		}

		copy(buff.bufferFloat32Interleaved, src.Data)
		clone.Data = buff.bufferFloat32Interleaved
		buff.tmp = clone

	case *Float32NonInterleaved:
		clone, ok := buff.tmp.(*Float32NonInterleaved)
		if ok {
			*clone = *src
		} else {
			copied := *src
			clone = &copied
		}

		neededSize := len(src.Data)
		if len(buff.bufferFloat32NonInterleaved) < neededSize {
			if cap(buff.bufferFloat32NonInterleaved) >= neededSize {
				buff.bufferFloat32NonInterleaved = buff.bufferFloat32NonInterleaved[:neededSize]
			} else {
				buff.bufferFloat32NonInterleaved = make([][]float32, neededSize)
			}
		}

		for i := range src.Data {
			neededSize := len(src.Data[i])
			if len(buff.bufferFloat32NonInterleaved[i]) < neededSize {
				if cap(buff.bufferFloat32NonInterleaved[i]) >= neededSize {
					buff.bufferFloat32NonInterleaved[i] = buff.bufferFloat32NonInterleaved[i][:neededSize]
				} else {
					buff.bufferFloat32NonInterleaved[i] = make([]float32, neededSize)
				}
			}

			copy(buff.bufferFloat32NonInterleaved[i], src.Data[i])
		}
		clone.Data = buff.bufferFloat32NonInterleaved
		buff.tmp = clone

	case *Int16Interleaved:
		clone, ok := buff.tmp.(*Int16Interleaved)
		if ok {
			*clone = *src
		} else {
			copied := *src
			clone = &copied
		}

		neededSize := len(src.Data)
		if len(buff.bufferInt16Interleaved) < neededSize {
			if cap(buff.bufferInt16Interleaved) >= neededSize {
				buff.bufferInt16Interleaved = buff.bufferInt16Interleaved[:neededSize]
			} else {
				buff.bufferInt16Interleaved = make([]int16, neededSize)
			}
		}

		copy(buff.bufferInt16Interleaved, src.Data)
		clone.Data = buff.bufferInt16Interleaved
		buff.tmp = clone

	case *Int16NonInterleaved:
		clone, ok := buff.tmp.(*Int16NonInterleaved)
		if ok {
			*clone = *src
		} else {
			copied := *src
			clone = &copied
		}

		neededSize := len(src.Data)
		if len(buff.bufferInt16NonInterleaved) < neededSize {
			if cap(buff.bufferInt16NonInterleaved) >= neededSize {
				buff.bufferInt16NonInterleaved = buff.bufferInt16NonInterleaved[:neededSize]
			} else {
				buff.bufferInt16NonInterleaved = make([][]int16, neededSize)
			}
		}

		for i := range src.Data {
			neededSize := len(src.Data[i])
			if len(buff.bufferInt16NonInterleaved[i]) < neededSize {
				if cap(buff.bufferInt16NonInterleaved[i]) >= neededSize {
					buff.bufferInt16NonInterleaved[i] = buff.bufferInt16NonInterleaved[i][:neededSize]
				} else {
					buff.bufferInt16NonInterleaved[i] = make([]int16, neededSize)
				}
			}

			copy(buff.bufferInt16NonInterleaved[i], src.Data[i])
		}
		clone.Data = buff.bufferInt16NonInterleaved
		buff.tmp = clone

	default:
		// TODO: Should have a routine to convert any format to one of the supported formats above
		panic(errUnsupportedFormat)
	}
}
