package mediadevices

import (
	"sync"

	"github.com/pion/webrtc/v4"
)

// MediaStream is an interface that represents a collection of existing tracks.
type MediaStream interface {
	// GetAudioTracks implements https://w3c.github.io/mediacapture-main/#dom-mediastream-getaudiotracks
	GetAudioTracks() []TrackLocal
	// GetVideoTracks implements https://w3c.github.io/mediacapture-main/#dom-mediastream-getvideotracks
	GetVideoTracks() []TrackLocal
	// GetTracks implements https://w3c.github.io/mediacapture-main/#dom-mediastream-gettracks
	GetTracks() []TrackLocal
	// AddTrack implements https://w3c.github.io/mediacapture-main/#dom-mediastream-addtrack
	AddTrack(t TrackLocal)
	// RemoveTrack implements https://w3c.github.io/mediacapture-main/#dom-mediastream-removetrack
	RemoveTrack(t TrackLocal)
}

type mediaStream struct {
	tracks map[TrackLocal]struct{}
	l      sync.RWMutex
}

const trackTypeDefault webrtc.RTPCodecType = 0

// NewMediaStream creates a MediaStream interface that's defined in
// https://w3c.github.io/mediacapture-main/#dom-mediastream
func NewMediaStream(tracks ...TrackLocal) (MediaStream, error) {
	m := mediaStream{tracks: make(map[TrackLocal]struct{})}

	for _, track := range tracks {
		if _, ok := m.tracks[track]; !ok {
			m.tracks[track] = struct{}{}
		}
	}

	return &m, nil
}

func (m *mediaStream) GetAudioTracks() []TrackLocal {
	return m.queryTracks(webrtc.RTPCodecTypeAudio)
}

func (m *mediaStream) GetVideoTracks() []TrackLocal {
	return m.queryTracks(webrtc.RTPCodecTypeVideo)
}

func (m *mediaStream) GetTracks() []TrackLocal {
	return m.queryTracks(trackTypeDefault)
}

// queryTracks returns all tracks that are the same kind as t.
// If t is 0, which is the default, queryTracks will return all the tracks.
func (m *mediaStream) queryTracks(t webrtc.RTPCodecType) []TrackLocal {
	m.l.RLock()
	defer m.l.RUnlock()

	result := make([]TrackLocal, 0)
	for track := range m.tracks {
		if track.Kind() == t || t == trackTypeDefault {
			result = append(result, track)
		}
	}

	return result
}

func (m *mediaStream) AddTrack(t TrackLocal) {
	m.l.Lock()
	defer m.l.Unlock()

	if _, ok := m.tracks[t]; ok {
		return
	}

	m.tracks[t] = struct{}{}
}

func (m *mediaStream) RemoveTrack(t TrackLocal) {
	m.l.Lock()
	defer m.l.Unlock()

	delete(m.tracks, t)
}
