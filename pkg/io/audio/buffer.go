package audio

import (
	"errors"

	"github.com/pion/mediadevices/pkg/wave"
)

var errUnsupported = errors.New("unsupported audio format")

// NewBuffer creates audio transform to buffer signal to have exact nSample samples.
func NewBuffer(nSamples int) TransformFunc {
	var inBuff wave.Audio

	return func(r Reader) Reader {
		return ReaderFunc(func() (wave.Audio, func(), error) {
			for {
				if inBuff != nil && inBuff.ChunkInfo().Len >= nSamples {
					break
				}

				buff, _, err := r.Read()
				if err != nil {
					return nil, func() {}, err
				}
				switch b := buff.(type) {
				case *wave.Float32Interleaved:
					ib, ok := inBuff.(*wave.Float32Interleaved)
					if !ok || ib.Size.Channels != b.Size.Channels {
						ib = wave.NewFloat32Interleaved(
							wave.ChunkInfo{
								SamplingRate: b.Size.SamplingRate,
								Channels:     b.Size.Channels,
								Len:          nSamples,
							},
						)
						ib.Data = ib.Data[:0]
						ib.Size.Len = 0
						inBuff = ib
					}
					ib.Data = append(ib.Data, b.Data...)
					ib.Size.Len += b.Size.Len

				case *wave.Int16Interleaved:
					ib, ok := inBuff.(*wave.Int16Interleaved)
					if !ok || ib.Size.Channels != b.Size.Channels {
						ib = wave.NewInt16Interleaved(
							wave.ChunkInfo{
								SamplingRate: b.Size.SamplingRate,
								Channels:     b.Size.Channels,
								Len:          nSamples,
							},
						)
						ib.Data = ib.Data[:0]
						ib.Size.Len = 0
						inBuff = ib
					}
					ib.Data = append(ib.Data, b.Data...)
					ib.Size.Len += b.Size.Len

				default:
					return nil, func() {}, errUnsupported
				}
			}
			switch ib := inBuff.(type) {
			case *wave.Int16Interleaved:
				ibCopy := *ib
				ibCopy.Size.Len = nSamples
				n := nSamples * ib.Size.Channels
				ibCopy.Data = make([]int16, n)
				copy(ibCopy.Data, ib.Data)
				ib.Data = ib.Data[n:]
				ib.Size.Len -= nSamples
				return &ibCopy, func() {}, nil

			case *wave.Float32Interleaved:
				ibCopy := *ib
				ibCopy.Size.Len = nSamples
				n := nSamples * ib.Size.Channels
				ibCopy.Data = make([]float32, n)
				copy(ibCopy.Data, ib.Data)
				ib.Data = ib.Data[n:]
				ib.Size.Len -= nSamples
				return &ibCopy, func() {}, nil
			}
			return nil, func() {}, errUnsupported
		})
	}
}
