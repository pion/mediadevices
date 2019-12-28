package mediadevices

import (
	"github.com/pion/mediadevices/pkg/driver"
)

// RegisterDriverAdapter allows user space level of driver registration
func RegisterDriverAdapter(a driver.Adapter) error {
	return driver.GetManager().Register(a)
}
