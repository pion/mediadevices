package mediadevices

import (
	"fmt"
	"io"
	"math/rand"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/driver"
	mio "github.com/pion/mediadevices/pkg/io"
	"github.com/pion/webrtc/v2"
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
	switch d.Info().Kind {
	case driver.Video:
		kind = webrtc.RTPCodecTypeVideo
	case driver.Audio:
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
	d       driver.VideoDriver
	setting driver.VideoSetting
	encoder io.ReadCloser
}

var _ Tracker = &videoTrack{}

func newVideoTrack(pc *webrtc.PeerConnection, d driver.VideoDriver, setting driver.VideoSetting, codecName string) (*videoTrack, error) {
	t, err := newTrack(pc, d, codecName)
	if err != nil {
		return nil, err
	}

	r, err := d.Start(setting)
	if err != nil {
		return nil, err
	}

	encoder, err := codec.BuildVideoEncoder(codecName, r, codec.VideoSetting{
		Width:         setting.Width,
		Height:        setting.Height,
		TargetBitRate: 1000000,
		FrameRate:     30,
	})
	if err != nil {
		return nil, err
	}

	vt := videoTrack{
		track:   t,
		d:       d,
		setting: setting,
		encoder: encoder,
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
	d       driver.AudioDriver
	setting driver.AudioSetting
	encoder io.ReadCloser
}

var _ Tracker = &audioTrack{}

func newAudioTrack(pc *webrtc.PeerConnection, d driver.AudioDriver, setting driver.AudioSetting, codecName string) (*audioTrack, error) {
	t, err := newTrack(pc, d, codecName)
	if err != nil {
		return nil, err
	}

	reader, err := d.Start(setting)
	if err != nil {
		return nil, err
	}

	codecSetting := codec.AudioSetting{
		InSampleRate: setting.SampleRate,
	}

	encoder, err := codec.BuildAudioEncoder(codecName, reader, codecSetting)
	if err != nil {
		return nil, err
	}

	at := audioTrack{
		track:   t,
		d:       d,
		setting: setting,
		encoder: encoder,
	}
	go at.start()
	return &at, nil
}

func (t *audioTrack) start() {
	buff := make([]byte, 1024)
	for {
		n, err := t.encoder.Read(buff)
		if err != nil {
			// TODO: better error handling
			panic(err)
		}
		t.s.sample(buff[:n])
	}
}

func (t *audioTrack) Stop() {
	t.d.Stop()
	t.encoder.Close()
}
