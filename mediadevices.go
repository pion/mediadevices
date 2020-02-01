package mediadevices

import (
	"fmt"
	"math"

	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/webrtc/v2"
)

// MediaDevices is an interface that's defined on https://developer.mozilla.org/en-US/docs/Web/API/MediaDevices
type MediaDevices interface {
	GetUserMedia(constraints MediaStreamConstraints) (MediaStream, error)
}

// NewMediaDevices creates MediaDevices interface that provides access to connected media input devices
// like cameras and microphones, as well as screen sharing.
// In essence, it lets you obtain access to any hardware source of media data.
func NewMediaDevices(pc *webrtc.PeerConnection) MediaDevices {
	return &mediaDevices{pc}
}

type mediaDevices struct {
	pc *webrtc.PeerConnection
}

// GetUserMedia prompts the user for permission to use a media input which produces a MediaStream
// with tracks containing the requested types of media.
// Reference: https://developer.mozilla.org/en-US/docs/Web/API/MediaDevices/getUserMedia
func (m *mediaDevices) GetUserMedia(constraints MediaStreamConstraints) (MediaStream, error) {
	// TODO: It should return media stream based on constraints
	trackers := make([]Tracker, 0)

	if constraints.Video.Enabled {
		tracker, err := m.videoSelect(constraints.Video)
		if err != nil {
			return nil, err
		}

		trackers = append(trackers, tracker)
	}

	if constraints.Audio.Enabled {
		tracker, err := m.audioSelect(constraints.Audio)
		if err != nil {
			return nil, err
		}

		trackers = append(trackers, tracker)
	}

	s, err := NewMediaStream(trackers...)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// videoSelect implements SelectSettings algorithm for video type.
// Reference: https://w3c.github.io/mediacapture-main/#dfn-selectsettings
func (m *mediaDevices) videoSelect(constraints VideoTrackConstraints) (Tracker, error) {
	drivers := driver.GetManager().VideoDrivers()

	var bestDriver driver.VideoDriver
	var bestSetting driver.VideoSetting
	minFitnessDist := math.Inf(1)

	for _, d := range drivers {
		wasClosed := d.Status() == driver.StateClosed

		if wasClosed {
			err := d.Open()
			if err != nil {
				// Skip this driver if we failed to open because we can't get the settings
				continue
			}
		}

		vd := d.(driver.VideoDriver)
		for _, setting := range vd.Settings() {
			fitnessDist := constraints.fitnessDistance(setting)

			if fitnessDist < minFitnessDist {
				minFitnessDist = fitnessDist
				bestDriver = vd
				bestSetting = setting
			}
		}

		if wasClosed {
			// Since it was closed, we should close it to avoid a leak
			d.Close()
		}
	}

	if bestDriver == nil {
		return nil, fmt.Errorf("failed to find the best setting")
	}

	if bestDriver.Status() == driver.StateClosed {
		err := bestDriver.Open()
		if err != nil {
			return nil, fmt.Errorf("failed in opening the best video driver")
		}
	}
	return newVideoTrack(m.pc, bestDriver, bestSetting, constraints.Codec)
}

// audioSelect implements SelectSettings algorithm for audio type.
// Reference: https://w3c.github.io/mediacapture-main/#dfn-selectsettings
func (m *mediaDevices) audioSelect(constraints AudioTrackConstraints) (Tracker, error) {
	drivers := driver.GetManager().AudioDrivers()

	var bestDriver driver.AudioDriver
	var bestSetting driver.AudioSetting
	minFitnessDist := math.Inf(1)

	for _, d := range drivers {
		wasClosed := d.Status() == driver.StateClosed

		if wasClosed {
			err := d.Open()
			if err != nil {
				// Skip this driver if we failed to open because we can't get the settings
				continue
			}
		}

		ad := d.(driver.AudioDriver)
		for _, setting := range ad.Settings() {
			fitnessDist := constraints.fitnessDistance(setting)

			if fitnessDist < minFitnessDist {
				minFitnessDist = fitnessDist
				bestDriver = ad
				bestSetting = setting
			}
		}

		if wasClosed {
			// Since it was closed, we should close it to avoid a leak
			d.Close()
		}
	}

	if bestDriver == nil {
		return nil, fmt.Errorf("failed to find the best setting")
	}

	if bestDriver.Status() == driver.StateClosed {
		err := bestDriver.Open()
		if err != nil {
			return nil, fmt.Errorf("failed in opening the best audio driver")
		}
	}
	return newAudioTrack(m.pc, bestDriver, bestSetting, constraints.Codec)
}
