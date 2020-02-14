package mediadevices

import (
	"github.com/pion/mediadevices/pkg/driver"
	_ "github.com/pion/mediadevices/pkg/driver/camera"
	_ "github.com/pion/mediadevices/pkg/driver/microphone"
	_ "github.com/pion/mediadevices/pkg/driver/screen"
)

// RegisterDriverAdapter allows user space level of driver registration
func RegisterDriverAdapter(a driver.Adapter, info driver.Info) error {
	return driver.GetManager().Register(a, info)
}
