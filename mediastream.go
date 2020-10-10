package mediadevices

import (
	"sync"
)

// MediaStream is an interface that represents a collection of existing tracks.
type MediaStream interface {
	// GetAudioTracks implements https://w3c.github.io/mediacapture-main/#dom-mediastream-getaudiotracks
	GetAudioTracks() []Tracker
	// GetVideoTracks implements https://w3c.github.io/mediacapture-main/#dom-mediastream-getvideotracks
	GetVideoTracks() []Tracker
	// GetTracks implements https://w3c.github.io/mediacapture-main/#dom-mediastream-gettracks
	GetTracks() []Tracker
	// AddTrack implements https://w3c.github.io/mediacapture-main/#dom-mediastream-addtrack
	AddTrack(t Tracker)
	// RemoveTrack implements https://w3c.github.io/mediacapture-main/#dom-mediastream-removetrack
	RemoveTrack(t Tracker)
}

type mediaStream struct {
	trackers map[Tracker]struct{}
	l        sync.RWMutex
}

const trackTypeDefault MediaDeviceType = 0

// NewMediaStream creates a MediaStream interface that's defined in
// https://w3c.github.io/mediacapture-main/#dom-mediastream
func NewMediaStream(trackers ...Tracker) (MediaStream, error) {
	m := mediaStream{trackers: make(map[Tracker]struct{})}

	for _, tracker := range trackers {
		if _, ok := m.trackers[tracker]; !ok {
			m.trackers[tracker] = struct{}{}
		}
	}

	return &m, nil
}

func (m *mediaStream) GetAudioTracks() []Tracker {
	return m.queryTracks(AudioInput)
}

func (m *mediaStream) GetVideoTracks() []Tracker {
	return m.queryTracks(VideoInput)
}

func (m *mediaStream) GetTracks() []Tracker {
	return m.queryTracks(trackTypeDefault)
}

// queryTracks returns all tracks that are the same kind as t.
// If t is 0, which is the default, queryTracks will return all the tracks.
func (m *mediaStream) queryTracks(t MediaDeviceType) []Tracker {
	m.l.RLock()
	defer m.l.RUnlock()

	result := make([]Tracker, 0)
	for tracker := range m.trackers {
		if tracker.Kind() == t || t == trackTypeDefault {
			result = append(result, tracker)
		}
	}

	return result
}

func (m *mediaStream) AddTrack(t Tracker) {
	m.l.Lock()
	defer m.l.Unlock()

	if _, ok := m.trackers[t]; ok {
		return
	}

	m.trackers[t] = struct{}{}
}

func (m *mediaStream) RemoveTrack(t Tracker) {
	m.l.Lock()
	defer m.l.Unlock()

	delete(m.trackers, t)
}
