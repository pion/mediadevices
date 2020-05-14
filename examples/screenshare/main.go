package main

import (
	"fmt"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/examples/internal/signal"
	"github.com/pion/mediadevices/pkg/codec/openh264"

	// This is required to use VP8/VP9 video encoder
	// _ "github.com/pion/mediadevices/pkg/driver/screen" // This is required to register screen capture adapter
	_ "github.com/pion/mediadevices/pkg/driver/videotest" // This is required to register screen capture adapter
	extwebrtc "github.com/pion/mediadevices/pkg/ext/webrtc"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/webrtc/v2"
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

	openh264Encoder, err := openh264.NewParams()
	if err != nil {
		panic(err)
	}
	openh264Encoder.BitRate = 100000 // 100kbps

	// Create a new RTCPeerConnection
	mediaEngine := extwebrtc.MediaEngine{}
	mediaEngine.AddEncoderBuilders(&openh264Encoder)
	api := extwebrtc.NewAPI(extwebrtc.WithMediaEngine(mediaEngine))
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})

	s, err := mediadevices.GetDisplayMedia(mediadevices.MediaStreamConstraints{
		Video: func(p *prop.Media) {},
	})
	if err != nil {
		panic(err)
	}

	for _, track := range s.GetTracks() {
		_, err = peerConnection.ExtAddTransceiverFromTrack(track,
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
