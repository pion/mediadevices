package driver

import (
	"fmt"
	"testing"
	"time"

	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

var (
	recordErr = fmt.Errorf("failed to start recording")
)

type adapterMock struct{}

func (a *adapterMock) Open() error              { return nil }
func (a *adapterMock) Close() error             { return nil }
func (a *adapterMock) Properties() []prop.Media { return []prop.Media{{}} }

type videoAdapterMock struct{ adapterMock }

func (a *videoAdapterMock) VideoRecord(p prop.Media) (r video.Reader, err error) { return nil, nil }

type videoAdapterBrokenMock struct{ adapterMock }

func (a *videoAdapterBrokenMock) VideoRecord(p prop.Media) (r video.Reader, err error) {
	return nil, recordErr
}

type videoAdapterWithProperties struct {
	videoAdapterMock
	properties []prop.Media
}

func (a *videoAdapterWithProperties) Properties() []prop.Media {
	return a.properties
}

type audioAdapterMock struct{ adapterMock }

func (a *audioAdapterMock) AudioRecord(p prop.Media) (r audio.Reader, err error) { return nil, nil }

type audioAdapterBrokenMock struct{ adapterMock }

func (a *audioAdapterBrokenMock) AudioRecord(p prop.Media) (r audio.Reader, err error) {
	return nil, recordErr
}

type audioAdapterWithProperties struct {
	audioAdapterMock
	properties []prop.Media
}

func (a *audioAdapterWithProperties) Properties() []prop.Media {
	return a.properties
}

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

	_, err = vr.VideoRecord(d.Properties()[0])
	if err != nil {
		t.Errorf("expected to successfully start recording, but got %v", err)
	}
}

func TestVideoWrapperWithBrokenRecorderState(t *testing.T) {
	var a videoAdapterBrokenMock
	d := wrapAdapter(&a, Info{})

	err := d.Open()
	if err != nil {
		t.Errorf("expected to open successfully")
	}

	vr := d.(VideoRecorder)
	_, err = vr.VideoRecord(d.Properties()[0])
	if err == nil {
		t.Errorf("expected to get an error")
	}

	if err != recordErr {
		t.Errorf("expected to get %v, but got %v", recordErr, err)
	}

	if d.Status() != StateClosed {
		t.Errorf("expected the status to be %v, but got %v", StateClosed, d.Status())
	}
}

