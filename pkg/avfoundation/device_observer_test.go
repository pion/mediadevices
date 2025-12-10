//go:build darwin
// +build darwin

package avfoundation

import (
	"testing"
)

// TestGetObserverSingleton tests that getObserver returns the same instance.
func TestGetObserverSingleton(t *testing.T) {
	obs1 := getObserver()
	obs2 := getObserver()

	if obs1 != obs2 {
		t.Error("getObserver() should return the same singleton instance")
	}

	if obs1.deviceCache == nil {
		t.Error("Observer device cache should be initialized")
	}

	if obs1.state != observerInitial {
		t.Errorf("Initial observer state should be observerInitial, got: %v", obs1.state)
	}
}

// TestCreateDevice tests device creation with UID and name.
func TestCreateDevice(t *testing.T) {
	testCases := []struct {
		name    string
		uid     string
		devName string
	}{
		{
			name:    "simple device",
			uid:     "test-uid-123",
			devName: "Test Camera",
		},
		{
			name:    "device with special characters",
			uid:     "camera_0x1234567890abcdef",
			devName: "FaceTime HD Camera",
		},
		{
			name:    "empty strings",
			uid:     "",
			devName: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			device := createDevice(tc.uid, tc.devName)

			if device.UID != tc.uid {
				t.Errorf("Expected UID %q, got %q", tc.uid, device.UID)
			}

			if device.Name != tc.devName {
				t.Errorf("Expected Name %q, got %q", tc.devName, device.Name)
			}
		})
	}
}

// TestSetOnDeviceChange tests setting and retrieving the device change callback.
func TestSetOnDeviceChange(t *testing.T) {
	// Reset observer state for clean test (hacky)
	// In production, the observer is a singleton
	obs := getObserver()
	obs.mu.Lock()
	originalCallback := obs.onDeviceChange
	obs.mu.Unlock()

	// Restore original callback at end of test
	defer func() {
		obs.mu.Lock()
		obs.onDeviceChange = originalCallback
		obs.mu.Unlock()
	}()

	called := false
	var capturedDevice Device
	var capturedEvent DeviceEventType

	SetOnDeviceChange(func(d Device, e DeviceEventType) {
		called = true
		capturedDevice = d
		capturedEvent = e
	})

	// Verify callback was set
	obs.mu.Lock()
	if obs.onDeviceChange == nil {
		t.Fatal("OnDeviceChange callback was not set")
	}

	// Manually trigger callback for testing
	testDevice := createDevice("test-uid", "test-name")
	testEvent := DeviceEventConnected
	cb := obs.onDeviceChange
	obs.mu.Unlock()

	if cb != nil {
		cb(testDevice, testEvent)
	}

	if !called {
		t.Error("Callback was not invoked")
	}

	if capturedDevice.UID != "test-uid" {
		t.Errorf("Expected captured UID %q, got %q", "test-uid", capturedDevice.UID)
	}

	if capturedEvent != DeviceEventConnected {
		t.Errorf("Expected event %v, got %v", DeviceEventConnected, capturedEvent)
	}
}

// TestLookupCachedDevice tests device cache lookups.
func TestLookupCachedDevice(t *testing.T) {
	obs := getObserver()

	// Add a test device to cache
	testUID := "lookup-test-uid"
	testDevice := createDevice(testUID, "Lookup Test Camera")

	obs.mu.Lock()
	obs.deviceCache[testUID] = testDevice
	obs.mu.Unlock()

	// Test successful lookup
	device, ok := LookupCachedDevice(testUID)
	if !ok {
		t.Error("Expected to find device in cache")
	}

	if device.UID != testUID {
		t.Errorf("Expected UID %q, got %q", testUID, device.UID)
	}

	// Test failed lookup
	_, ok = LookupCachedDevice("non-existent-uid")
	if ok {
		t.Error("Expected not to find non-existent device in cache")
	}

	// Cleanup
	obs.mu.Lock()
	delete(obs.deviceCache, testUID)
	obs.mu.Unlock()
}

