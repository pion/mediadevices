package main

import (
	"fmt"
	"net"
	"os"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/vpx"       // This is required to use VP8/VP9 video encoder
	_ "github.com/pion/mediadevices/pkg/driver/camera" // This is required to register camera adapter
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/prop"
)

const (
	mtu = 1000
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s host:port\n", os.Args[0])
		return
	}
	dest := os.Args[1]

	vp8Params, err := vpx.NewVP8Params()
	must(err)
	vp8Params.BitRate = 100000 // 100kbps

	codecSelector := mediadevices.NewCodecSelector(
		mediadevices.WithVideoEncoders(&vp8Params),
	)

	mediaStream, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(c *mediadevices.MediaTrackConstraints) {
			c.FrameFormat = prop.FrameFormat(frame.FormatYUY2)
			c.Width = prop.Int(640)
			c.Height = prop.Int(480)
		},
		Codec: codecSelector,
	})
	must(err)

	videoTrack := mediaStream.GetVideoTracks()[0]
	defer videoTrack.Close()

	rtpReader, err := videoTrack.NewRTPReader(vp8Params.RTPCodec().Name, mtu)
	must(err)

	addr, err := net.ResolveUDPAddr("udp", dest)
	must(err)
	conn, err := net.DialUDP("udp", nil, addr)
	must(err)

	buff := make([]byte, mtu)
	for {
		pkts, release, err := rtpReader.Read()
		must(err)

		for _, pkt := range pkts {
			n, err := pkt.MarshalTo(buff)
			must(err)

			_, err = conn.Write(buff[:n])
			must(err)
		}

		release()
	}
}
