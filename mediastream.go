package mediadevices

import (
	"sync"

	"github.com/pion/webrtc/v3"
)

// MediaStream is an interface that represents a collection of existing tracks.
type MediaStream interface {
	// GetAudioTracks implements https://w3c.github.io/mediacapture-main/#dom-mediastream-getaudiotracks
	GetAudioTracks() []Track
	// GetVideoTracks implements https://w3c.github.io/mediacapture-main/#dom-mediastream-getvideotracks
	GetVideoTracks() []Track
	// GetTracks implements https://w3c.github.io/mediacapture-main/#dom-mediastream-gettracks
	GetTracks() []Track
	// AddTrack implements https://w3c.github.io/mediacapture-main/#dom-mediastream-addtrack
	AddTrack(t Track)
	// RemoveTrack implements https://w3c.github.io/mediacapture-main/#dom-mediastream-removetrack
	RemoveTrack(t Track)
}

type mediaStream struct {
	tracks map[Track]struct{}
	l      sync.RWMutex
}

const trackTypeDefault webrtc.RTPCodecType = 0

// NewMediaStream creates a MediaStream interface that's defined in
// https://w3c.github.io/mediacapture-main/#dom-mediastream
func NewMediaStream(tracks ...Track) (MediaStream, error) {
	m := mediaStream{tracks: make(map[Track]struct{})}

	for _, track := range tracks {
		if _, ok := m.tracks[track]; !ok {
			m.tracks[track] = struct{}{}
		}
	}

	return &m, nil
}

func (m *mediaStream) GetAudioTracks() []Track {
	return m.queryTracks(webrtc.RTPCodecTypeAudio)
}

func (m *mediaStream) GetVideoTracks() []Track {
	return m.queryTracks(webrtc.RTPCodecTypeVideo)
}

func (m *mediaStream) GetTracks() []Track {
	return m.queryTracks(trackTypeDefault)
}

// queryTracks returns all tracks that are the same kind as t.
// If t is 0, which is the default, queryTracks will return all the tracks.
func (m *mediaStream) queryTracks(t webrtc.RTPCodecType) []Track {
	m.l.RLock()
	defer m.l.RUnlock()

	result := make([]Track, 0)
	for track := range m.tracks {
		if track.Kind() == t || t == trackTypeDefault {
			result = append(result, track)
		}
	}

	return result
}

func (m *mediaStream) AddTrack(t Track) {
	m.l.Lock()
	defer m.l.Unlock()

	if _, ok := m.tracks[t]; ok {
		return
	}

	m.tracks[t] = struct{}{}
}

func (m *mediaStream) RemoveTrack(t Track) {
	m.l.Lock()
	defer m.l.Unlock()

	delete(m.tracks, t)
}
