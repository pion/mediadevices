package webrtc

import (
	"fmt"
	"math/rand"

	"github.com/pion/mediadevices"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v2"
)

type Track interface {
	mediadevices.Track
}

type LocalTrack interface {
	Track
	ReadRTP() ([]*rtp.Packet, error)
}

type EncoderBuilder interface {
	Codec() *webrtc.RTPCodec
	Kind() webrtc.RTPCodecType
	BuildEncoder(Track) (LocalTrack, error)
}

type MediaEngine struct {
	webrtc.MediaEngine
	encoderBuilders []EncoderBuilder
}

func (engine *MediaEngine) AddEncoderBuilders(builders ...EncoderBuilder) {
	engine.encoderBuilders = append(engine.encoderBuilders, builders...)
	for _, builder := range builders {
		engine.RegisterCodec(builder.Codec())
	}
}

type API struct {
	webrtc.API
	mediaEngine MediaEngine
}

func NewAPI(options ...func(*API)) *API {
	var api API
	for _, option := range options {
		option(&api)
	}
	return &api
}

func WithMediaEngine(m MediaEngine) func(*API) {
	return func(a *API) {
		a.mediaEngine = m
	}
}

func (api *API) NewPeerConnection(configuration webrtc.Configuration) (*PeerConnection, error) {
	return api.NewPeerConnection(configuration)
}

type PeerConnection struct {
	webrtc.PeerConnection
}

func buildEncoder(encoderBuilders []EncoderBuilder, track Track) EncoderBuilder {
	for _, encoderBuilder := range encoderBuilders {
		encoder, err := encoderBuilder(track)
		if err == nil {
			return encoder
		}
	}
	return nil
}

func (pc *PeerConnection) ExtAddTransceiverFromTrack(track Track, init ...webrtc.RtpTransceiverInit) (*webrtc.RTPTransceiver, error) {
	builder := buildEncoder(pc.mediaEngine.encoderBuilders, track)
	if builder == nil {
		return nil, fmt.Errorf("failed to find a compatible encoder")
	}

	rtpCodec := builder.Codec()
	trackImpl, err := pc.NewTrack(rtpCodec.PayloadType, rand.Uint32(), track.ID(), rtpCodec.Type.String())
	if err != nil {
		return nil, err
	}

	localTrack, err := builder.BuildEncoder(track)
	if err != nil {
		return nil, err
	}

	trans, err := pc.AddTransceiverFromTrack(trackImpl, init...)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			rtpPackets, err := localTrack.ReadRTP()
			if err != nil {
				return
			}

			for _, rtpPacket := range rtpPackets {
				err = trackImpl.WriteRTP(rtpPacket)
				if err != nil {
					return
				}
			}
		}
	}()

	return trans, nil
}
