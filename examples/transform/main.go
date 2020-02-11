package main

import (
	"fmt"
	"image"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/examples/internal/signal"
	_ "github.com/pion/mediadevices/pkg/codec/openh264" // This is required to register h264 video encoder
	_ "github.com/pion/mediadevices/pkg/codec/opus"     // This is required to register opus audio encoder
	_ "github.com/pion/mediadevices/pkg/codec/vpx"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/webrtc/v2"
)

const (
	videoCodecName = webrtc.VP8
)

func removeBlue(r video.Reader) video.Reader {
	return video.ReaderFunc(func() (img image.Image, err error) {
		img, err = r.Read()
		if err != nil {
			return
		}

		yuvImg, ok := img.(*image.YCbCr)
		if !ok {
			return img, nil
		}

		yuvImg.Cb = make([]uint8, len(yuvImg.Cb))
		return
	})
}

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

	md := mediadevices.NewMediaDevices(peerConnection)

	s, err := md.GetUserMedia(mediadevices.MediaStreamConstraints{
		Audio: func(c *mediadevices.MediaTrackConstraints) {
			c.CodecName = webrtc.Opus
			c.Enabled = true
			c.BitRate = 32000 // 32kbps
		},
		Video: func(c *mediadevices.MediaTrackConstraints) {
			c.CodecName = videoCodecName
			c.FrameFormat = frame.FormatI420 // most of the encoder accepts I420
			c.Enabled = true
			c.Width = 640
			c.Height = 480
			c.BitRate = 100000 // 100kbps
			c.VideoTransform = removeBlue
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
		_, err = peerConnection.AddTrack(t)
		if err != nil {
			panic(err)
		}
	}

	// Tweak transceiver direction to work with Firefox
	for _, t := range peerConnection.GetTransceivers() {
		t.Direction = webrtc.RTPTransceiverDirectionSendonly
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
