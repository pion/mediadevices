package main

import (
	"fmt"
	"net"
	"os"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/vpx"       // This is required to use VP8/VP9 video encoder
	_ "github.com/pion/mediadevices/pkg/driver/camera" // This is required to register camera adapter
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
)

const (
	mtu = 1000
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s host:port\n", os.Args[0])
		return
	}

	vp8Params, err := vpx.NewVP8Params()
	if err != nil {
		panic(err)
	}
	vp8Params.BitRate = 100000 // 100kbps

	md := mediadevices.NewMediaDevicesFromCodecs(
		map[webrtc.RTPCodecType][]*webrtc.RTPCodec{
			webrtc.RTPCodecTypeVideo: []*webrtc.RTPCodec{
				webrtc.NewRTPVP8Codec(100, 90000),
			},
		},
		mediadevices.WithTrackGenerator(
			func(_ uint8, _ uint32, id, _ string, codec *webrtc.RTPCodec) (
				mediadevices.LocalTrack, error,
			) {
				return newTrack(codec, id, os.Args[1]), nil
			},
		),
		mediadevices.WithVideoEncoders(&vp8Params),
	)

	_, err = md.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(p *prop.Media) {
			p.Width = 640
			p.Height = 480
		},
	})
	if err != nil {
		panic(err)
	}

	select {}
}

type track struct {
	codec      *webrtc.RTPCodec
	packetizer rtp.Packetizer
	id         string
	conn       net.Conn
}

func newTrack(codec *webrtc.RTPCodec, id, dest string) *track {
	addr, err := net.ResolveUDPAddr("udp", dest)
	if err != nil {
		panic(err)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}
	return &track{
		codec: codec,
		packetizer: rtp.NewPacketizer(
			mtu,
			codec.PayloadType,
			1,
			codec.Payloader,
			rtp.NewRandomSequencer(),
			codec.ClockRate,
		),
		id:   id,
		conn: conn,
	}
}

func (t *track) WriteSample(s media.Sample) error {
	buf := make([]byte, mtu)
	pkts := t.packetizer.Packetize(s.Data, s.Samples)
	for _, p := range pkts {
		n, err := p.MarshalTo(buf)
		if err != nil {
			panic(err)
		}
		_, _ = t.conn.Write(buf[:n])
	}
	return nil
}

func (t *track) Codec() *webrtc.RTPCodec {
	return t.codec
}

func (t *track) ID() string {
	return t.id
}

func (t *track) Kind() webrtc.RTPCodecType {
	return t.codec.Type
}
