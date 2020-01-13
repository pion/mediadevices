package mediadevices

import (
	"fmt"
	"math/rand"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/frame"
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
	decoder frame.Decoder
	encoder codec.VideoEncoder
}

var _ Tracker = &videoTrack{}

func newVideoTrack(pc *webrtc.PeerConnection, d driver.VideoDriver, setting driver.VideoSetting, codecName string) (*videoTrack, error) {
	t, err := newTrack(pc, d, codecName)
	if err != nil {
		return nil, err
	}

	decoder, err := frame.NewDecoder(setting.FrameFormat)
	if err != nil {
		return nil, err
	}

	encoder, err := codec.BuildVideoEncoder(codecName, codec.VideoSetting{
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
		decoder: decoder,
		encoder: encoder,
	}

	err = d.Start(setting, vt.dataCb)
	if err != nil {
		encoder.Close()
		return nil, err
	}

	return &vt, nil
}

func (vt *videoTrack) dataCb(b []byte) {
	img, err := vt.decoder.Decode(b, vt.setting.Width, vt.setting.Height)
	if err != nil {
		// TODO: probably do some logging here
		return
	}

	encoded, err := vt.encoder.Encode(img)
	if err != nil {
		// TODO: probably do some logging here
		return
	}

	err = vt.s.sample(encoded)
	if err != nil {
		// TODO: probably do some logging here
		return
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
	encoder codec.AudioEncoder
}

var _ Tracker = &audioTrack{}

func newAudioTrack(pc *webrtc.PeerConnection, d driver.AudioDriver, setting driver.AudioSetting, codecName string) (*audioTrack, error) {
	t, err := newTrack(pc, d, codecName)
	if err != nil {
		return nil, err
	}

	encoder, err := codec.BuildAudioEncoder(codecName, codec.AudioSetting{
		SampleRate: setting.SampleRate,
	})
	if err != nil {
		return nil, err
	}

	vt := audioTrack{
		track:   t,
		d:       d,
		setting: setting,
		encoder: encoder,
	}

	err = d.Start(setting, vt.dataCb)
	if err != nil {
		encoder.Close()
		return nil, err
	}

	return &vt, nil
}

func (vt *audioTrack) dataCb(b []int16) {
	encoded, err := vt.encoder.Encode(b)
	if err != nil {
		// TODO: probably do some logging here
		return
	}

	err = vt.s.sample(encoded)
	if err != nil {
		// TODO: probably do some logging here
		return
	}
}

func (vt *audioTrack) Stop() {
	vt.d.Stop()
	vt.encoder.Close()
}
