package avfoundation

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AVFoundation -framework Foundation
#include <stdlib.h>
#include "deviceobserver.h"

extern void deviceEventBridge(void *userData, DeviceEventType eventType, DeviceInfo *device);

static const char* DeviceObserverInitWithBridge() {
    return DeviceObserverInit(deviceEventBridge, NULL);
}
*/
import "C"
import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

var (
	observerLock   sync.Mutex
	deviceCache    = make(map[string]Device)
	observing      bool
	stopObserver   chan struct{}
	onDeviceChange func(Device, DeviceEventType)
)

type DeviceEventType int

const (
	DeviceEventConnected    DeviceEventType = C.DeviceEventConnected
	DeviceEventDisconnected DeviceEventType = C.DeviceEventDisconnected
)

func SetOnDeviceChange(f func(Device, DeviceEventType)) {
	observerLock.Lock()
	defer observerLock.Unlock()
	onDeviceChange = f
}

//export goDeviceEventCallback
func goDeviceEventCallback(userData unsafe.Pointer, eventType C.int, device *C.DeviceInfo) {
	uid := C.GoString(&device.uid[0])
	name := C.GoString(&device.name[0])

	observerLock.Lock()
	defer observerLock.Unlock()

	d := createDevice(uid, name)
	et := DeviceEventType(eventType)

	if eventType == C.DeviceEventConnected {
		deviceCache[uid] = d
	} else if eventType == C.DeviceEventDisconnected {
		delete(deviceCache, uid)
	}

	if onDeviceChange != nil {
		onDeviceChange(d, et)
	}
}

// StartObserver starts the device observer and a background run loop
func StartObserver() error {
	observerLock.Lock()
	if observing {
		observerLock.Unlock()
		return nil
	}
	observerLock.Unlock() // Unlock to allow goroutine to acquire it during populate

	initErrCh := make(chan error)
	stopObserver = make(chan struct{})

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		if status := C.DeviceObserverInitWithBridge(); status != nil {
			initErrCh <- fmt.Errorf("failed to init observer: %s", C.GoString(status))
			return
		}

		// Start
		if status := C.DeviceObserverStart(); status != nil {
			initErrCh <- fmt.Errorf("failed to start observer: %s", C.GoString(status))
			return
		}

		// Initial population safely
		var devices [C.MAX_DEVICES]C.DeviceInfo
		var count C.int
		status := C.DeviceObserverGetDevices(&devices[0], &count)

		observerLock.Lock()
		if status == nil {
			// Clear existing cache to be sure
			deviceCache = make(map[string]Device)
			for i := 0; i < int(count); i++ {
				uid := C.GoString(&devices[i].uid[0])
				name := C.GoString(&devices[i].name[0])
				deviceCache[uid] = createDevice(uid, name)
			}
		}
		observing = true
		observerLock.Unlock()

		// Signal success
		close(initErrCh)

		// Run Loop
		for {
			select {
			case <-stopObserver:
				C.DeviceObserverStop()
				C.DeviceObserverDestroy()
				return
			default:
				// Run for small intervals to allow checking stop channel
				C.DeviceObserverRunFor(0.1)
			}
		}
	}()

	// Wait for initialization
	return <-initErrCh
}

// StopObserver stops the device observer
func StopObserver() error {
	observerLock.Lock()
	defer observerLock.Unlock()

	if !observing {
		return nil
	}

	close(stopObserver)
	observing = false

	// The cleanup happens in the background goroutine
	return nil
}

// GetDevices returns the currently connected video devices.
// If the observer is running, it returns the cached state.
// If not, it queries the system directly.
func GetDevices() ([]Device, error) {
	observerLock.Lock()
	defer observerLock.Unlock()

	if observing {
		var result []Device
		for _, d := range deviceCache {
			result = append(result, d)
		}
		return result, nil
	}

	// Fallback to direct query if not observing
	var devices [C.MAX_DEVICES]C.DeviceInfo
	var count C.int
	status := C.DeviceObserverGetDevices(&devices[0], &count)
	if status != nil {
		return nil, fmt.Errorf("%s", C.GoString(status))
	}

	result := make([]Device, 0, count)
	for i := 0; i < int(count); i++ {
		uid := C.GoString(&devices[i].uid[0])
		name := C.GoString(&devices[i].name[0])
		result = append(result, createDevice(uid, name))
	}
	return result, nil
}
