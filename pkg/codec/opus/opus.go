package opus

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"unsafe"

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

var errDecUninitialized = fmt.Errorf("opus decoder uninitialized")

type Decoder struct {
	p *C.struct_OpusDecoder
	// Same purpose as encoder struct
	mem         []byte
	sample_rate int
	channels    int
}

// NewDecoder allocates a new Opus decoder and initializes it with the
// appropriate parameters. All related memory is managed by the Go GC.
func NewDecoder(sample_rate int, channels int) (*Decoder, error) {
	var dec Decoder
	err := dec.Init(sample_rate, channels)
	if err != nil {
		return nil, err
	}
	return &dec, nil
}

func (dec *Decoder) Init(sample_rate int, channels int) error {
	if dec.p != nil {
		return fmt.Errorf("opus decoder already initialized")
	}
	if channels != 1 && channels != 2 {
		return fmt.Errorf("Number of channels must be 1 or 2: %d", channels)
	}
	size := C.opus_decoder_get_size(C.int(channels))
	dec.sample_rate = sample_rate
	dec.channels = channels
	dec.mem = make([]byte, size)
	dec.p = (*C.OpusDecoder)(unsafe.Pointer(&dec.mem[0]))
	errno := C.opus_decoder_init(
		dec.p,
		C.opus_int32(sample_rate),
		C.int(channels))
	if errno != 0 {
		return Error(errno)
	}
	return nil
}

// Decode encoded Opus data into the supplied buffer. On success, returns the
// number of samples correctly written to the target buffer.
func (dec *Decoder) Decode(data []byte, pcm []int16) (int, error) {
	if dec.p == nil {
		return 0, errDecUninitialized
	}
	if len(data) == 0 {
		return 0, fmt.Errorf("opus: no data supplied")
	}
	if len(pcm) == 0 {
		return 0, fmt.Errorf("opus: target buffer empty")
	}
	if cap(pcm)%dec.channels != 0 {
		return 0, fmt.Errorf("opus: target buffer capacity must be multiple of channels")
	}
	n := int(C.opus_decode(
		dec.p,
		(*C.uchar)(&data[0]),
		C.opus_int32(len(data)),
		(*C.opus_int16)(&pcm[0]),
		C.int(cap(pcm)/dec.channels),
		0))
	if n < 0 {
		return 0, Error(n)
	}
	return n, nil
}

// Decode encoded Opus data into the supplied buffer. On success, returns the
// number of samples correctly written to the target buffer.
func (dec *Decoder) DecodeFloat32(data []byte, pcm []float32) (int, error) {
	if dec.p == nil {
		return 0, errDecUninitialized
	}
	if len(data) == 0 {
		return 0, fmt.Errorf("opus: no data supplied")
	}
	if len(pcm) == 0 {
		return 0, fmt.Errorf("opus: target buffer empty")
	}
	if cap(pcm)%dec.channels != 0 {
		return 0, fmt.Errorf("opus: target buffer capacity must be multiple of channels")
	}
	n := int(C.opus_decode_float(
		dec.p,
		(*C.uchar)(&data[0]),
		C.opus_int32(len(data)),
		(*C.float)(&pcm[0]),
		C.int(cap(pcm)/dec.channels),
		0))
	if n < 0 {
		return 0, Error(n)
	}
	return n, nil
}

// DecodeFEC encoded Opus data into the supplied buffer with forward error
// correction.
//
// It is to be used on the packet directly following the lost one.  The supplied
// buffer needs to be exactly the duration of audio that is missing
//
// When a packet is considered "lost", DecodeFEC can be called on the next
// packet in order to try and recover some of the lost data. The PCM needs to be
// exactly the duration of audio that is missing.  `LastPacketDuration()` can be
// used on the decoder to get the length of the last packet.  Note also that in
// order to use this feature the encoder needs to be configured with
// SetInBandFEC(true) and SetPacketLossPerc(x) options.
//
// Note that DecodeFEC automatically falls back to PLC when no FEC data is
// available in the provided packet.
func (dec *Decoder) DecodeFEC(data []byte, pcm []int16) error {
	if dec.p == nil {
		return errDecUninitialized
	}
	if len(data) == 0 {
		return fmt.Errorf("opus: no data supplied")
	}
	if len(pcm) == 0 {
		return fmt.Errorf("opus: target buffer empty")
	}
	if cap(pcm)%dec.channels != 0 {
		return fmt.Errorf("opus: target buffer capacity must be multiple of channels")
	}
	n := int(C.opus_decode(
		dec.p,
		(*C.uchar)(&data[0]),
		C.opus_int32(len(data)),
		(*C.opus_int16)(&pcm[0]),
		C.int(cap(pcm)/dec.channels),
		1))
	if n < 0 {
		return Error(n)
	}
	return nil
}

