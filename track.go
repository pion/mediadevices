package mediadevices

import (
	"math/rand"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/codec/h264"
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
	encoder codec.Encoder
}

func newVideoTrack(pc *webrtc.PeerConnection, d driver.VideoDriver, setting driver.VideoSetting, codecName Codec) (*videoTrack, error) {
	var err error
	decoder, err := frame.NewDecoder(setting.FrameFormat)
	if err != nil {
		return nil, err
	}

	var payloadType uint8
	var encoder codec.Encoder
	switch codecName {
	default:
		payloadType = webrtc.DefaultPayloadTypeH264
		encoder, err = h264.NewEncoder(h264.Options{
			Width:        setting.Width,
			Height:       setting.Height,
			Bitrate:      1000000,
			MaxFrameRate: 30,
		})
	}

	if err != nil {
		return nil, err
	}

	track, err := pc.NewTrack(payloadType, rand.Uint32(), "video", d.ID())
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

	go d.Start(setting, vt.dataCb)
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
