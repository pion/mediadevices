package mediadevices

import (
	"sync"

	"github.com/pion/webrtc/v2"
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
	trackers map[string]Tracker
	l        sync.RWMutex
}

const rtpCodecTypeDefault webrtc.RTPCodecType = 0

// NewMediaStream creates a MediaStream interface that's defined in
// https://w3c.github.io/mediacapture-main/#dom-mediastream
func NewMediaStream(trackers ...Tracker) (MediaStream, error) {
	m := mediaStream{trackers: make(map[string]Tracker)}

	for _, tracker := range trackers {
		id := tracker.LocalTrack().ID()
		if _, ok := m.trackers[id]; !ok {
			m.trackers[id] = tracker
		}
	}

	return &m, nil
}

func (m *mediaStream) GetAudioTracks() []Tracker {
	return m.queryTracks(webrtc.RTPCodecTypeAudio)
}

func (m *mediaStream) GetVideoTracks() []Tracker {
	return m.queryTracks(webrtc.RTPCodecTypeVideo)
}

func (m *mediaStream) GetTracks() []Tracker {
	return m.queryTracks(rtpCodecTypeDefault)
}

// queryTracks returns all tracks that are the same kind as t.
// If t is 0, which is the default, queryTracks will return all the tracks.
func (m *mediaStream) queryTracks(t webrtc.RTPCodecType) []Tracker {
	m.l.RLock()
	defer m.l.RUnlock()

	result := make([]Tracker, 0)
	for _, tracker := range m.trackers {
		if tracker.LocalTrack().Kind() == t || t == rtpCodecTypeDefault {
			result = append(result, tracker)
		}
	}

	return result
}

func (m *mediaStream) AddTrack(t Tracker) {
	m.l.Lock()
	defer m.l.Unlock()

	id := t.LocalTrack().ID()
	if _, ok := m.trackers[id]; ok {
		return
	}

	m.trackers[id] = t
}

func (m *mediaStream) RemoveTrack(t Tracker) {
	m.l.Lock()
	defer m.l.Unlock()

	delete(m.trackers, t.LocalTrack().ID())
}
