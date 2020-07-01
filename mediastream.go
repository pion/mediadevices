package mediadevices

import (
	"sync"

	"github.com/pion/webrtc/v2"
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
	tracks map[string]Track
	l      sync.RWMutex
}

const rtpCodecTypeDefault webrtc.RTPCodecType = 0

// NewMediaStream creates a MediaStream interface that's defined in
// https://w3c.github.io/mediacapture-main/#dom-mediastream
func NewMediaStream(tracks ...Track) (MediaStream, error) {
	m := mediaStream{tracks: make(map[string]Track)}

	for _, track := range tracks {
		id := track.ID()
		if _, ok := m.tracks[id]; !ok {
			m.tracks[id] = track
		}
	}

	return &m, nil
}

func (m *mediaStream) GetAudioTracks() []Track {
	return m.queryTracks(func(t Track) bool { return t.Kind() == TrackKindAudio })
}

func (m *mediaStream) GetVideoTracks() []Track {
	return m.queryTracks(func(t Track) bool { return t.Kind() == TrackKindVideo })
}

func (m *mediaStream) GetTracks() []Track {
	return m.queryTracks(func(t Track) bool { return true })
}

// queryTracks returns all tracks that are the same kind as t.
// If t is 0, which is the default, queryTracks will return all the tracks.
func (m *mediaStream) queryTracks(filter func(track Track) bool) []Track {
	m.l.RLock()
	defer m.l.RUnlock()

	result := make([]Track, 0)
	for _, track := range m.tracks {
		if filter(track) {
			result = append(result, track)
		}
	}

	return result
}

func (m *mediaStream) AddTrack(t Track) {
	m.l.Lock()
	defer m.l.Unlock()

	id := t.ID()
	if _, ok := m.tracks[id]; ok {
		return
	}

	m.tracks[id] = t
}

func (m *mediaStream) RemoveTrack(t Track) {
	m.l.Lock()
	defer m.l.Unlock()

	delete(m.tracks, t.ID())
}
