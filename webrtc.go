package mediadevices

import (
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v2"
)

// == WebRTC v3 design ==

// Reader is an interface to handle incoming RTP stream.
type Reader interface {
	ReadRTP() (*rtp.Packet, error)
	WriteRTCP(rtcp.Packet) error
}

// TrackBase represents common MediaStreamTrack functionality of LocalTrack and RemoteTrack.
type TrackBase interface {
	ID() string
}

type LocalRTPTrack interface {
	TrackBase
	Reader

	// SetParameters sets information about how the data is to be encoded.
	// This will be called by PeerConnection according to the result of
	// SDP based negotiation.
	// It will be called via RTPSender.Parameters() by PeerConnection to
	// tell the negotiated media codec information.
	//
	// This is pion's extension to process data without having encoder/decoder
	// in webrtc package.
	SetParameters(RTPParameters) error
}

// RTPParameters represents RTCRtpParameters which contains information about
// how the RTC data is to be encoded/decoded.
//
// ref: https://developer.mozilla.org/en-US/docs/Web/API/RTCRtpSendParameters
type RTPParameters struct {
	SSRC          uint32
	SelectedCodec *webrtc.RTPCodec
	Codecs        []*webrtc.RTPCodec
}
