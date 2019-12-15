package mediadevices

import (
	"sync"

	"github.com/pion/webrtc/v2"
)

type MediaStream interface {
	GetAudioTracks() []*webrtc.Track
	GetVideoTracks() []*webrtc.Track
	GetTracks() []*webrtc.Track
	AddTrack(t *webrtc.Track)
	RemoveTrack(t *webrtc.Track)
}

type mediaStream struct {
	tracks map[string]*webrtc.Track
	l      sync.RWMutex
}

const rtpCodecTypeDefault webrtc.RTPCodecType = 0

// NewMediaStream creates a MediaStream interface that's defined in
// https://w3c.github.io/mediacapture-main/#dom-mediastream
func NewMediaStream(tracks ...*webrtc.Track) (MediaStream, error) {
	m := mediaStream{tracks: make(map[string]*webrtc.Track)}

	for _, track := range tracks {
		id := track.ID()
		if _, ok := m.tracks[id]; !ok {
			m.tracks[id] = track
		}
	}

	return &m, nil
}

// GetAudioTracks implements https://w3c.github.io/mediacapture-main/#dom-mediastream-getaudiotracks
func (m *mediaStream) GetAudioTracks() []*webrtc.Track {
	return m.queryTracks(webrtc.RTPCodecTypeAudio)
}

// GetVideoTracks implements https://w3c.github.io/mediacapture-main/#dom-mediastream-getvideotracks
func (m *mediaStream) GetVideoTracks() []*webrtc.Track {
	return m.queryTracks(webrtc.RTPCodecTypeVideo)
}

// GetTracks implements https://w3c.github.io/mediacapture-main/#dom-mediastream-gettracks
func (m *mediaStream) GetTracks() []*webrtc.Track {
	return m.queryTracks(rtpCodecTypeDefault)
}

// queryTracks returns all tracks that are the same kind as t.
// If t is 0, which is the default, queryTracks will return all the tracks.
func (m *mediaStream) queryTracks(t webrtc.RTPCodecType) []*webrtc.Track {
	m.l.RLock()
	defer m.l.RUnlock()

	result := make([]*webrtc.Track, 0)
	for _, track := range m.tracks {
		if track.Kind() == t || t == rtpCodecTypeDefault {
			result = append(result, track)
		}
	}

	return result
}

// AddTrack implements https://w3c.github.io/mediacapture-main/#dom-mediastream-addtrack
func (m *mediaStream) AddTrack(t *webrtc.Track) {
	m.l.Lock()
	defer m.l.Unlock()

	id := t.ID()
	if _, ok := m.tracks[id]; ok {
		return
	}

	m.tracks[id] = t
}

// RemoveTrack implements https://w3c.github.io/mediacapture-main/#dom-mediastream-removetrack
func (m *mediaStream) RemoveTrack(t *webrtc.Track) {
	m.l.Lock()
	defer m.l.Unlock()

	delete(m.tracks, t.ID())
}
