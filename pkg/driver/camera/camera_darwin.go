package camera

import (
	"context"
	"errors"
	"image"
	"io"
	"time"

	"github.com/pion/mediadevices/pkg/avfoundation"
	"github.com/pion/mediadevices/pkg/driver"
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
	manager := driver.GetManager()
	for _, d := range manager.Query(driver.FilterVideoRecorder()) {
		manager.Delete(d.ID())
	}

	devices, err := avfoundation.Devices(avfoundation.Video)
	if err != nil {
		panic(err)
	}

	for _, device := range devices {
		cam := newCamera(device)
		manager.Register(cam, driver.Info{
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
				manager.Delete(d.ID())
			}
		}
	})

	return avfoundation.StartObserver()
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
	if cam.rcClose != nil {
		cam.rcClose()
	}

	if cam.cancel != nil {
		cam.cancel()
	}

	return cam.session.Close()
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
