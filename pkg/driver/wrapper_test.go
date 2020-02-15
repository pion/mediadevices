package driver

import (
	"testing"

	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

type adapterMock struct{}

func (a *adapterMock) Open() error              { return nil }
func (a *adapterMock) Close() error             { return nil }
func (a *adapterMock) Properties() []prop.Media { return []prop.Media{prop.Media{}} }

type videoAdapterMock struct{ adapterMock }

func (a *videoAdapterMock) VideoRecord(p prop.Media) (r video.Reader, err error) { return nil, nil }

type audioAdapterMock struct{ adapterMock }

func (a *audioAdapterMock) AudioRecord(p prop.Media) (r audio.Reader, err error) { return nil, nil }

func TestVideoWrapperState(t *testing.T) {
	var a videoAdapterMock
	d := wrapAdapter(&a, Info{})

	if d.Properties() != nil {
		t.Errorf("expected nil, but got %v", d.Properties())
	}

	vr := d.(VideoRecorder)
	_, err := vr.VideoRecord(prop.Media{})
	if err == nil {
		t.Errorf("expected to get an invalid state")
	}

	err = d.Open()
	if err != nil {
		t.Errorf("expected to successfully open, but got %v", err)
	}

	_, err = vr.VideoRecord(prop.Media{})
	if err != nil {
		t.Errorf("expected to successfully start recording, but got %v", err)
	}
}

func TestAudioWrapperState(t *testing.T) {
	var a audioAdapterMock
	d := wrapAdapter(&a, Info{})

	if d.Properties() != nil {
		t.Errorf("expected nil, but got %v", d.Properties())
	}

	ar := d.(AudioRecorder)
	_, err := ar.AudioRecord(prop.Media{})
	if err == nil {
		t.Errorf("expected to get an invalid state")
	}

	err = d.Open()
	if err != nil {
		t.Errorf("expected to successfully open, but got %v", err)
	}

	_, err = ar.AudioRecord(prop.Media{})
	if err != nil {
		t.Errorf("expected to successfully start recording, but got %v", err)
	}
}
