package mediadevices

import (
	"fmt"
	"math"

	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/prop"
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
	codecs := make(map[webrtc.RTPCodecType][]*webrtc.RTPCodec)
	for _, kind := range []webrtc.RTPCodecType{
		webrtc.RTPCodecTypeAudio,
		webrtc.RTPCodecTypeVideo,
	} {
		codecs[kind] = pc.GetRegisteredRTPCodecs(kind)
	}

	return &mediaDevices{codecs}
}

// NewMediaDevicesFromCodecs creates MediaDevices interface from lists of the available codecs
// that provides access to connected media input devices like cameras and microphones,
// as well as screen sharing.
// In essence, it lets you obtain access to any hardware source of media data.
func NewMediaDevicesFromCodecs(codecs map[webrtc.RTPCodecType][]*webrtc.RTPCodec) MediaDevices {
	return &mediaDevices{codecs}
}

type mediaDevices struct {
	codecs map[webrtc.RTPCodecType][]*webrtc.RTPCodec
}

// GetUserMedia prompts the user for permission to use a media input which produces a MediaStream
// with tracks containing the requested types of media.
// Reference: https://developer.mozilla.org/en-US/docs/Web/API/MediaDevices/getUserMedia
func (m *mediaDevices) GetUserMedia(constraints MediaStreamConstraints) (MediaStream, error) {
	// TODO: It should return media stream based on constraints
	trackers := make([]Tracker, 0)

	var videoConstraints, audioConstraints MediaTrackConstraints
	if constraints.Video != nil {
		constraints.Video(&videoConstraints)
	}

	if constraints.Audio != nil {
		constraints.Audio(&audioConstraints)
	}

	if videoConstraints.Enabled {
		tracker, err := m.selectVideo(videoConstraints)
		if err != nil {
			return nil, err
		}

		trackers = append(trackers, tracker)
	}

	if audioConstraints.Enabled {
		tracker, err := m.selectAudio(audioConstraints)
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

func queryDriverProperties(filter driver.FilterFn) map[driver.Driver][]prop.Media {
	var needToClose []driver.Driver
	drivers := driver.GetManager().Query(filter)
	m := make(map[driver.Driver][]prop.Media)

	for _, d := range drivers {
		if d.Status() == driver.StateClosed {
			err := d.Open()
			if err != nil {
				// Skip this driver if we failed to open because we can't get the properties
				continue
			}
			needToClose = append(needToClose, d)
		}

		m[d] = d.Properties()
	}

	for _, d := range needToClose {
		// Since it was closed, we should close it to avoid a leak
		d.Close()
	}

	return m
}

// select implements SelectSettings algorithm.
// Reference: https://w3c.github.io/mediacapture-main/#dfn-selectsettings
func selectBestDriver(filter driver.FilterFn, constraints MediaTrackConstraints) (driver.Driver, MediaTrackConstraints, error) {
	var bestDriver driver.Driver
	var bestProp prop.Media
	minFitnessDist := math.Inf(1)

	driverProperties := queryDriverProperties(filter)
	for d, props := range driverProperties {
		for _, p := range props {
			fitnessDist := constraints.Media.FitnessDistance(p)
			if fitnessDist < minFitnessDist {
				minFitnessDist = fitnessDist
				bestDriver = d
				bestProp = p
			}
		}
	}

	if bestDriver == nil {
		return nil, MediaTrackConstraints{}, fmt.Errorf("failed to find the best driver that fits the constraints")
	}

	// Reset Codec because bestProp only contains either audio.Prop or video.Prop
	bestProp.Codec = constraints.Codec
	bestConstraint := MediaTrackConstraints{
		Media:   bestProp,
		Enabled: true,
	}
	return bestDriver, bestConstraint, nil
}

func (m *mediaDevices) selectAudio(constraints MediaTrackConstraints) (Tracker, error) {
	d, c, err := selectBestDriver(driver.FilterAudioRecorder(), constraints)
	if err != nil {
		return nil, err
	}

	return newAudioTrack(m.codecs[webrtc.RTPCodecTypeAudio], d, c)
}
func (m *mediaDevices) selectVideo(constraints MediaTrackConstraints) (Tracker, error) {
	d, c, err := selectBestDriver(driver.FilterVideoRecorder(), constraints)
	if err != nil {
		return nil, err
	}

	return newVideoTrack(m.codecs[webrtc.RTPCodecTypeVideo], d, c)
}
