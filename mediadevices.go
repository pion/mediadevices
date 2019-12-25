package mediadevices

import (
	"github.com/pion/mediadevices/pkg/driver"
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
	d := driver.Manager.Query()[0]
	err := d.Open()
	if err != nil {
		return nil, err
	}
	vd := d.(driver.VideoDriver)
	spec := vd.Specs()[0]

	tracker, err := newVideoTrack(m.pc, vd, spec, webrtc.H264)
	if err != nil {
		return nil, err
	}

	s, err := NewMediaStream(tracker)
	if err != nil {
		return nil, err
	}

	return s, nil
}
