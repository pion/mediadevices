// +build dragonfly freebsd linux netbsd openbsd solaris

// Package vaapi implements hardware accelerated codecs.
// This package requires libva headers and libraries to be built.
// libva requires supported graphic card and its driver.
package vaapi

import (
	"fmt"
	"unsafe"
)

// #cgo pkg-config: libva libva-drm
// #cgo CFLAGS: -DHAS_VAAPI
// #include <stdlib.h>
// #include <fcntl.h>
// #include <unistd.h>
// #include <va/va.h>
// #include <va/va_drm.h>
// #include "helper.h"
import "C"

const (
	bufferCoded = iota
	bufferSeqParam
	bufferPicParam
	bufferHRDParam
	bufferQMatrix
	bufferFRParam
	bufferRCParam
	bufferNum
)

type hrdParam struct {
	hdr  C.VAEncMiscParameterBuffer
	data C.VAEncMiscParameterHRD
}
type frParam struct {
	hdr  C.VAEncMiscParameterBuffer
	data C.VAEncMiscParameterFrameRate
}
type rcParam struct {
	hdr  C.VAEncMiscParameterBuffer
	data C.VAEncMiscParameterRateControl
}

func openDisplay(devPath string) (C.VADisplay, C.int, error) {
	// Try using dri
	path := C.CString(devPath)
	defer C.free(unsafe.Pointer(path))
	fdDRI := C.open2(path, C.O_RDWR)
	if fdDRI < 0 {
		C.close(fdDRI)
		return nil, 0, fmt.Errorf("failed to open %s", devPath)
	}
	return C.vaGetDisplayDRM(fdDRI), fdDRI, nil
}

func closeDisplay(d C.VADisplay, fd C.int) {
	C.vaTerminate(d)
	C.close(fd)
}
