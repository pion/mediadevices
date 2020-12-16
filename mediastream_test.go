package mediadevices

import (
	"io"
	"testing"

	"github.com/pion/webrtc/v3"
)

type mockMediaStreamTrack struct {
	kind MediaDeviceType
}

func (track *mockMediaStreamTrack) ID() string {
	return ""
}

func (track *mockMediaStreamTrack) StreamID() string {
	return ""
}

func (track *mockMediaStreamTrack) Close() error {
	return nil
}

func (track *mockMediaStreamTrack) Kind() webrtc.RTPCodecType {
	switch track.kind {
	case AudioInput:
		return webrtc.RTPCodecTypeAudio
	case VideoInput:
		return webrtc.RTPCodecTypeVideo
	default:
		panic("invalid track kind")
	}
}

func (track *mockMediaStreamTrack) OnEnded(handler func(error)) {
}

func (track *mockMediaStreamTrack) Bind(ctx webrtc.TrackLocalContext) (webrtc.RTPCodecParameters, error) {
	return webrtc.RTPCodecParameters{}, nil
}

func (track *mockMediaStreamTrack) Unbind(ctx webrtc.TrackLocalContext) error {
	return nil
}

func (track *mockMediaStreamTrack) NewRTPReader(codecName string, ssrc uint32, mtu int) (RTPReadCloser, error) {
	return nil, nil
}

func (track *mockMediaStreamTrack) NewEncodedReader(codecName string) (EncodedReadCloser, error) {
	return nil, nil
}

func (track *mockMediaStreamTrack) NewEncodedIOReader(codecName string) (io.ReadCloser, error) {
	return nil, nil
}

func TestMediaStreamFilters(t *testing.T) {
	audioTracks := []Track{
		&mockMediaStreamTrack{AudioInput},
		&mockMediaStreamTrack{AudioInput},
		&mockMediaStreamTrack{AudioInput},
		&mockMediaStreamTrack{AudioInput},
		&mockMediaStreamTrack{AudioInput},
	}

	videoTracks := []Track{
		&mockMediaStreamTrack{VideoInput},
		&mockMediaStreamTrack{VideoInput},
		&mockMediaStreamTrack{VideoInput},
	}

	tracks := append(audioTracks, videoTracks...)
	stream, err := NewMediaStream(tracks...)
	if err != nil {
		t.Fatal(err)
	}

	expect := func(t *testing.T, actual, expected []Track) {
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
