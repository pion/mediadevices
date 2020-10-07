package mediadevices

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/driver"
	_ "github.com/pion/mediadevices/pkg/driver/audiotest"
	_ "github.com/pion/mediadevices/pkg/driver/videotest"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

func TestGetUserMedia(t *testing.T) {
	videoParams := mockParams{
		BaseParams: codec.BaseParams{
			BitRate: 100000,
		},
		name: "MockVideo",
	}
	audioParams := mockParams{
		BaseParams: codec.BaseParams{
			BitRate: 32000,
		},
		name: "MockAudio",
	}
	md := NewMediaDevicesFromCodecs(
		map[webrtc.RTPCodecType][]*webrtc.RTPCodec{
			webrtc.RTPCodecTypeVideo: {
				{Type: webrtc.RTPCodecTypeVideo, Name: "MockVideo", PayloadType: 1},
			},
			webrtc.RTPCodecTypeAudio: {
				{Type: webrtc.RTPCodecTypeAudio, Name: "MockAudio", PayloadType: 2},
			},
		},
		WithTrackGenerator(
			func(_ uint8, _ uint32, id, _ string, codec *webrtc.RTPCodec) (
				LocalTrack, error,
			) {
				return newMockTrack(codec, id), nil
			},
		),
	)
	constraints := MediaStreamConstraints{
		Video: func(c *MediaTrackConstraints) {
			c.Enabled = true
			c.Width = prop.Int(640)
			c.Height = prop.Int(480)
			params := videoParams
			c.VideoEncoderBuilders = []codec.VideoEncoderBuilder{&params}
		},
		Audio: func(c *MediaTrackConstraints) {
			c.Enabled = true
			params := audioParams
			c.AudioEncoderBuilders = []codec.AudioEncoderBuilder{&params}
		},
	}
	constraintsWrong := MediaStreamConstraints{
		Video: func(c *MediaTrackConstraints) {
			c.Enabled = true
			c.Width = prop.Int(640)
			c.Height = prop.Int(480)
			params := videoParams
			params.BitRate = 0
			c.VideoEncoderBuilders = []codec.VideoEncoderBuilder{&params}
		},
		Audio: func(c *MediaTrackConstraints) {
			c.Enabled = true
			params := audioParams
			c.AudioEncoderBuilders = []codec.AudioEncoderBuilder{&params}
		},
	}

	// GetUserMedia with broken parameters
	ms, err := md.GetUserMedia(constraintsWrong)
	if err == nil {
		t.Fatal("Expected error, but got nil")
	}

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
	for _, track := range tracks {
		track.Stop()
	}
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

func (params *mockParams) RTPCodec() *codec.RTPCodec {
	rtpCodec := codec.NewRTPH264Codec(90000)
	rtpCodec.Name = params.name
	return rtpCodec
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

func TestSelectBestDriverConstraintsResultIsSetProperly(t *testing.T) {
	filterFn := driver.FilterVideoRecorder()
	drivers := driver.GetManager().Query(filterFn)
	if len(drivers) == 0 {
		t.Fatal("expect to get at least 1 driver")
	}

	driver := drivers[0]
	err := driver.Open()
	if err != nil {
		t.Fatal("expect to open driver successfully")
	}
	defer driver.Close()

	if len(driver.Properties()) == 0 {
		t.Fatal("expect to get at least 1 property")
	}
	expectedProp := driver.Properties()[0]
	// Since this is a continuous value, bestConstraints should be set with the value that user specified
	expectedProp.FrameRate = 30.0

	wantConstraints := MediaTrackConstraints{
		MediaConstraints: prop.MediaConstraints{
			VideoConstraints: prop.VideoConstraints{
				// By reducing the width from the driver by a tiny amount, this property should be chosen.
				// At the same time, we'll be able to find out if the return constraints will be properly set
				// to the best constraints.
				Width:       prop.Int(expectedProp.Width - 1),
				Height:      prop.Int(expectedProp.Width),
				FrameFormat: prop.FrameFormat(expectedProp.FrameFormat),
				FrameRate:   prop.Float(30.0),
			},
		},
	}

	bestDriver, bestConstraints, err := selectBestDriver(filterFn, wantConstraints)
	if err != nil {
		t.Fatal(err)
	}

	if driver != bestDriver {
		t.Fatal("best driver is not expected")
	}

	s := bestConstraints.selectedMedia
	if s.Width != expectedProp.Width ||
		s.Height != expectedProp.Height ||
		s.FrameFormat != expectedProp.FrameFormat ||
		s.FrameRate != expectedProp.FrameRate {
		t.Fatalf("failed to return best constraints\nexpected:\n%v\n\ngot:\n%v", expectedProp, bestConstraints.selectedMedia)
	}
}
