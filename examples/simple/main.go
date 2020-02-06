package main

import (
	"fmt"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/examples/internal/signal"
	_ "github.com/pion/mediadevices/pkg/codec/openh264" // This is required to register h264 video encoder
	_ "github.com/pion/mediadevices/pkg/codec/opus"     // This is required to register opus audio encoder
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

	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})

	md := mediadevices.NewMediaDevices(peerConnection)

	s, err := md.GetUserMedia(mediadevices.MediaStreamConstraints{
		Audio: func(c *mediadevices.MediaTrackConstraints) {
			c.Codec = webrtc.Opus
			c.Enabled = true
		},
		Video: func(c *mediadevices.MediaTrackConstraints) {
			c.Codec = webrtc.H264
			c.Enabled = true
			c.Width = 800
			c.Height = 480
		},
	})
	if err != nil {
		panic(err)
	}

	for _, tracker := range s.GetTracks() {
		_, err = peerConnection.AddTrack(tracker.Track())
		if err != nil {
			panic(err)
		}
	}

	// Wait for the offer to be pasted
	offer := webrtc.SessionDescription{}
	signal.Decode(signal.MustReadStdin(), &offer)

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
