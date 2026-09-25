/*
Package camera provides a video camera driver.

Device Label Generation Rules

On Linux, the device label will be in the format of:
	pci-0000:00:00.0-usb-0:0:0.0-video-index0;video0
If /dev/v4l/by-path/* is not available (for example in a docker container without
bindings in /dev/v4l/by-path/), it will be:
	video0;video0
*/
package camera

import (
	"errors"
	"os"
	"strconv"
)

// LabelSeparator is used to separate labels for a driver that
// is found from multiple locations on a host.
const LabelSeparator = ";"

var errReadTimeout = errors.New("read timeout")

func getCameraReadTimeout() uint32 {
	// default to 5 seconds
	var readTimeoutSec uint32 = 5
	if val, ok := os.LookupEnv("PION_MEDIADEVICES_CAMERA_READ_TIMEOUT"); ok {
		if valInt, err := strconv.Atoi(val); err == nil {
			if valInt > 0 {
				readTimeoutSec = uint32(valInt)
			}
		}
	}
	return readTimeoutSec
}
