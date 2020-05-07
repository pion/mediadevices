package mediadevices

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"

	"github.com/pion/mediadevices/pkg/codec"
	_ "github.com/pion/mediadevices/pkg/driver/audiotest"
	_ "github.com/pion/mediadevices/pkg/driver/videotest"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

func TestGetUserMedia(t *testing.T) {
	brokenVideoParams := mockParams{
		name: "MockVideo",
	}
	videoParams := brokenVideoParams
	videoParams.BitRate = 100000
	audioParams := mockParams{
		BaseParams: codec.BaseParams{
			BitRate: 32000,
		},
		name: "MockAudio",
	}
	constraints := MediaStreamConstraints{
		Video: func(p *prop.Media) {
			p.Width = 640
			p.Height = 480
		},
		Audio: func(p *prop.Media) {},
	}

	md := NewMediaDevicesFromCodecs(
		map[webrtc.RTPCodecType][]*webrtc.RTPCodec{
			webrtc.RTPCodecTypeVideo: []*webrtc.RTPCodec{
				&webrtc.RTPCodec{Type: webrtc.RTPCodecTypeVideo, Name: "MockVideo", PayloadType: 1},
			},
			webrtc.RTPCodecTypeAudio: []*webrtc.RTPCodec{
				&webrtc.RTPCodec{Type: webrtc.RTPCodecTypeAudio, Name: "MockAudio", PayloadType: 2},
			},
		},
		WithTrackGenerator(
			func(_ uint8, _ uint32, id, _ string, codec *webrtc.RTPCodec) (
				LocalTrack, error,
			) {
				return newMockTrack(codec, id), nil
			},
		),
		WithVideoEncoders(&brokenVideoParams),
		WithAudioEncoders(&audioParams),
	)

	// GetUserMedia with broken parameters
	ms, err := md.GetUserMedia(constraints)
	if err == nil {
		t.Fatal("Expected error, but got nil")
	}

	md = NewMediaDevicesFromCodecs(
		map[webrtc.RTPCodecType][]*webrtc.RTPCodec{
			webrtc.RTPCodecTypeVideo: []*webrtc.RTPCodec{
				&webrtc.RTPCodec{Type: webrtc.RTPCodecTypeVideo, Name: "MockVideo", PayloadType: 1},
			},
			webrtc.RTPCodecTypeAudio: []*webrtc.RTPCodec{
				&webrtc.RTPCodec{Type: webrtc.RTPCodecTypeAudio, Name: "MockAudio", PayloadType: 2},
			},
		},
		WithTrackGenerator(
			func(_ uint8, _ uint32, id, _ string, codec *webrtc.RTPCodec) (
				LocalTrack, error,
			) {
				return newMockTrack(codec, id), nil
			},
		),
		WithVideoEncoders(&videoParams),
		WithAudioEncoders(&audioParams),
	)

	// GetUserMedia with correct parameters
	ms, err = md.GetUserMedia(constraints)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	tracks := ms.GetTracks()
	if l := len(tracks); l != 2 {
		t.Fatalf("Number of the tracks is expected to be 2, got %d", l)
	}
	for _, track := range tracks {
		track.OnEnded(func(err error) {
			if err != io.EOF {
				t.Errorf("OnEnded called: %v", err)
			}
		})
	}
	time.Sleep(50 * time.Millisecond)

	for _, track := range tracks {
		track.Stop()
	}

	// Stop and retry GetUserMedia
	ms, err = md.GetUserMedia(constraints)
	if err != nil {
		t.Fatalf("Failed to GetUserMedia after the previsous tracks stopped: %v", err)
	}
	tracks = ms.GetTracks()
	if l := len(tracks); l != 2 {
		t.Fatalf("Number of the tracks is expected to be 2, got %d", l)
	}
	for _, track := range tracks {
		track.OnEnded(func(err error) {
			if err != io.EOF {
				t.Errorf("OnEnded called: %v", err)
			}
		})
	}
	time.Sleep(50 * time.Millisecond)
}

type mockTrack struct {
	codec *webrtc.RTPCodec
	id    string
}

func newMockTrack(codec *webrtc.RTPCodec, id string) *mockTrack {
	return &mockTrack{
		codec: codec,
		id:    id,
	}
}

func (t *mockTrack) WriteSample(s media.Sample) error {
	return nil
}

func (t *mockTrack) Codec() *webrtc.RTPCodec {
	return t.codec
}

func (t *mockTrack) ID() string {
	return t.id
}

func (t *mockTrack) Kind() webrtc.RTPCodecType {
	return t.codec.Type
}

type mockParams struct {
	codec.BaseParams
	name string
}

func (params *mockParams) Name() string {
	return params.name
}

func (params *mockParams) BuildVideoEncoder(r video.Reader, p prop.Media) (codec.ReadCloser, error) {
	if params.BitRate == 0 {
		// This is a dummy error to test the failure condition.
		return nil, errors.New("wrong codec parameter")
	}
	return &mockVideoCodec{
		r:      r,
		closed: make(chan struct{}),
	}, nil
}

func (params *mockParams) BuildAudioEncoder(r audio.Reader, p prop.Media) (codec.ReadCloser, error) {
	return &mockAudioCodec{
		r:      r,
		closed: make(chan struct{}),
	}, nil
}

type mockCodec struct{}

func (e *mockCodec) SetBitRate(b int) error {
	return nil
}

func (e *mockCodec) ForceKeyFrame() error {
	return nil
}

type mockVideoCodec struct {
	mockCodec
	r      video.Reader
	closed chan struct{}
}

func (m *mockVideoCodec) Read(b []byte) (int, error) {
	if _, err := m.r.Read(); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (m *mockVideoCodec) Close() error { return nil }

type mockAudioCodec struct {
	mockCodec
	r      audio.Reader
	closed chan struct{}
}

func (m *mockAudioCodec) Read(b []byte) (int, error) {
	if _, err := m.r.Read(); err != nil {
		return 0, err
	}
	return len(b), nil
}
func (m *mockAudioCodec) Close() error { return nil }
