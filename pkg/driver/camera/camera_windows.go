package camera

// #cgo LDFLAGS: -lstrmiids -lole32 -loleaut32 -lquartz
// #include <dshow.h>
// #include "camera_windows.hpp"
import "C"

import (
	"fmt"
	"image"
	"io"
	"sync"
	"unsafe"

	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/driver/availability"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

var (
	callbacks   = make(map[uintptr]*camera)
	callbacksMu sync.RWMutex
)

type camera struct {
	name  string
	cam   *C.camera
	ch    chan []byte
	buf   []byte
	bufGo []byte
}

func init() {
	Initialize()
}

// Initialize finds and registers camera devices. This is part of an experimental API.
func Initialize() {
	C.CoInitializeEx(nil, C.COINIT_MULTITHREADED)

	var list C.cameraList
	var errStr *C.char
	if C.listCamera(&list, &errStr) != 0 {
		// Failed to list camera
		fmt.Printf("Failed to list camera: %s\n", C.GoString(errStr))
		return
	}

	for i := 0; i < int(list.num); i++ {
		label := C.GoString(C.getName(&list, C.int(i)))
		info := driver.Info{
			Label:      label,
			DeviceType: driver.Camera,
		}
		if fn := C.getFriendlyName(&list, C.int(i)); fn != nil {
			info.Name = C.GoString(fn)
		}
		driver.GetManager().Register(&camera{name: label}, info)
	}

	C.freeCameraList(&list, &errStr)
}

// SetupObserver is a stub implementation for Windows.
func SetupObserver() error {
	return availability.ErrUnimplemented
}

// StartObserver is a stub implementation for Windows.
func StartObserver() error {
	return availability.ErrUnimplemented
}

// DestroyObserver is a stub implementation for Windows.
func DestroyObserver() error {
	return availability.ErrUnimplemented
}

func (c *camera) Open() error {
	// COM is per-thread on Windows. Go goroutines can run on any OS thread,
	// so ensure COM is initialized on the current one before making COM calls.
	C.CoInitializeEx(nil, C.COINIT_MULTITHREADED)

	c.ch = make(chan []byte)
	c.cam = &C.camera{
		name: C.CString(c.name),
	}

	var errStr *C.char
	if C.listResolution(c.cam, &errStr) != 0 {
		C.free(unsafe.Pointer(c.cam.name))
		return fmt.Errorf("failed to open device: %s", C.GoString(errStr))
	}

	return nil
}

//export imageCallback
func imageCallback(cam uintptr) {
	callbacksMu.RLock()
	cb, ok := callbacks[uintptr(unsafe.Pointer(cam))]
	callbacksMu.RUnlock()
	if !ok {
		return
	}

	copy(cb.bufGo, cb.buf)
	cb.ch <- cb.bufGo
}

func (c *camera) Close() error {
	callbacksMu.Lock()
	key := uintptr(unsafe.Pointer(c.cam))
	if _, ok := callbacks[key]; ok {
		delete(callbacks, key)
	}
	callbacksMu.Unlock()
	close(c.ch)

	if c.cam != nil {
		C.free(unsafe.Pointer(c.cam.name))
		C.freeCamera(c.cam)
		c.cam = nil
	}
	return nil
}

func (c *camera) VideoRecord(p prop.Media) (video.Reader, error) {
	C.CoInitializeEx(nil, C.COINIT_MULTITHREADED)

	nPix := p.Width * p.Height
	c.buf = make([]byte, nPix*2)
	c.bufGo = make([]byte, nPix*2)
	c.cam.width = C.int(p.Width)
	c.cam.height = C.int(p.Height)

	switch p.FrameFormat {
	case frame.FormatNV12:
		c.cam.fcc = fourccNV12
	default:
		c.cam.fcc = fourccYUY2
	}

	c.cam.buf = C.size_t(uintptr(unsafe.Pointer(&c.buf[0])))

	var errStr *C.char
	if C.openCamera(c.cam, &errStr) != 0 {
		return nil, fmt.Errorf("failed to open device: %s", C.GoString(errStr))
	}

	callbacksMu.Lock()
	callbacks[uintptr(unsafe.Pointer(c.cam))] = c
	callbacksMu.Unlock()

	img := &image.YCbCr{}

	r := video.ReaderFunc(func() (image.Image, func(), error) {
		b, ok := <-c.ch
		if !ok {
			return nil, func() {}, io.EOF
		}

		if p.FrameFormat == frame.FormatNV12 {
			// I420: Y plane (nPix) + U plane (nPix/4) + V plane (nPix/4)
			img.Y = b[:nPix]
			img.Cb = b[nPix : nPix+nPix/4]
			img.Cr = b[nPix+nPix/4 : nPix+nPix/2]
			img.YStride = p.Width
			img.CStride = p.Width / 2
			img.SubsampleRatio = image.YCbCrSubsampleRatio420
		} else { // YUY2
			// I422: Y plane (nPix) + Cb plane (nPix/2) + Cr plane (nPix/2)
			img.Y = b[:nPix]
			img.Cb = b[nPix : nPix+nPix/2]
			img.Cr = b[nPix+nPix/2 : nPix*2]
			img.YStride = p.Width
			img.CStride = p.Width / 2
			img.SubsampleRatio = image.YCbCrSubsampleRatio422
		}
		img.Rect = image.Rect(0, 0, p.Width, p.Height)
		return img, func() {}, nil
	})
	return r, nil
}

func (c *camera) Properties() []prop.Media {
	properties := []prop.Media{}
	for i := 0; i < int(c.cam.numProps); i++ {
		p := C.getProp(c.cam, C.int(i))
		var fmt frame.Format
		switch p.fcc {
		case fourccYUY2:
			fmt = frame.FormatYUY2
		case fourccNV12:
			fmt = frame.FormatNV12
		default:
			continue
		}
		properties = append(properties, prop.Media{
			Video: prop.Video{
				Width:       int(p.width),
				Height:      int(p.height),
				FrameFormat: fmt,
			},
		})
	}
	return properties
}

const (
	fourccYUY2 = 0x32595559
	fourccNV12 = 0x3231564E
)
