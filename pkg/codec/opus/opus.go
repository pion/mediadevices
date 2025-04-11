package opus

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/mediadevices/pkg/wave"
	"github.com/pion/mediadevices/pkg/wave/mixer"
)

/*
#include <opus.h>

int pion_set_encoder_bitrate(OpusEncoder *e, opus_int32 bitrate)
{
	return opus_encoder_ctl(e, OPUS_SET_BITRATE(bitrate));
}
*/
import "C"

type encoder struct {
	inBuff wave.Audio
	reader audio.Reader
	engine *C.OpusEncoder

	mu sync.Mutex
}

func newEncoder(r audio.Reader, p prop.Media, params Params) (codec.ReadCloser, error) {
	var cerror C.int

	if p.SampleRate == 0 {
		return nil, fmt.Errorf("opus: inProp.SampleRate is required")
	}

	if params.BitRate == 0 {
		params.BitRate = 32000
	}

	if params.ChannelMixer == nil {
		params.ChannelMixer = &mixer.MonoMixer{}
	}

	if !params.Latency.Validate() {
		return nil, fmt.Errorf("opus: unsupported latency %v", params.Latency)
	}

	channels := p.ChannelCount

	engine := C.opus_encoder_create(
		C.opus_int32(p.SampleRate),
		C.int(channels),
		C.OPUS_APPLICATION_VOIP,
		&cerror,
	)
	if cerror != C.OPUS_OK {
		return nil, errors.New("failed to create encoder engine")
	}

	rMix := audio.NewChannelMixer(channels, params.ChannelMixer)
	rBuf := audio.NewBuffer(params.Latency.samples(p.SampleRate))
	e := encoder{
		engine: engine,
		reader: rMix(rBuf(r)),
	}

	err := e.SetBitRate(params.BitRate)
	if err != nil {
		e.Close()
		return nil, err
	}
	return &e, nil
}

func (e *encoder) Read() ([]byte, func(), error) {
	buff, _, err := e.reader.Read()
	if err != nil {
		return nil, func() {}, err
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	if e.engine == nil {
		return nil, nil, io.EOF
	}

	encoded := make([]byte, 1024)
	var n C.opus_int32
	switch b := buff.(type) {
	case *wave.Int16Interleaved:
		n = C.opus_encode(
			e.engine,
			(*C.opus_int16)(&b.Data[0]),
			C.int(b.ChunkInfo().Len),
			(*C.uchar)(&encoded[0]),
			C.opus_int32(cap(encoded)),
		)
	case *wave.Float32Interleaved:
		n = C.opus_encode_float(
			e.engine,
			(*C.float)(&b.Data[0]),
			C.int(b.ChunkInfo().Len),
			(*C.uchar)(&encoded[0]),
			C.opus_int32(cap(encoded)),
		)
	default:
		err = errors.New("unknown type of audio buffer")
	}

	if n < 0 {
		err = errors.New("failed to encode")
	}

	return encoded[:n:n], func() {}, err
}

func (e *encoder) SetBitRate(bitRate int) error {
	cerror := C.pion_set_encoder_bitrate(
		e.engine,
		C.int(bitRate),
	)
	if cerror != C.OPUS_OK {
		return fmt.Errorf("failed to set encoder's bitrate to %d", bitRate)
	}

	return nil
}

func (e *encoder) Controller() codec.EncoderController {
	return e
}

func (e *encoder) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.engine == nil {
		return nil
	}
	C.opus_encoder_destroy(e.engine)
	e.engine = nil
	return nil
}
