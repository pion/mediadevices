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
	observerInitial   observerStateType = iota
	observerSetup                       // KVO initialized on main thread but not pumping run loop
	observerStarting                    // Starting run loop (transitioning to running)
	observerRunning                     // Run loop is actively pumping
	observerDestroyed                   // Destroyed and cannot be restarted
)

// deviceObserver manages the AVFoundation device observer lifecycle with the singleton pattern.
// The observer is single-use. Once DestroyObserver is called, it cannot be restarted.
type deviceObserver struct {
	// mu protects all non-channel, non-waitgroup observer state fields.
	// Must not be held when invoking user callbacks to avoid deadlock (double lock acquisition).
	mu             sync.Mutex
	deviceCache    map[string]Device
	state          observerStateType
	onDeviceChange func(Device, DeviceEventType)

	// Signals the goroutine to start pumping NSRunLoop
	startPumping chan struct{}
	// Signals destroy to the observer goroutine
	destroyObserver chan struct{}
	// Coordinates waiting for the observer goroutine to complete
	wg sync.WaitGroup
	// Allows concurrent callers to wait on same setup result
	setupDone chan struct{}
	// setupErr is the error returned by the setup goroutine. It is protected by mu
	setupErr error
	// Allows concurrent StartObserver callers to wait on same start result
	startDone chan struct{}
}

var (
	observerSingleton     *deviceObserver
	observerSingletonOnce sync.Once
)

func getObserver() *deviceObserver {
	observerSingletonOnce.Do(func() {
		observerSingleton = &deviceObserver{
			deviceCache: make(map[string]Device),
			state:       observerInitial,
		}
	})
	return observerSingleton
}

type DeviceEventType int

const (
	DeviceEventConnected    DeviceEventType = C.DeviceEventConnected
	DeviceEventDisconnected DeviceEventType = C.DeviceEventDisconnected
)

func SetOnDeviceChange(f func(Device, DeviceEventType)) {
	obs := getObserver()
	obs.mu.Lock()
	defer obs.mu.Unlock()
	obs.onDeviceChange = f
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

	obs := getObserver()
	obs.mu.Lock()
	if eventType == C.DeviceEventConnected {
		obs.deviceCache[uid] = d
	} else if eventType == C.DeviceEventDisconnected {
		delete(obs.deviceCache, uid)
	}
	cb := obs.onDeviceChange
	obs.mu.Unlock()

	if cb != nil {
		cb(d, et)
	}
}

// setup initializes the device observer and starts a goroutine locked to a thread for NSRunLoop,
// but does not begin pumping the run loop yet. The goroutine waits idle until start is called.
func (obs *deviceObserver) setup() error {
	obs.mu.Lock()

	if obs.state == observerSetup || obs.state == observerStarting || obs.state == observerRunning {
		// Already setup or beyond
		obs.mu.Unlock()
		return nil
	}

	if obs.state == observerDestroyed {
		obs.mu.Unlock()
		return fmt.Errorf("device observer is single-use and was destroyed, so it cannot be restarted")
	}

	if obs.setupDone != nil {
		done := obs.setupDone
		obs.mu.Unlock()
		<-done
		obs.mu.Lock()
		err := obs.setupErr
		obs.mu.Unlock()
		return err
	}

	// We're first to setup, initialize the channels
	obs.setupDone = make(chan struct{})
	obs.setupErr = nil
	obs.startPumping = make(chan struct{})
	obs.startDone = make(chan struct{})
	obs.destroyObserver = make(chan struct{})
	obs.wg.Add(1)
	obs.mu.Unlock()

	go func() {
		// Since caller is expected to invoke setup from the main thread,
		// we can lock this bg goroutine to the main thread here to set up C observer on.
		defer obs.wg.Done()
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		var err error
		if status := C.DeviceObserverInitWithBridge(); status != nil {
			err = fmt.Errorf("failed to init observer: %s", C.GoString(status))
		} else if status := C.DeviceObserverStart(); status != nil {
			C.DeviceObserverDestroy() // remember to clean up C objects on error
			err = fmt.Errorf("failed to start observer: %s", C.GoString(status))
		}

		if err != nil {
			obs.mu.Lock()
			obs.state = observerInitial
			obs.setupErr = err
			obs.mu.Unlock()
			close(obs.setupDone) // unblock setupDone waiters with error
			return
		}

		// Populate device cache and prepare initial device list for callbacks
		var devices [C.MAX_DEVICES]C.DeviceInfo
		var count C.int
		status := C.DeviceObserverGetDevices(&devices[0], &count)
		var initialDevices []Device
		obs.mu.Lock()
		if status == nil {
			obs.deviceCache = make(map[string]Device)
			for i := 0; i < int(count); i++ {
				uid := C.GoString(&devices[i].uid[0])
				name := C.GoString(&devices[i].name[0])
				dev := createDevice(uid, name)
				obs.deviceCache[uid] = dev
				initialDevices = append(initialDevices, dev)
			}
		}
		obs.state = observerSetup
		obs.mu.Unlock()

		close(obs.setupDone) // Setup complete, unblock waiters

		// Wait for signal to start pumping or destroy
		select {
		case <-obs.destroyObserver:
			// Destroyed before pumping the run loop
			C.DeviceObserverStop()
			C.DeviceObserverDestroy()
			return
		case <-obs.startPumping:
			// Proceed to pump loop
		}

		// Transition to running
		obs.mu.Lock()
		cb := obs.onDeviceChange
		obs.state = observerRunning
		obs.mu.Unlock()

		close(obs.startDone)

		// Replay current devices so downstream observers register them
		if cb != nil {
			for _, dev := range initialDevices {
				cb(dev, DeviceEventConnected)
			}
		}

		for {
			select {
			case <-obs.destroyObserver:
				C.DeviceObserverStop()
				C.DeviceObserverDestroy()
				return
			default:
				C.DeviceObserverRunFor(0.1)
			}
		}
	}()

	// Wait for goroutine to complete setup
	<-obs.setupDone
	obs.mu.Lock()
	err := obs.setupErr
	obs.mu.Unlock()
	return err
}