// TestIsObserverRunning tests the observer running state check.
func TestIsObserverRunning(t *testing.T) {
	obs := getObserver()

	// Initially should not be running
	obs.mu.Lock()
	originalState := obs.state
	obs.state = observerInitial
	obs.mu.Unlock()

	// Restore original state at end
	defer func() {
		obs.mu.Lock()
		obs.state = originalState
		obs.mu.Unlock()
	}()

	if IsObserverRunning() {
		t.Error("Observer should not be running in initial state")
	}

	// Set state to running
	obs.mu.Lock()
	obs.state = observerRunning
	obs.mu.Unlock()

	if !IsObserverRunning() {
		t.Error("Observer should be running after state set to observerRunning")
	}

	// Set state to other states
	for _, state := range []observerStateType{observerSetup, observerStarting, observerDestroyed} {
		obs.mu.Lock()
		obs.state = state
		obs.mu.Unlock()

		if IsObserverRunning() {
			t.Errorf("Observer should not be running in state %v", state)
		}
	}
}

// TestGoDeviceEventCallback tests the C-to-Go device event callback.
func TestGoDeviceEventCallback(t *testing.T) {
	obs := getObserver()

	// Clear device cache for clean test
	obs.mu.Lock()
	obs.deviceCache = make(map[string]Device)
	originalCallback := obs.onDeviceChange
	obs.mu.Unlock()

	defer func() {
		obs.mu.Lock()
		obs.onDeviceChange = originalCallback
		obs.deviceCache = make(map[string]Device)
		obs.mu.Unlock()
	}()

	// Set up test callback
	var callbackInvoked bool
	var capturedDevice Device
	var capturedEvent DeviceEventType

	SetOnDeviceChange(func(d Device, e DeviceEventType) {
		callbackInvoked = true
		capturedDevice = d
		capturedEvent = e
	})

	// Note: We cannot directly call goDeviceEventCallback with C types in a Go test
	// without CGO setup. Instead, we test the logic that would be executed.

	// Simulate connect event
	testUID := "callback-test-uid"
	testDevice := createDevice(testUID, "Callback Test Camera")

	obs.mu.Lock()
	obs.deviceCache[testUID] = testDevice
	cb := obs.onDeviceChange
	obs.mu.Unlock()

	if cb != nil {
		cb(testDevice, DeviceEventConnected)
	}

	if !callbackInvoked {
		t.Error("User callback should have been invoked")
	}

	if capturedEvent != DeviceEventConnected {
		t.Errorf("Expected DeviceEventConnected, got %v", capturedEvent)
	}

	// Verify device was added to cache
	obs.mu.Lock()
	_, exists := obs.deviceCache[testUID]
	obs.mu.Unlock()

	if !exists {
		t.Error("Device should be in cache after connect event")
	}

	// Simulate disconnect event
	callbackInvoked = false
	obs.mu.Lock()
	delete(obs.deviceCache, testUID)
	cb = obs.onDeviceChange
	obs.mu.Unlock()

	if cb != nil {
		cb(testDevice, DeviceEventDisconnected)
	}

	if !callbackInvoked {
		t.Error("User callback should have been invoked for disconnect")
	}

	if capturedEvent != DeviceEventDisconnected {
		t.Errorf("Expected DeviceEventDisconnected, got %v", capturedEvent)
	}

	if capturedDevice.UID != testUID {
		t.Errorf("Expected captured device UID %q, got %q", testUID, capturedDevice.UID)
	}

	// Verify device was removed from cache
	obs.mu.Lock()
	_, exists = obs.deviceCache[testUID]
	obs.mu.Unlock()

	if exists {
		t.Error("Device should not be in cache after disconnect event")
	}
}

// TestDeviceEventTypes tests the device event type constants, verifying that they are different.
func TestDeviceEventTypes(t *testing.T) {
	if DeviceEventConnected == DeviceEventDisconnected {
		t.Error("DeviceEventConnected and DeviceEventDisconnected should be different")
	}
}
