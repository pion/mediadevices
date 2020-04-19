package main

import (
	"fmt"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/examples/internal/signal"
	"github.com/pion/webrtc/v2"

	// This is required to use opus audio encoder
	"github.com/pion/mediadevices/pkg/codec/opus"
	"github.com/pion/mediadevices/pkg/prop"

	// If you don't like vpx, you can also use x264 by importing as below
	// "github.com/pion/mediadevices/pkg/codec/x264" // This is required to use h264 video encoder
	// or you can also use openh264 for alternative h264 implementation
	// "github.com/pion/mediadevices/pkg/codec/openh264"
	"github.com/pion/mediadevices/pkg/codec/vpx" // This is required to use VP8/VP9 video encoder

	// Note: If you don't have a camera or microphone or your adapters are not supported,
	//       you can always swap your adapters with our dummy adapters below.
	// _ "github.com/pion/mediadevices/pkg/driver/videotest"
	// _ "github.com/pion/mediadevices/pkg/driver/audiotest"
	_ "github.com/pion/mediadevices/pkg/driver/camera"     // This is required to register camera adapter
	_ "github.com/pion/mediadevices/pkg/driver/microphone" // This is required to register microphone adapter
)

const (
	videoCodecName = webrtc.VP8
)

func main() {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Wait for the offer to be pasted
	offer := webrtc.SessionDescription{}
	signal.Decode(signal.MustReadStdin(), &offer)

	// Create a new RTCPeerConnection
	mediaEngine := webrtc.MediaEngine{}
	if err := mediaEngine.PopulateFromSDP(offer); err != nil {
		panic(err)
	}
	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})

	opusParams, err := opus.NewParams()
	if err != nil {
		panic(err)
	}
	opusParams.BitRate = 32000 // 32kbps

	vp8Params, err := vpx.NewVP8Params()
	if err != nil {
		panic(err)
	}
	vp8Params.BitRate = 100000 // 100kbps

	md := mediadevices.NewMediaDevices(
		peerConnection,
		mediadevices.WithVideoEncoders(&vp8Params),
		mediadevices.WithAudioEncoders(&opusParams),
	)

	s, err := md.GetUserMedia(mediadevices.MediaStreamConstraints{
		Audio: func(p *prop.Media) {},
		Video: func(p *prop.Media) {
			p.Width = 640
			p.Height = 480
		},
	})
	if err != nil {
		panic(err)
	}

	for _, tracker := range s.GetTracks() {
		t := tracker.Track()
		tracker.OnEnded(func(err error) {
			fmt.Printf("Track (ID: %s, Label: %s) ended with error: %v\n",
				t.ID(), t.Label(), err)
		})
		_, err = peerConnection.AddTransceiverFromTrack(t,
			webrtc.RtpTransceiverInit{
				Direction: webrtc.RTPTransceiverDirectionSendonly,
			},
		)
		if err != nil {
			panic(err)
		}
	}

	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}

	// Create an answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	// Output the answer in base64 so we can paste it in browser
	fmt.Println(signal.Encode(answer))
	select {}
}
