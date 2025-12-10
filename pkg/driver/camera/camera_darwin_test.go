//go:build darwin
// +build darwin

// $ go test -v . -tags darwin -run="^TestCameraFrameFormatSupport$"

package camera

import (
	"testing"

	"github.com/pion/mediadevices/pkg/avfoundation"
	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/driver/availability"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/prop"
)

func TestCameraFrameFormatSupport(t *testing.T) {
	devices, err := avfoundation.Devices(avfoundation.Video)
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) > 0 {
		c := newCamera(devices[0])
		if err := c.Open(); err != nil {
			t.Fatal(err)
		}
		defer c.Close()

		supportedFormats := make(map[frame.Format]struct{})
		for _, p := range c.Properties() {
			supportedFormats[p.FrameFormat] = struct{}{}
		}

		for _, format := range []frame.Format{
			frame.FormatI420,
			frame.FormatNV12,
			frame.FormatNV21,
			frame.FormatYUY2,
			frame.FormatUYVY,
		} {
			if _, ok := supportedFormats[format]; !ok {
				t.Logf("[%v] UNSUPPORTED", format)
				continue
			}
			r, err := c.VideoRecord(prop.Media{
				Video: prop.Video{
					Width:       640,
					Height:      480,
					FrameFormat: format,
				}})
			if err != nil {
				t.Logf("[%v] Failed to capture image: %v", format, err)
				continue
			}
			for i := 0; i < 10; i++ {
				_, _, err := r.Read()
				if err != nil {
					t.Logf("[%v] Failed to read: %v", format, err)
					continue
				}
			}
			t.Logf("[%v] OK", format)
		}
	}
}

// TestCameraCloseIdempotency tests that Close can be called multiple times safely.
func TestCameraCloseIdempotency(t *testing.T) {
	devices, err := avfoundation.Devices(avfoundation.Video)
	if err != nil {
		t.Fatal(err)
	}

	if len(devices) == 0 {
		t.Skip("No video devices available for testing")
	}

	cam := newCamera(devices[0])
	if err := cam.Open(); err != nil {
		t.Fatal(err)
	}

	// Close multiple times should not error
	for i := 0; i < 3; i++ {
		if err := cam.Close(); err != nil {
			t.Errorf("Close call %d failed: %v", i+1, err)
		}
	}

	// Verify internal state was cleared
	if cam.session != nil {
		t.Error("Session should be nil after close")
	}
	if cam.rcClose != nil {
		t.Error("rcClose should be nil after close")
	}
	if cam.cancel != nil {
		t.Error("cancel should be nil after close")
	}
}

// TestCameraIsAvailableObserverNotRunning tests IsAvailable when observer is not running.
func TestCameraIsAvailableObserverNotRunning(t *testing.T) {
	devices, err := avfoundation.Devices(avfoundation.Video)
	if err != nil {
		t.Fatal(err)
	}

	if len(devices) == 0 {
		t.Skip("No video devices available for testing")
	}

	cam := newCamera(devices[0])

	available, err := cam.IsAvailable()
	if available {
		t.Error("Camera should not be available when observer is not running")
	}

	if err != availability.ErrObserverUnavailable {
		t.Errorf("Expected ErrObserverUnavailable, got: %v", err)
	}
}

// TestNewCamera tests camera constructor.
func TestNewCamera(t *testing.T) {
	testDevice := avfoundation.Device{
		UID:  "test-uid",
		Name: "Test Camera",
	}

	cam := newCamera(testDevice)

	if cam == nil {
		t.Fatal("newCamera returned nil")
	}

	if cam.device.UID != testDevice.UID {
		t.Errorf("Expected device UID %q, got %q", testDevice.UID, cam.device.UID)
	}

	if cam.device.Name != testDevice.Name {
		t.Errorf("Expected device name %q, got %q", testDevice.Name, cam.device.Name)
	}
}

// TestSyncVideoRecorders tests the syncVideoRecorders function.
func TestSyncVideoRecorders(t *testing.T) {
	manager := driver.GetManager()

	// Initial state
	initialDrivers := manager.Query(driver.FilterVideoRecorder())
	initialCount := len(initialDrivers)

	// Run sync
	err := syncVideoRecorders(manager)
	if err != nil {
		t.Fatalf("syncVideoRecorders failed: %v", err)
	}

	// Verify drivers were synced
	afterDrivers := manager.Query(driver.FilterVideoRecorder())
	afterCount := len(afterDrivers)

	// The count should match the actual devices available
	devices, err := avfoundation.Devices(avfoundation.Video)
	if err != nil {
		t.Fatal(err)
	}

	if afterCount != len(devices) {
		t.Logf("Warning: Expected %d drivers after sync, got %d (initial: %d)",
			len(devices), afterCount, initialCount)
	}
}

// TestObserverFunctionsIdempotent tests that observer functions can be called multiple times.
func TestObserverFunctionsIdempotent(t *testing.T) {
	// This test may have side effects on the global observer state
	// In a real scenario, you'd want to reset the observer between tests

	// SetupObserver should be idempotent
	for i := 0; i < 2; i++ {
		if err := SetupObserver(); err != nil {
			t.Errorf("SetupObserver call %d failed: %v", i+1, err)
		}
	}

	// StartObserver should be idempotent
	for i := 0; i < 2; i++ {
		if err := StartObserver(); err != nil {
			t.Errorf("StartObserver call %d failed: %v", i+1, err)
		}
	}

	// Cleanup
	if err := DestroyObserver(); err != nil {
		t.Errorf("DestroyObserver failed: %v", err)
	}
}
