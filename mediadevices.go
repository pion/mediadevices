package mediadevices

import (
	"github.com/pion/mediadevices/driver"
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
	r := driver.Manager.Query()[0]
	err := r.Driver.Open()
	if err != nil {
		return nil, err
	}
	d := r.Driver.(driver.VideoDriver)
	spec := d.Specs()[0]

	tracker, err := newVideoTrack(m.pc, r.ID, d, spec, webrtc.H264)
	if err != nil {
		return nil, err
	}

	s, err := NewMediaStream(tracker)
	if err != nil {
		return nil, err
	}

	return s, nil
}
