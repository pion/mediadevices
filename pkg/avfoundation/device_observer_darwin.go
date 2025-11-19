package avfoundation

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AVFoundation -framework Foundation -framework CoreMedia -framework CoreVideo
#include <stdlib.h>
#include <string.h>
#include "AVFoundationBind/DeviceObserver.h"

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

type observerStateType int

const (
	observerStopped observerStateType = iota
	observerStarting
	observerRunning
)

var (
	// observerLock protects all observer state variables.
	// Must NOT be held when invoking user callbacks to avoid deadlock (double lock acquisition).
	observerLock   sync.Mutex
	deviceCache    = make(map[string]Device)
	observerState  observerStateType
	onDeviceChange func(Device, DeviceEventType)

	// Signals stop to the observer goroutine
	stopObserver chan struct{}
	// Coordinates waiting for the observer goroutine to stop
	observerWg sync.WaitGroup
	// Allows concurrent StartObserver callers to wait on same init result
	initDone chan struct{}
	initErr  error
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

func createDevice(uid, name string) Device {
	var d Device
	d.UID = uid
	d.Name = name

	// Copy strings to C char arrays
	cUID := C.CString(uid)
	defer C.free(unsafe.Pointer(cUID))
	C.strncpy(&d.cDevice.uid[0], cUID, C.MAX_DEVICE_UID_CHARS)
	d.cDevice.uid[C.MAX_DEVICE_UID_CHARS] = 0

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	C.strncpy(&d.cDevice.name[0], cName, C.MAX_DEVICE_NAME_CHARS)
	d.cDevice.name[C.MAX_DEVICE_NAME_CHARS] = 0

	return d
}

//export goDeviceEventCallback
func goDeviceEventCallback(userData unsafe.Pointer, eventType C.int, device *C.DeviceInfo) {
	uid := C.GoString(&device.uid[0])
	name := C.GoString(&device.name[0])

	d := createDevice(uid, name)
	et := DeviceEventType(eventType)

	observerLock.Lock()
	if eventType == C.DeviceEventConnected {
		deviceCache[uid] = d
	} else if eventType == C.DeviceEventDisconnected {
		delete(deviceCache, uid)
	}
	cb := onDeviceChange
	observerLock.Unlock()

	if cb != nil {
		cb(d, et) // invoke outside of observerLock to avoid deadlock (double lock acquisition)
	}
}

// StartObserver starts the device observer and a background run loop.
// Safe to call concurrently; only one observer will be started.
func StartObserver() error {
	observerLock.Lock()

	switch observerState {
	case observerRunning:
		observerLock.Unlock()
		return nil
	case observerStarting:
		// Another goroutine is starting the observer; wait on same result
		done := initDone
		observerLock.Unlock()
		<-done
		observerLock.Lock()
		err := initErr
		observerLock.Unlock()
		return err
	case observerStopped:
		// start
	}

	observerState = observerStarting
	stopObserver = make(chan struct{})
	initDone = make(chan struct{})
	initErr = nil
	observerWg.Add(1)

	go func() {
		defer observerWg.Done()
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		var err error

		if status := C.DeviceObserverInitWithBridge(); status != nil {
			err = fmt.Errorf("failed to init observer: %s", C.GoString(status))
		} else if status := C.DeviceObserverStart(); status != nil {
			C.DeviceObserverDestroy() // prevents objective-c object leak on error
			err = fmt.Errorf("failed to start observer: %s", C.GoString(status))
		}

		if err != nil {
			observerLock.Lock()
			observerState = observerStopped
			initErr = err
			observerLock.Unlock()
			close(initDone) // unblock all waiters
			return
		}

		// Initial population
		var devices [C.MAX_DEVICES]C.DeviceInfo
		var count C.int
		status := C.DeviceObserverGetDevices(&devices[0], &count)

		var initialDevices []Device
		observerLock.Lock()
		if status == nil {
			deviceCache = make(map[string]Device)
			initialDevices = make([]Device, 0, int(count))
			for i := 0; i < int(count); i++ {
				uid := C.GoString(&devices[i].uid[0])
				name := C.GoString(&devices[i].name[0])
				dev := createDevice(uid, name)
				deviceCache[uid] = dev
				initialDevices = append(initialDevices, dev)
			}
		}
		cb := onDeviceChange
		observerState = observerRunning
		observerLock.Unlock()

		// Signal success to all waiters
		close(initDone)

		// Replay current devices so downstream observers register them.
		if cb != nil {
			for _, dev := range initialDevices {
				cb(dev, DeviceEventConnected)
			}
		}

		// Run Loop
		for {
			select {
			case <-stopObserver:
				C.DeviceObserverStop()
				C.DeviceObserverDestroy()
				return
			default:
				C.DeviceObserverRunFor(0.1)
			}
		}
	}()

	observerLock.Unlock()

	// Wait for initialization
	<-initDone
	observerLock.Lock()
	err := initErr
	observerLock.Unlock()
	return err
}

// StopObserver stops the device observer.
// Safe to call concurrently or when already stopped.
func StopObserver() error {
	observerLock.Lock()
	if observerState != observerRunning {
		observerLock.Unlock()
		return nil
	}

	close(stopObserver)
	observerState = observerStopped
	observerLock.Unlock()

	observerWg.Wait()

	return nil
}
