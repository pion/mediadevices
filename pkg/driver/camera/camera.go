/*
Package camera provides a video camera driver.

Device Label Generation Rules:
On Linux, the device label will be in the format of:
	/dev/video0;/dev/v4l/by-path/pci-0000:00:00.0-usb-0:0:0.0-video-index0
If /dev/v4l/by-path/* is not available (for example in a docker container without
bindings in /dev/v4l/by-path/), it will be:
	/dev/video0;/dev/video0
*/
package camera

// LabelSeparator is used to separate labels for a driver that
// is found from multiple locations on a host.
const LabelSeparator = ";"
