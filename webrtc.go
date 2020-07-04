package mediadevices

import (
	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/webrtc/v2"
)

// PeerConnection is an extension of webrtc.PeerConnection which allows to operate with raw video/audio
type PeerConnection struct {
	*webrtc.PeerConnection
	audioEncoders []codec.AudioEncoderBuilder
	videoEncoders []codec.VideoEncoderBuilder
}

// PeerConnectionOption is a function that configures the underlying PeerConnection
type PeerConnectionOption func(*PeerConnection)

func WithVideoEncoders(encoders ...codec.VideoEncoderBuilder) PeerConnectionOption {
	return PeerConnectionOption(func(pc *PeerConnection) {
		pc.videoEncoders = encoders
	})
}

func WithAudioEncoders(encoders ...codec.AudioEncoderBuilder) PeerConnectionOption {
	return PeerConnectionOption(func(pc *PeerConnection) {
		pc.audioEncoders = encoders
	})
}

func ExtendPeerConnection(pc *webrtc.PeerConnection, opts ...PeerConnectionOption) (*PeerConnection, error) {
	extPC := PeerConnection{
		PeerConnection: pc,
	}

	for _, opt := range opts {
		opt(&extPC)
	}

	return &extPC, nil
}

func (pc *PeerConnection) ExtAddTransceiverFromTrack(track Track, init ...webrtc.RtpTransceiverInit) (*webrtc.RTPTransceiver, error) {
}
