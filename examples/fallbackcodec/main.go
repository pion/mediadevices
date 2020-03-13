package main

import (
	"fmt"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/examples/internal/signal"
	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/webrtc/v2"

	_ "github.com/pion/mediadevices/pkg/codec/opus" // This is required to register opus audio encoder
	"github.com/pion/mediadevices/pkg/codec/vaapi"  // Load hardware accelerated codecs
	"github.com/pion/mediadevices/pkg/codec/vpx"    // Load software codecs

	// Note: If you don't have a camera or microphone or your adapters are not supported,
	//       you can always swap your adapters with our dummy adapters below.
	// _ "github.com/pion/mediadevices/pkg/driver/videotest"
	// _ "github.com/pion/mediadevices/pkg/driver/audiotest"
	_ "github.com/pion/mediadevices/pkg/driver/camera"     // This is required to register camera adapter
	_ "github.com/pion/mediadevices/pkg/driver/microphone" // This is required to register microphone adapter
)

func main() {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Register vaapi as a primary codec, and vpx as a fallback codec.
	codec.Register(webrtc.VP8, codec.VideoEncoderFallbacks(
		// Use if hardware acceleration is available.
		codec.NamedVideoCodec{
			Name:  "vaapi",
			Codec: codec.VideoEncoderBuilder(vaapi.NewVP8Encoder),
		},
		// Software implementation should always available.
		codec.NamedVideoCodec{
			Name:  "vpx",
			Codec: codec.VideoEncoderBuilder(vpx.NewVP8Encoder),
		},
	))

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

	md := mediadevices.NewMediaDevices(peerConnection)

	s, err := md.GetUserMedia(mediadevices.MediaStreamConstraints{
		Audio: func(c *mediadevices.MediaTrackConstraints) {
			c.CodecName = webrtc.Opus
			c.Enabled = true
			c.BitRate = 32000 // 32kbps
		},
		Video: func(c *mediadevices.MediaTrackConstraints) {
			c.CodecName = webrtc.VP8
			c.FrameFormat = frame.FormatYUY2
			c.Enabled = true
			c.Width = 640
			c.Height = 480
			c.FrameRate = 30
			c.BitRate = 400000

			// Codec specific parameter should be given for each implementation
			vaapiParam, err := vaapi.NewVP8Param()
			if err != nil {
				panic(err)
			}
			vaapiParam.RateControlMode = vaapi.RateControlVBR
			vaapiParam.RateControl.BitsPerSecond = 400000
			vaapiParam.RateControl.TargetPercentage = 95

			vpxParam, err := vpx.NewVP8Param()
			if err != nil {
				panic(err)
			}
			vpxParam.RateControlEndUsage = vpx.RateControlVBR
			vpxParam.RateControlUndershootPercent = 100
			vpxParam.RateControlOvershootPercent = 5

			c.CodecParams = map[string]interface{}{
				"vaapi": vpxParam,
				"vpx":   vpxParam,
			}
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
