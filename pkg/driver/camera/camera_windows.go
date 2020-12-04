package camera

// #cgo LDFLAGS: -lstrmiids -lole32 -lquartz
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
	C.CoInitializeEx(nil, C.COINIT_MULTITHREADED)

	var list C.cameraList
	var errStr *C.char
	if C.listCamera(&list, &errStr) != 0 {
		// Failed to list camera
		fmt.Printf("Failed to list camera: %s\n", C.GoString(errStr))
		return
	}

	for i := 0; i < int(list.num); i++ {
		name := C.GoString(C.getName(&list, C.int(i)))
		driver.GetManager().Register(&camera{name: name}, driver.Info{
			Label:      name,
			DeviceType: driver.Camera,
		})
	}

	C.freeCameraList(&list, &errStr)
}

func (c *camera) Open() error {
	c.ch = make(chan []byte)
	c.cam = &C.camera{
		name: C.CString(c.name),
	}

	var errStr *C.char
	if C.listResolution(c.cam, &errStr) != 0 {
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
	nPix := p.Width * p.Height
	c.buf = make([]byte, nPix*2) // for YUY2
	c.bufGo = make([]byte, nPix*2)
	c.cam.width = C.int(p.Width)
	c.cam.height = C.int(p.Height)
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
		img.Y = b[:nPix]
		img.Cb = b[nPix : nPix+nPix/2]
		img.Cr = b[nPix+nPix/2 : nPix*2]
		img.YStride = p.Width
		img.CStride = p.Width / 2
		img.SubsampleRatio = image.YCbCrSubsampleRatio422
		img.Rect = image.Rect(0, 0, p.Width, p.Height)
		return img, func() {}, nil
	})
	return r, nil
}

func (c *camera) Properties() []prop.Media {
	properties := []prop.Media{}
	for i := 0; i < int(c.cam.numProps); i++ {
		p := C.getProp(c.cam, C.int(i))
		// TODO: support other FOURCC
		if p.fcc == fourccYUY2 {
			properties = append(properties, prop.Media{
				Video: prop.Video{
					Width:       int(p.width),
					Height:      int(p.height),
					FrameFormat: frame.FormatYUY2,
				},
			})
		}
	}
	return properties
}

const (
	fourccYUY2 = 0x32595559
)
