package mediadevices

import (
	"testing"

	"github.com/pion/webrtc/v2"
)

type mockMediaStreamTrack struct {
	kind MediaDeviceType
}

func (track *mockMediaStreamTrack) Track() *webrtc.Track {
	return nil
}

func (track *mockMediaStreamTrack) LocalTrack() LocalTrack {
	return nil
}

func (track *mockMediaStreamTrack) Stop() {
}

func (track *mockMediaStreamTrack) Kind() MediaDeviceType {
	return track.kind
}

func (track *mockMediaStreamTrack) OnEnded(handler func(error)) {
}

func TestMediaStreamFilters(t *testing.T) {
	audioTracks := []Tracker{
		&mockMediaStreamTrack{AudioInput},
		&mockMediaStreamTrack{AudioInput},
		&mockMediaStreamTrack{AudioInput},
		&mockMediaStreamTrack{AudioInput},
		&mockMediaStreamTrack{AudioInput},
	}

	videoTracks := []Tracker{
		&mockMediaStreamTrack{VideoInput},
		&mockMediaStreamTrack{VideoInput},
		&mockMediaStreamTrack{VideoInput},
	}

	tracks := append(audioTracks, videoTracks...)
	stream, err := NewMediaStream(tracks...)
	if err != nil {
		t.Fatal(err)
	}

	expect := func(t *testing.T, actual, expected []Tracker) {
		if len(actual) != len(expected) {
			t.Fatalf("%s: Expected to get %d trackers, but got %d trackers", t.Name(), len(expected), len(actual))
		}

		for _, a := range actual {
			found := false
			for _, e := range expected {
				if e == a {
					found = true
					break
				}
			}

			if !found {
				t.Fatalf("%s: Expected to find %p in the query results", t.Name(), a)
			}
		}
	}

	t.Run("GetAudioTracks", func(t *testing.T) {
		expect(t, stream.GetAudioTracks(), audioTracks)
	})

	t.Run("GetVideoTracks", func(t *testing.T) {
		expect(t, stream.GetVideoTracks(), videoTracks)
	})

	t.Run("GetTracks", func(t *testing.T) {
		expect(t, stream.GetTracks(), tracks)
	})
}
