package main

import (
        "flag"
        "log"
        "fmt"

        "github.com/pion/mediadevices"
        "github.com/pion/mediadevices/examples/internal/signal"
        "github.com/pion/mediadevices/pkg/frame"
        "github.com/pion/mediadevices/pkg/prop"
        "github.com/pion/webrtc/v3"

        // If you don't like x264, you can also use vpx by importing as below
        // "github.com/pion/mediadevices/pkg/codec/vpx" // This is required to use VP8/VP9 video encoder
        // or you can also use openh264 for alternative h264 implementation
        // "github.com/pion/mediadevices/pkg/codec/openh264"
        // or if you use a raspberry pi like, you can use mmal for using its hardware encoder
        "github.com/pion/mediadevices/pkg/codec/mmal"
        "github.com/pion/mediadevices/pkg/codec/opus" // This is required to use opus audio encoder
        //"github.com/pion/mediadevices/pkg/codec/x264" // This is required to use h264 video encoder

        // Note: If you don't have a camera or microphone or your adapters are not supported,
        //       you can always swap your adapters with our dummy adapters below.
        // _ "github.com/pion/mediadevices/pkg/driver/videotest"
        // _ "github.com/pion/mediadevices/pkg/driver/audiotest"
        _ "github.com/pion/mediadevices/pkg/driver/camera"     // This is required to register camera adapter
        _ "github.com/pion/mediadevices/pkg/driver/microphone" // This is required to register microphone adapter


)



func echo(w http.ResponseWriter, r *http.Request) {



func main() {

                        config := webrtc.Configuration{
                                ICEServers: []webrtc.ICEServer{
                                        {
                                                URLs: []string{"stun:stun.l.google.com:19302"},
                                        },
                                },
                        }



                        // Create a new RTCPeerConnection
                        mmalParams, err := mmal.NewParams()
                        if err != nil {
                                panic(err)
                        }
                        mmalParams.BitRate = 500_000 // 500kbps

                        opusParams, err := opus.NewParams()
                        if err != nil {
                                panic(err)
                        }

                        codecSelector := mediadevices.NewCodecSelector(
                                mediadevices.WithVideoEncoders(&mmalParams),
                                mediadevices.WithAudioEncoders(&opusParams),
                        )

                        mediaEngine := webrtc.MediaEngine{}
                        codecSelector.Populate(&mediaEngine)

                        api := webrtc.NewAPI(webrtc.WithMediaEngine(&mediaEngine))
                        peerConnection, err := api.NewPeerConnection(config)
                        if err != nil {
                                panic(err)
                        }

                        // Set the handler for ICE connection state
                        // This will notify you when the peer has connected/disconnected
                        peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
                                fmt.Printf("Connection State has changed %s \n", connectionState.String())
                        })

                        s, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
                                Video: func(c *mediadevices.MediaTrackConstraints) {
                                        c.FrameFormat = prop.FrameFormat(frame.FormatYUY2)
                                        //c.FrameFormat = prop.FrameFormatExact(frame.FormatI420)
                                        c.Width = prop.Int(352)
                                        c.Height = prop.Int(288)
                                },
                                Audio: func(c *mediadevices.MediaTrackConstraints) {
                                },
                                Codec: codecSelector,
                        })
                        if err != nil {
                                panic(err)
                        }

                        for _, track := range s.GetTracks() {
                                track.OnEnded(func(err error) {
                                        fmt.Printf("Track (ID: %s) ended with error: %v\n",
                                                track.ID(), err)
                                })

                                _, err = peerConnection.AddTransceiverFromTrack(track,
                                        webrtc.RtpTransceiverInit{
                                                Direction: webrtc.RTPTransceiverDirectionSendonly,
                                        },
                                )
                                if err != nil {
                                        panic(err)
                                }
                        }



                        // Create an offer
                        offer, err := peerConnection.CreateOffer(nil)
                        if err != nil {
                                panic(err)
                        }

                        //Create a channel that is blocked until ICE Gathering is complete
                        gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

                        // Sets the LocalDescription, and starts our UDP listeners
                        err = peerConnection.SetLocalDescription(offer)
                        if err != nil {
                                panic(err)
                        }

                        //Wait for ICE gathering to complete (non-trickle ICE)
                        <-gatherComplete

                        //Output the offer SDP with the final ICE candidate
                        log.Println( signal.Encode(*peerConnection.LocalDescription()) )


			//Wait for the browser answer to be pasted in the console
                	answer := webrtc.SessionDescription{}
                        signal.Decode(signal.MustReadStdin(), &answer)
                        

               		// Set the remote SessionDescription
                        err = peerConnection.SetRemoteDescription(answer)
                        if err != nil {
                                panic(err)
                        }


                select {}  //Blocks main routine from exiting
}
