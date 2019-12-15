package mediadevices

import (
	"github.com/pion/mediadevices/camera"
	"github.com/pion/webrtc/v2"
)

// MediaDevices is an interface that's defined on https://developer.mozilla.org/en-US/docs/Web/API/MediaDevices
type MediaDevices interface {
	GetUserMedia(constraints MediaStreamConstraints) (MediaStream, error)
}

func NewMediaDevices(pc *webrtc.PeerConnection) MediaDevices {
	return &mediaDevices{pc}
}

type mediaDevices struct {
	pc *webrtc.PeerConnection
}

func (m *mediaDevices) GetUserMedia(constraints MediaStreamConstraints) (MediaStream, error) {
	// TODO: It should return media stream based on constraints
	c, err := camera.New(camera.Options{
		PC:     m.pc,
		Codec:  webrtc.H264,
		Width:  640,
		Height: 480,
	})
	if err != nil {
		return nil, err
	}

	s, err := NewMediaStream(c.Track())
	if err != nil {
		return nil, err
	}

	go c.Start()
	return s, nil
}
