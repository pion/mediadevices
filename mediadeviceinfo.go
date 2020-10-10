package mediadevices

import "github.com/pion/mediadevices/pkg/driver"

// MediaDeviceType enumerates type of media device.
type MediaDeviceType int

// MediaDeviceType definitions.
const (
	VideoInput MediaDeviceType = iota + 1
	AudioInput
	AudioOutput
)

// MediaDeviceInfo represents https://w3c.github.io/mediacapture-main/#dom-mediadeviceinfo
type MediaDeviceInfo struct {
	DeviceID   string
	Kind       MediaDeviceType
	Label      string
	DeviceType driver.DeviceType
}