// DecodeFECFloat32 encoded Opus data into the supplied buffer with forward error
// correction. It is to be used on the packet directly following the lost one.
// The supplied buffer needs to be exactly the duration of audio that is missing
func (dec *Decoder) DecodeFECFloat32(data []byte, pcm []float32) error {
	if dec.p == nil {
		return errDecUninitialized
	}
	if len(data) == 0 {
		return fmt.Errorf("opus: no data supplied")
	}
	if len(pcm) == 0 {
		return fmt.Errorf("opus: target buffer empty")
	}
	if cap(pcm)%dec.channels != 0 {
		return fmt.Errorf("opus: target buffer capacity must be multiple of channels")
	}
	n := int(C.opus_decode_float(
		dec.p,
		(*C.uchar)(&data[0]),
		C.opus_int32(len(data)),
		(*C.float)(&pcm[0]),
		C.int(cap(pcm)/dec.channels),
		1))
	if n < 0 {
		return Error(n)
	}
	return nil
}

// DecodePLC recovers a lost packet using Opus Packet Loss Concealment feature.
//
// The supplied buffer needs to be exactly the duration of audio that is missing.
// When a packet is considered "lost", `DecodePLC` and `DecodePLCFloat32` methods
// can be called in order to obtain something better sounding than just silence.
// The PCM needs to be exactly the duration of audio that is missing.
// `LastPacketDuration()` can be used on the decoder to get the length of the
// last packet.
//
// This option does not require any additional encoder options. Unlike FEC,
// PLC does not introduce additional latency. It is calculated from the previous
// packet, not from the next one.
func (dec *Decoder) DecodePLC(pcm []int16) error {
	if dec.p == nil {
		return errDecUninitialized
	}
	if len(pcm) == 0 {
		return fmt.Errorf("opus: target buffer empty")
	}
	if cap(pcm)%dec.channels != 0 {
		return fmt.Errorf("opus: output buffer capacity must be multiple of channels")
	}
	n := int(C.opus_decode(
		dec.p,
		nil,
		0,
		(*C.opus_int16)(&pcm[0]),
		C.int(cap(pcm)/dec.channels),
		0))
	if n < 0 {
		return Error(n)
	}
	return nil
}

// DecodePLCFloat32 recovers a lost packet using Opus Packet Loss Concealment feature.
// The supplied buffer needs to be exactly the duration of audio that is missing.
func (dec *Decoder) DecodePLCFloat32(pcm []float32) error {
	if dec.p == nil {
		return errDecUninitialized
	}
	if len(pcm) == 0 {
		return fmt.Errorf("opus: target buffer empty")
	}
	if cap(pcm)%dec.channels != 0 {
		return fmt.Errorf("opus: output buffer capacity must be multiple of channels")
	}
	n := int(C.opus_decode_float(
		dec.p,
		nil,
		0,
		(*C.float)(&pcm[0]),
		C.int(cap(pcm)/dec.channels),
		0))
	if n < 0 {
		return Error(n)
	}
	return nil
}

// LastPacketDuration gets the duration (in samples)
// of the last packet successfully decoded or concealed.
func (dec *Decoder) LastPacketDuration() (int, error) {
	var samples C.opus_int32
	// res := C.bridge_decoder_get_last_packet_duration(dec.p, &samples)
	// if res != C.OPUS_OK {
	// 	return 0, Error(res)
	// }
	return int(samples), nil
}
