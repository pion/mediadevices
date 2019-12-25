package mediadevices

import (
	"sync"

	"github.com/pion/webrtc/v2"
)

type MediaStream interface {
	GetAudioTracks() []tracker
	GetVideoTracks() []tracker
	GetTracks() []tracker
	AddTrack(t tracker)
	RemoveTrack(t tracker)
}

type mediaStream struct {
	trackers map[string]tracker
	l        sync.RWMutex
}

const rtpCodecTypeDefault webrtc.RTPCodecType = 0

// NewMediaStream creates a MediaStream interface that's defined in
// https://w3c.github.io/mediacapture-main/#dom-mediastream
func NewMediaStream(trackers ...tracker) (MediaStream, error) {
	m := mediaStream{trackers: make(map[string]tracker)}

	for _, tracker := range trackers {
		id := tracker.Track().ID()
		if _, ok := m.trackers[id]; !ok {
			m.trackers[id] = tracker
		}
	}

	return &m, nil
}

// GetAudioTracks implements https://w3c.github.io/mediacapture-main/#dom-mediastream-getaudiotracks
func (m *mediaStream) GetAudioTracks() []tracker {
	return m.queryTracks(webrtc.RTPCodecTypeAudio)
}

// GetVideoTracks implements https://w3c.github.io/mediacapture-main/#dom-mediastream-getvideotracks
func (m *mediaStream) GetVideoTracks() []tracker {
	return m.queryTracks(webrtc.RTPCodecTypeVideo)
}

// GetTracks implements https://w3c.github.io/mediacapture-main/#dom-mediastream-gettracks
func (m *mediaStream) GetTracks() []tracker {
	return m.queryTracks(rtpCodecTypeDefault)
}

// queryTracks returns all tracks that are the same kind as t.
// If t is 0, which is the default, queryTracks will return all the tracks.
func (m *mediaStream) queryTracks(t webrtc.RTPCodecType) []tracker {
	m.l.RLock()
	defer m.l.RUnlock()

	result := make([]tracker, 0)
	for _, tracker := range m.trackers {
		if tracker.Track().Kind() == t || t == rtpCodecTypeDefault {
			result = append(result, tracker)
		}
	}

	return result
}

// AddTrack implements https://w3c.github.io/mediacapture-main/#dom-mediastream-addtrack
func (m *mediaStream) AddTrack(t tracker) {
	m.l.Lock()
	defer m.l.Unlock()

	id := t.Track().ID()
	if _, ok := m.trackers[id]; ok {
		return
	}

	m.trackers[id] = t
}

// RemoveTrack implements https://w3c.github.io/mediacapture-main/#dom-mediastream-removetrack
func (m *mediaStream) RemoveTrack(t tracker) {
	m.l.Lock()
	defer m.l.Unlock()

	delete(m.trackers, t.Track().ID())
}
