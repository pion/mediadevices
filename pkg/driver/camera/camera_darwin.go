package camera

import (
	"context"
	"errors"
	"image"
	"io"
	"time"

	"github.com/pion/mediadevices/pkg/avfoundation"
	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/driver/availability"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

type camera struct {
	device  avfoundation.Device
	session *avfoundation.Session
	rcClose func()
	cancel  context.CancelFunc
}

const readTimeout = 3 * time.Second

func init() {
	Initialize()
}

// Initialize finds and registers camera devices. This is part of an experimental API.
func Initialize() {
	devices, err := avfoundation.Devices(avfoundation.Video)
	if err != nil {
		panic(err)
	}

	for _, device := range devices {
		cam := newCamera(device)
		driver.GetManager().Register(cam, driver.Info{
			Label:      device.UID,
			DeviceType: driver.Camera,
			Name:       device.Name,
		})
	}
}

// StartObserver starts the background observer to monitor for device changes.
func StartObserver() error {
	manager := driver.GetManager()

	avfoundation.SetOnDeviceChange(func(device avfoundation.Device, event avfoundation.DeviceEventType) {
		switch event {
		case avfoundation.DeviceEventConnected:
			drivers := manager.Query(func(d driver.Driver) bool {
				return d.Info().Label == device.UID
			})
			if len(drivers) > 0 {
				return
			}

			cam := newCamera(device)
			manager.Register(cam, driver.Info{
				Label:      device.UID,
				DeviceType: driver.Camera,
				Name:       device.Name,
			})

		case avfoundation.DeviceEventDisconnected:
			drivers := manager.Query(func(d driver.Driver) bool {
				return d.Info().Label == device.UID
			})
			for _, d := range drivers {
				status := d.Status()
				if status != driver.StateClosed {
					if err := d.Close(); err != nil {
					}
				}
				manager.Delete(d.ID())
			}
		}
	})

	if err := avfoundation.StartObserver(); err != nil {
		return err
	}

	return syncVideoRecorders(manager)
}

func StopObserver() error {
	return avfoundation.StopObserver()
}

func newCamera(device avfoundation.Device) *camera {
	return &camera{
		device: device,
	}
}

func (cam *camera) Open() error {
	var err error
	cam.session, err = avfoundation.NewSession(cam.device)
	return err
}

func (cam *camera) Close() error {
	if cam.cancel != nil {
		cam.cancel()
		cam.cancel = nil
	}

	if cam.rcClose != nil {
		cam.rcClose()
		cam.rcClose = nil
	}

	if cam.session != nil {
		err := cam.session.Close()
		cam.session = nil
		return err
	}

	return nil
}

func (cam *camera) VideoRecord(property prop.Media) (video.Reader, error) {
	decoder, err := frame.NewDecoder(property.FrameFormat)
	if err != nil {
		return nil, err
	}

	rc, err := cam.session.Open(property)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	cam.cancel = cancel
	cam.rcClose = rc.Close
	r := video.ReaderFunc(func() (image.Image, func(), error) {
		if ctx.Err() != nil {
			// Return EOF if the camera is already closed.
			return nil, func() {}, io.EOF
		}

		readCtx, cancel := context.WithTimeout(ctx, readTimeout)
		defer cancel()

		frame, _, err := rc.ReadContext(readCtx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				return nil, func() {}, io.EOF
			}
			return nil, func() {}, err
		}
		return decoder.Decode(frame, property.Width, property.Height)
	})
	return r, nil
}

func (cam *camera) Properties() []prop.Media {
	return cam.session.Properties()
}

func (cam *camera) IsAvailable() (bool, error) {
	// If this camera already has an active session, check if the device is still physically present
	if cam.session != nil {
		if avfoundation.IsObserverRunning() {
			if _, ok := avfoundation.LookupCachedDevice(cam.device.UID); !ok {
				// Device was physically unplugged while session was active
				// The session is now invalid
				return false, availability.ErrNoDevice
			}
		} else {
			// Observer not running, enumerate devices to check
			devices, err := avfoundation.Devices(avfoundation.Video)
			if err != nil {
				return false, err
			}
			deviceFound := false
			for _, device := range devices {
				if device.UID == cam.device.UID {
					deviceFound = true
					break
				}
			}
			if !deviceFound {
				// Device was physically unplugged while session was active
				return false, availability.ErrNoDevice
			}
		}
		// Device is still present, but session is active (in use by this camera)
		// Not available for others to use
		return false, errors.New("camera session already active")
	}

	// No active session, check if device is available
	if avfoundation.IsObserverRunning() {
		if _, ok := avfoundation.LookupCachedDevice(cam.device.UID); ok {
			return true, nil
		}
		return false, availability.ErrNoDevice
	}

	// If the observer is not running, fallback to stale device list from startup
	devices, err := avfoundation.Devices(avfoundation.Video)
	if err != nil {
		return false, err
	}

	for _, device := range devices {
		if device.UID == cam.device.UID {
			return true, nil
		}
	}

	return false, availability.ErrNoDevice
}

// syncVideoRecorders keeps the manager in lockstep with the hardware before the first user query.
func syncVideoRecorders(manager *driver.Manager) error {
	devices, err := avfoundation.Devices(avfoundation.Video)
	if err != nil {
		return err
	}

	current := make(map[string]struct{}, len(devices))
	for _, device := range devices {
		current[device.UID] = struct{}{}
	}

	registered := manager.Query(driver.FilterVideoRecorder())
	registeredByLabel := make(map[string]struct{}, len(registered))

	// drop any registered drivers whose UID isn't currently present
	for _, d := range registered {
		label := d.Info().Label
		registeredByLabel[label] = struct{}{}
		if _, ok := current[label]; !ok {
			manager.Delete(d.ID())
			delete(registeredByLabel, label)
		}
	}

	// register any new devices that appeared between the init() call and the observer start
	for _, device := range devices {
		if _, ok := registeredByLabel[device.UID]; ok {
			continue
		}

		cam := newCamera(device)
		manager.Register(cam, driver.Info{
			Label:      device.UID,
			DeviceType: driver.Camera,
			Name:       device.Name,
		})
		registeredByLabel[device.UID] = struct{}{}
	}

	return nil
}