// start signals the observer goroutine to begin pumping the run loop.
func (obs *deviceObserver) start() error {
	obs.mu.Lock()

	for {
		switch obs.state {
		case observerDestroyed:
			obs.mu.Unlock()
			return fmt.Errorf("cannot start observer: observer has been destroyed and cannot be restarted")
		case observerRunning:
			obs.mu.Unlock()
			return nil
		case observerStarting:
			// Another goroutine is starting the run loop; wait on same result
			done := obs.startDone
			obs.mu.Unlock()
			<-done
			return nil
		case observerInitial:
			// Need to setup first
			obs.mu.Unlock()
			if err := obs.setup(); err != nil {
				return err
			}
			obs.mu.Lock()
			continue // re-check state as it may have changed by another goroutine e.g. destroyed
		case observerSetup:
			// proceed to start
		}
		break
	}

	obs.state = observerStarting
	pump := obs.startPumping
	obs.mu.Unlock()

	close(pump)

	<-obs.startDone
	return nil
}

// destroy destroys the device observer and releases all C/Objective-C resources.
// The observer cannot be restarted after being destroyed.
func (obs *deviceObserver) destroy() error {
	obs.mu.Lock()

	switch obs.state {
	case observerInitial, observerDestroyed:
		obs.state = observerDestroyed
		obs.mu.Unlock()
		return nil
	case observerSetup, observerRunning:
		// Proceed to destroy
	case observerStarting:
		// Wait for transition to running
		done := obs.startDone
		obs.mu.Unlock()
		<-done
		obs.mu.Lock()
		// Proceed to destroy
	}

	destroy := obs.destroyObserver
	obs.mu.Unlock()

	close(destroy)
	obs.wg.Wait()

	obs.mu.Lock()
	obs.state = observerDestroyed
	obs.mu.Unlock()

	return nil
}

// SetupObserver initializes the device observer and starts a goroutine
// locked to a thread for NSRunLoop, but does not begin pumping the run loop yet.
// The goroutine waits idle until StartObserver is called, avoiding CPU overhead.
// Safe to call concurrently and idempotently.
func SetupObserver() error {
	return getObserver().setup()
}

// StartObserver signals the observer goroutine to begin pumping the run loop.
// If SetupObserver has not been called, StartObserver will call it first.
// Safe to call concurrently and idempotently.
func StartObserver() error {
	return getObserver().start()
}

// DestroyObserver destroys the device observer and releases all C/Objective-C resources.
// The observer is single-use and cannot be restarted after being destroyed.
// Safe to call concurrently and idempotently.
func DestroyObserver() error {
	return getObserver().destroy()
}

// LookupCachedDevice returns the cached device that matches the provided UID.
// The returned boolean indicates whether the device was present in the cache.
// Callers should verify IsObserverRunning before relying on the result.
func LookupCachedDevice(uid string) (Device, bool) {
	obs := getObserver()
	obs.mu.Lock()
	defer obs.mu.Unlock()

	dev, ok := obs.deviceCache[uid]
	return dev, ok
}

// IsObserverRunning reports whether the device observer has successfully started
// and populated the in-memory cache.
func IsObserverRunning() bool {
	obs := getObserver()
	obs.mu.Lock()
	defer obs.mu.Unlock()
	return obs.state == observerRunning
}
