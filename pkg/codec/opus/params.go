package opus

import (
	"time"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/mediadevices/pkg/wave/mixer"
)

// Latency is a type of OPUS codec frame duration.
type Latency time.Duration

// Latency values available in OPUS codec.
const (
	Latency2500us Latency = Latency(2500 * time.Microsecond)
	Latency5ms    Latency = Latency(5 * time.Millisecond)
	Latency10ms   Latency = Latency(10 * time.Millisecond)
	Latency20ms   Latency = Latency(20 * time.Millisecond)
	Latency40ms   Latency = Latency(40 * time.Millisecond)
	Latency60ms   Latency = Latency(60 * time.Millisecond)
)

// Validate that the Latency is allowed in OPUS.
func (l Latency) Validate() bool {
	switch l {
	case Latency2500us, Latency5ms, Latency10ms, Latency20ms, Latency40ms, Latency60ms:
		return true
	default:
		return false
	}
}

// Duration returns latency in time.Duration.
func (l Latency) Duration() time.Duration {
	return time.Duration(l)
}

// samples returns number of samples for given sample rate.
func (l Latency) samples(sampleRate int) int {
	return int(l.Duration() * time.Duration(sampleRate) / time.Second)
}

// Params stores opus specific encoding parameters.
type Params struct {
	codec.BaseParams
	// ChannelMixer is a mixer to be used if number of given and expected channels differ.
	ChannelMixer mixer.ChannelMixer

	// Expected latency of the codec.
	Latency Latency
}

// NewParams returns default opus codec specific parameters.
func NewParams() (Params, error) {
	return Params{
		Latency: Latency20ms,
	}, nil
}

// RTPCodec represents the codec metadata
func (p *Params) RTPCodec() *codec.RTPCodec {
	c := codec.NewRTPOpusCodec(48000)
	c.Latency = time.Duration(p.Latency)
	return c
}

// BuildAudioEncoder builds opus encoder with given params
func (p *Params) BuildAudioEncoder(r audio.Reader, property prop.Media) (codec.ReadCloser, error) {
	return newEncoder(r, property, *p)
}
