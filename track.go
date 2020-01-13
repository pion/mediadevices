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

type videoTrack struct {
	t       *webrtc.Track
	s       *sampler
	d       driver.VideoDriver
	setting driver.VideoSetting
	decoder frame.Decoder
	encoder codec.VideoEncoder
}

func newVideoTrack(pc *webrtc.PeerConnection, d driver.VideoDriver, setting driver.VideoSetting, codecName string) (*videoTrack, error) {
	var err error
	decoder, err := frame.NewDecoder(setting.FrameFormat)
	if err != nil {
		return nil, err
	}

	var selectedCodec *webrtc.RTPCodec
	codecs := pc.GetRegisteredRTPCodecs(webrtc.RTPCodecTypeVideo)
	for _, c := range codecs {
		if c.Name == codecName {
			selectedCodec = c
			break
		}
	}
	if selectedCodec == nil {
		return nil, fmt.Errorf("video track: %s is not registered in media engine", codecName)
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

	track, err := pc.NewTrack(selectedCodec.PayloadType, rand.Uint32(), "video", d.ID())
	if err != nil {
		encoder.Close()
		return nil, err
	}

	vt := videoTrack{
		t:       track,
		s:       newSampler(track.Codec().ClockRate),
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

	sample := vt.s.sample(encoded)
	err = vt.t.WriteSample(sample)
	if err != nil {
		// TODO: probably do some logging here
		return
	}
}

func (vt *videoTrack) Track() *webrtc.Track {
	return vt.t
}

func (vt *videoTrack) Stop() {
	vt.d.Stop()
	vt.encoder.Close()
}

type audioTrack struct {
	t       *webrtc.Track
	s       *sampler
	d       driver.AudioDriver
	setting driver.AudioSetting
	encoder codec.AudioEncoder
}

func newAudioTrack(pc *webrtc.PeerConnection, d driver.AudioDriver, setting driver.AudioSetting, codecName string) (*audioTrack, error) {
	var err error

	var selectedCodec *webrtc.RTPCodec
	codecs := pc.GetRegisteredRTPCodecs(webrtc.RTPCodecTypeAudio)
	for _, c := range codecs {
		if c.Name == codecName {
			selectedCodec = c
			break
		}
	}
	if selectedCodec == nil {
		return nil, fmt.Errorf("audio track: %s is not registered in media engine", codecName)
	}

	encoder, err := codec.BuildAudioEncoder(codecName, codec.AudioSetting{
		SampleRate: setting.SampleRate,
	})
	if err != nil {
		return nil, err
	}

	track, err := pc.NewTrack(selectedCodec.PayloadType, rand.Uint32(), "audio", d.ID())
	if err != nil {
		encoder.Close()
		return nil, err
	}

	vt := audioTrack{
		t:       track,
		s:       newSampler(track.Codec().ClockRate),
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

	sample := vt.s.sample(encoded)
	err = vt.t.WriteSample(sample)
	if err != nil {
		// TODO: probably do some logging here
		return
	}
}

func (vt *audioTrack) Track() *webrtc.Track {
	return vt.t
}

func (vt *audioTrack) Stop() {
	vt.d.Stop()
	vt.encoder.Close()
}
