package mediadevices

import (
	"fmt"
	"math/rand"

	"github.com/pion/codec"
	"github.com/pion/codec/h264"
	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/webrtc/v2"
)

type tracker interface {
	Track() *webrtc.Track
	Stop()
}

type videoTrack struct {
	t       *webrtc.Track
	s       *sampler
	d       driver.VideoDriver
	spec    driver.VideoSpec
	decoder frame.Decoder
	encoder codec.Encoder
}

func newVideoTrack(pc *webrtc.PeerConnection, d driver.VideoDriver, spec driver.VideoSpec, codecName string) (*videoTrack, error) {
	var err error
	decoder, err := frame.NewDecoder(spec.FrameFormat)
	if err != nil {
		return nil, err
	}

	var payloadType uint8
	var encoder codec.Encoder
	switch codecName {
	case webrtc.H264:
		payloadType = webrtc.DefaultPayloadTypeH264
		encoder, err = h264.NewEncoder(h264.Options{
			Width:        spec.Width,
			Height:       spec.Height,
			Bitrate:      1000000,
			MaxFrameRate: 30,
		})
	default:
		err = fmt.Errorf("%s is currently not supported", codecName)
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
		spec:    spec,
		decoder: decoder,
		encoder: encoder,
	}

	go d.Start(spec, vt.dataCb)
	return &vt, nil
}

func (vt *videoTrack) dataCb(b []byte) {
	img, err := vt.decoder.Decode(b, vt.spec.Width, vt.spec.Height)
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