func TestVideoWrapperWithProperties(t *testing.T) {
	var a videoAdapterWithProperties
	newProp := func(width, height int, frameRate float32, frameFormat frame.Format) prop.Media {
		var p prop.Media
		p.Width = width
		p.Height = height
		p.FrameRate = frameRate
		p.FrameFormat = frameFormat
		return p
	}
	cases := map[string]struct {
		haveProps []prop.Media
		wantProp  prop.Media
		expectErr bool
	}{
		"Invalid prop 1": {
			haveProps: []prop.Media{
				newProp(2, 2, 3.0, frame.FormatI420),
			},
			wantProp:  newProp(1, 2, 3.0, frame.FormatI420),
			expectErr: true,
		},
		"Invalid prop 2": {
			haveProps: []prop.Media{
				newProp(1, 2, 4.0, frame.FormatI420),
			},
			wantProp:  newProp(1, 2, 3.0, frame.FormatI420),
			expectErr: true,
		},
		"Invalid prop 3": {
			haveProps: []prop.Media{
				newProp(1, 2, 3.0, frame.FormatI444),
			},
			wantProp:  newProp(1, 2, 3.0, frame.FormatI420),
			expectErr: true,
		},
		"Invalid prop 4": {
			haveProps: nil,
			wantProp:  newProp(1, 2, 3.0, frame.FormatI420),
			expectErr: true,
		},
		"Valid prop": {
			haveProps: []prop.Media{
				newProp(1, 2, 3.0, frame.FormatI420),
			},
			wantProp:  newProp(1, 2, 3.0, frame.FormatI420),
			expectErr: false,
		},
	}

	d := wrapAdapter(&a, Info{})
	ar := d.(VideoRecorder)
	for testCaseName, testCase := range cases {
		t.Run(testCaseName, func(t *testing.T) {
			err := d.Open()
			if err != nil {
				t.Fatalf("expect to open successfully. But, failed with %s", err)
			}

			a.properties = testCase.haveProps
			// TODO: Find a better way to get the device id
			var deviceID string
			if len(d.Properties()) > 0 {
				deviceID = d.Properties()[0].DeviceID
			}
			testCase.wantProp.DeviceID = deviceID
			_, err = ar.VideoRecord(testCase.wantProp)
			if testCase.expectErr && err == nil {
				t.Fatalf("expect an error but got nil")
			} else if !testCase.expectErr && err != nil {
				t.Fatalf("expect no error but got %s", err)
			} else if !testCase.expectErr {
				// since it successfully opened, we need to close it
				d.Close()
			}
		})
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

	_, err = ar.AudioRecord(d.Properties()[0])
	if err != nil {
		t.Errorf("expected to successfully start recording, but got %v", err)
	}
}

func TestAudioWrapperWithBrokenRecorderState(t *testing.T) {
	var a audioAdapterBrokenMock
	d := wrapAdapter(&a, Info{})

	err := d.Open()
	if err != nil {
		t.Errorf("expected to open successfully")
	}

	ar := d.(AudioRecorder)
	_, err = ar.AudioRecord(d.Properties()[0])
	if err == nil {
		t.Errorf("expected to get an error")
	}

	if err != recordErr {
		t.Errorf("expected to get %v, but got %v", recordErr, err)
	}

	if d.Status() != StateClosed {
		t.Errorf("expected the status to be %v, but got %v", StateClosed, d.Status())
	}
}

func TestAudioWrapperWithProperties(t *testing.T) {
	newProp := func(channelCount int, latency time.Duration, sampleRate int, sampleSize int) prop.Media {
		var p prop.Media
		p.ChannelCount = channelCount
		p.Latency = latency
		p.SampleRate = sampleRate
		p.SampleSize = sampleSize
		return p
	}
	cases := map[string]struct {
		haveProps []prop.Media
		wantProp  prop.Media
		expectErr bool
	}{
		"Invalid prop 1": {
			haveProps: []prop.Media{
				newProp(1, time.Second, 3, 5),
			},
			wantProp:  newProp(1, time.Second, 3, 4),
			expectErr: true,
		},
		"Invalid prop 2": {
			haveProps: []prop.Media{
				newProp(1, time.Minute, 3, 4),
			},
			wantProp:  newProp(1, time.Second, 3, 4),
			expectErr: true,
		},
		"Invalid prop 3": {
			haveProps: nil,
			wantProp:  newProp(1, time.Second, 3, 4),
			expectErr: true,
		},
		"Valid prop": {
			haveProps: []prop.Media{
				newProp(1, time.Second, 3, 4),
			},
			wantProp:  newProp(1, time.Second, 3, 4),
			expectErr: false,
		},
	}

	var a audioAdapterWithProperties
	d := wrapAdapter(&a, Info{})
	ar := d.(AudioRecorder)
	for testCaseName, testCase := range cases {
		t.Run(testCaseName, func(t *testing.T) {
			err := d.Open()
			if err != nil {
				t.Fatalf("expect to open successfully. But, failed with %s", err)
			}

			a.properties = testCase.haveProps
			// TODO: Find a better way to get the device id
			var deviceID string
			if len(d.Properties()) > 0 {
				deviceID = d.Properties()[0].DeviceID
			}
			testCase.wantProp.DeviceID = deviceID
			_, err = ar.AudioRecord(testCase.wantProp)
			if testCase.expectErr && err == nil {
				t.Fatalf("expect an error but got nil")
			} else if !testCase.expectErr && err != nil {
				t.Fatalf("expect no error but got %s", err)
			} else if !testCase.expectErr {
				// since it successfully opened, we need to close it
				d.Close()
			}
		})
	}
}
