package mediadevices

import (
	"io"
	"testing"
	"time"

	"github.com/pion/mediadevices/pkg/driver"
	_ "github.com/pion/mediadevices/pkg/driver/audiotest"
	_ "github.com/pion/mediadevices/pkg/driver/videotest"
	"github.com/pion/mediadevices/pkg/prop"
)

func TestGetUserMedia(t *testing.T) {
	constraints := MediaStreamConstraints{
		Video: func(c *MediaTrackConstraints) {
			c.Width = prop.Int(640)
			c.Height = prop.Int(480)
		},
		Audio: func(c *MediaTrackConstraints) {
		},
	}
	constraintsWrong := MediaStreamConstraints{
		Video: func(c *MediaTrackConstraints) {
			c.Width = prop.IntExact(10000)
			c.Height = prop.Int(480)
		},
		Audio: func(c *MediaTrackConstraints) {
		},
	}

	// GetUserMedia with broken parameters
	ms, err := GetUserMedia(constraintsWrong)
	if err == nil {
		t.Fatal("Expected error, but got nil")
	}

	// GetUserMedia with correct parameters
	ms, err = GetUserMedia(constraints)
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
		track.Close()
	}

	// Stop and retry GetUserMedia
	ms, err = GetUserMedia(constraints)
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
		track.Close()
	}
}

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
