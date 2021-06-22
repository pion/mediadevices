package mediadevices

import (
	"fmt"

	"github.com/pion/mediadevices/pkg/driver"
)

// MediaDeviceType enumerates type of media device.
type MediaDeviceType int

// MediaDeviceType definitions.
const (
	VideoInput MediaDeviceType = iota + 1
	AudioInput
	AudioOutput
)

func (mdt MediaDeviceType) String() string {
	switch mdt {
	case VideoInput:
		return "VideoInput"
	case AudioInput:
		return "AudioInput"
	case AudioOutput:
		return "AudioOutput"
	}
	return "Unknown"
}

// MediaDeviceInfo represents https://w3c.github.io/mediacapture-main/#dom-mediadeviceinfo
type MediaDeviceInfo struct {
	DeviceID     string
	Kind         MediaDeviceType
	Name         string
	Manufacturer string
	ModelID      string
	Label        string
	DeviceType   driver.DeviceType
}

func (mdi MediaDeviceInfo) String() string {
	return fmt.Sprintf("'%s' %s %v %s %s [%s %s]", mdi.Name, mdi.Label, mdi.DeviceID, mdi.Kind, mdi.DeviceType, mdi.Manufacturer, mdi.ModelID)
}
func (mdi MediaDeviceInfo) Serialize() string {
	return fmt.Sprintf("d/%s/%v/%s/%s/%s/%s", mdi.DeviceID, mdi.Kind, mdi.DeviceType, mdi.Name, mdi.Manufacturer, mdi.ModelID)
}
