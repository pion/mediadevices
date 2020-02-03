package mediadevices

import (
	"fmt"
	"io"
	"math/rand"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/driver"
	mio "github.com/pion/mediadevices/pkg/io"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
)

// Tracker is an interface that represent MediaStreamTrack
// Reference: https://w3c.github.io/mediacapture-main/#mediastreamtrack
type Tracker interface {
	Track() *webrtc.Track
	Stop()
}

type track struct {
	pc *webrtc.PeerConnection
	t  *webrtc.Track
	s  *sampler
}

func newTrack(pc *webrtc.PeerConnection, d driver.Driver, codecName string) (*track, error) {
	var kind webrtc.RTPCodecType
	switch d.(type) {
	case driver.VideoDriver:
		kind = webrtc.RTPCodecTypeVideo
	case driver.AudioDriver:
		kind = webrtc.RTPCodecTypeAudio
	}

	codecs := pc.GetRegisteredRTPCodecs(kind)
	var selectedCodec *webrtc.RTPCodec
	for _, c := range codecs {
		if c.Name == codecName {
			selectedCodec = c
			break
		}
	}
	if selectedCodec == nil {
		return nil, fmt.Errorf("track: %s is not registered in media engine", codecName)
	}

	t, err := pc.NewTrack(selectedCodec.PayloadType, rand.Uint32(), kind.String(), d.ID())
	if err != nil {
		return nil, err
	}

	return &track{
		pc: pc,
		t:  t,
		s:  newSampler(t),
	}, nil
}

func (t *track) Track() *webrtc.Track {
	return t.t
}

type videoTrack struct {
	*track
	d        driver.VideoDriver
	property video.AdvancedProperty
	encoder  io.ReadCloser
}

var _ Tracker = &videoTrack{}

func newVideoTrack(pc *webrtc.PeerConnection, d driver.VideoDriver, prop video.AdvancedProperty, codecName string) (*videoTrack, error) {
	t, err := newTrack(pc, d, codecName)
	if err != nil {
		return nil, err
	}

	r, err := d.Start(prop)
	if err != nil {
		return nil, err
	}

	// TODO: Remove hardcoded bitrate
	prop.BitRate = 100000
	encoder, err := codec.BuildVideoEncoder(codecName, r, prop)
	if err != nil {
		return nil, err
	}

	vt := videoTrack{
		track:    t,
		d:        d,
		property: prop,
		encoder:  encoder,
	}

	go vt.start()
	return &vt, nil
}

func (vt *videoTrack) start() {
	var n int
	var err error
	buff := make([]byte, 1024)
	for {
		n, err = vt.encoder.Read(buff)
		if err != nil {
			if e, ok := err.(*mio.InsufficientBufferError); ok {
				buff = make([]byte, 2*e.RequiredSize)
				continue
			}

			// TODO: better error handling
			panic(err)
		}

		vt.s.sample(buff[:n])
	}
}

func (vt *videoTrack) Stop() {
	vt.d.Stop()
	vt.encoder.Close()
}

type audioTrack struct {
	*track
	d        driver.AudioDriver
	property audio.AdvancedProperty
	encoder  io.ReadCloser
}

var _ Tracker = &audioTrack{}

func newAudioTrack(pc *webrtc.PeerConnection, d driver.AudioDriver, prop audio.AdvancedProperty, codecName string) (*audioTrack, error) {
	t, err := newTrack(pc, d, codecName)
	if err != nil {
		return nil, err
	}

	reader, err := d.Start(prop)
	if err != nil {
		return nil, err
	}

	// TODO: Not sure how to decide inProp and outProp
	inProp := prop
	outProp := prop

	encoder, err := codec.BuildAudioEncoder(codecName, reader, inProp, outProp)
	if err != nil {
		return nil, err
	}

	at := audioTrack{
		track:    t,
		d:        d,
		property: prop,
		encoder:  encoder,
	}
	go at.start()
	return &at, nil
}

func (t *audioTrack) start() {
	buff := make([]byte, 1024)
	sampleSize := uint32(float64(t.property.SampleRate) * t.property.Latency.Seconds())
	for {
		n, err := t.encoder.Read(buff)
		if err != nil {
			// TODO: better error handling
			panic(err)
		}

		t.t.WriteSample(media.Sample{
			Data:    buff[:n],
			Samples: sampleSize,
		})
	}
}

func (t *audioTrack) Stop() {
	t.d.Stop()
	t.encoder.Close()
}
