package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/examples/internal/signal"
	"github.com/pion/mediadevices/pkg/codec/opus"
	"github.com/pion/mediadevices/pkg/wave"
	"github.com/pion/webrtc/v3"
)

const (
	sampleRate = 48000
	channels   = 2
	sampleSize = 2
)

type AudioFile struct {
	rawReader      *os.File
	bufferedReader *bufio.Reader
	rawBuffer      []byte
	decoder        wave.Decoder
	ticker         *time.Ticker
}

func NewAudioFile(path string) (*AudioFile, error) {
	// Assume 48000 sample rate, mono channel, and S16LE interleaved
	latency := time.Millisecond * 120
	readFrequency := time.Second / latency
	readLen := sampleRate * channels * sampleSize / int(readFrequency)
	decoder, err := wave.NewDecoder(&wave.RawFormat{
		SampleSize:  sampleSize,
		IsFloat:     false,
		Interleaved: true,
	})

	fmt.Printf(`
Latency: %s
Read Frequency: %d Hz
Buffer Len: %d bytes
`, latency, readFrequency, readLen)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return &AudioFile{
		rawReader:      f,
		bufferedReader: bufio.NewReader(f),
		rawBuffer:      make([]byte, readLen),
		decoder:        decoder,
		ticker:         time.NewTicker(latency),
	}, nil
}

func (file *AudioFile) Read() (chunk wave.Audio, release func(), err error) {
	_, err = io.ReadFull(file.bufferedReader, file.rawBuffer)
	if err != nil {
		// Keep looping the audio
		file.rawReader.Seek(0, 0)
		_, err = io.ReadFull(file.bufferedReader, file.rawBuffer)
		if err != nil {
			return
		}
	}

	chunk, err = file.decoder.Decode(binary.LittleEndian, file.rawBuffer, channels)
	if err != nil {
		return
	}

	int16Chunk := chunk.(*wave.Int16Interleaved)
	int16Chunk.Size.SamplingRate = sampleRate

	// Slow down reading so that it matches 48 KHz
	<-file.ticker.C
	return
}

func (file *AudioFile) Close() error {
	return file.rawReader.Close()
}

func (file *AudioFile) ID() string {
	return "raw-audio-from-file"
}

func must(err error) {
	if err != nil {
		panic(err)
	}
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

	opusParams, err := opus.NewParams()
	if err != nil {
		panic(err)
	}
	opusParams.Latency = opus.Latency20ms

	codecSelector := mediadevices.NewCodecSelector(
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

	audioSource, err := NewAudioFile("audio.raw")
	must(err)

	audioTrack := mediadevices.NewAudioTrack(audioSource, codecSelector)

	audioTrack.OnEnded(func(err error) {
		fmt.Printf("Track (ID: %s) ended with error: %v\n",
			audioTrack.ID(), err)
	})

	_, err = peerConnection.AddTransceiverFromTrack(audioTrack,
		webrtc.RtpTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionSendonly,
		},
	)
	must(err)

	// Set the remote SessionDescription
	must(peerConnection.SetRemoteDescription(offer))

	// Create an answer
	answer, err := peerConnection.CreateAnswer(nil)
	must(err)

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Sets the LocalDescription, and starts our UDP listeners
	must(peerConnection.SetLocalDescription(answer))

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	// Output the answer in base64 so we can paste it in browser
	fmt.Println(signal.Encode(*peerConnection.LocalDescription()))

	// Block forever
	select {}
}
