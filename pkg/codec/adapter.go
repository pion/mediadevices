package codec

import (
	"math/rand"

	mio "github.com/pion/mediadevices/pkg/io"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v2"
)

const (
	defaultMTU = 1200
)

type rtpReadCloserImpl struct {
	packetize        func(payload []byte) []*rtp.Packet
	encoder          ReadCloser
	buff             []byte
	unreadRTPPackets []*rtp.Packet
}

func NewRTPReadCloser(codec *webrtc.RTPCodec, reader ReadCloser, sample SamplerFunc) (RTPReadCloser, error) {
	packetizer := rtp.NewPacketizer(
		defaultMTU,
		codec.PayloadType,
		rand.Uint32(),
		codec.Payloader,
		rtp.NewRandomSequencer(),
		codec.ClockRate,
	)
	return &rtpReadCloserImpl{
		packetize: func(payload []byte) []*rtp.Packet {
			return packetizer.Packetize(payload, sample())
		},
	}, nil
}

func (rc *rtpReadCloserImpl) ReadRTP() (packet *rtp.Packet, err error) {
	var n int

	packet = rc.readRTPPacket()
	if packet != nil {
		return
	}

	for {
		n, err = rc.encoder.Read(rc.buff)
		if err == nil {
			break
		}

		e, ok := err.(*mio.InsufficientBufferError)
		if !ok {
			return nil, err
		}

		rc.buff = make([]byte, 2*e.RequiredSize)
	}

	rc.unreadRTPPackets = rc.packetize(rc.buff[:n])
	return rc.readRTPPacket(), nil
}

// readRTPPacket reads unreadRTPPackets and mark the rtp packet as "read",
// which essentially removes it from the list. If the return value is nil,
// it means that there's no unread rtp packets.
func (rc *rtpReadCloserImpl) readRTPPacket() (packet *rtp.Packet) {
	if len(rc.unreadRTPPackets) == 0 {
		return
	}
	packet, rc.unreadRTPPackets = rc.unreadRTPPackets[0], rc.unreadRTPPackets[1:]
	return
}

func (rc *rtpReadCloserImpl) Close() {
	rc.encoder.Close()
}
