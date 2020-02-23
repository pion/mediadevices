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
	cam   *C.camera
	ch    chan []byte
	buf   []byte
	bufGo []byte
}

func init() {
	C.CoInitializeEx(nil, C.COINIT_MULTITHREADED)

	driver.GetManager().Register(&camera{}, driver.Info{
		Label:      "windows_camera_default",
		DeviceType: driver.Camera,
	})
}

func (c *camera) Open() error {
	c.ch = make(chan []byte)
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
	_ = cb

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
		C.freeCamera(c.cam)
		c.cam = nil
	}
	return nil
}

func (c *camera) VideoRecord(p prop.Media) (video.Reader, error) {
	c.buf = make([]byte, p.Width*p.Height*4)
	c.bufGo = make([]byte, p.Width*p.Height*4)
	cam := C.camera{
		width:  C.int(p.Width),
		height: C.int(p.Height),
		buf:    C.size_t(uintptr(unsafe.Pointer(&c.buf[0]))),
	}
	var errStr *C.char
	if C.openCamera(&cam, &errStr) != 0 {
		return nil, fmt.Errorf("failed to open device: %s", C.GoString(errStr))
	}
	c.cam = &cam

	callbacksMu.Lock()
	callbacks[uintptr(unsafe.Pointer(c.cam))] = c
	callbacksMu.Unlock()

	rgba := &image.RGBA{}
	r := video.ReaderFunc(func() (img image.Image, err error) {
		b, ok := <-c.ch
		if !ok {
			return nil, io.EOF
		}
		rgba.Pix = b
		rgba.Stride = p.Width * 4
		rgba.Rect = image.Rect(0, 0, p.Width, p.Height)
		return rgba, nil
	})
	return r, nil
}

func (c *camera) Properties() []prop.Media {
	properties := []prop.Media{}
	properties = append(properties, prop.Media{
		Video: prop.Video{
			// TODO: enum formats at beginning
			Width:  640,
			Height: 480,
			// TODO: DirectShow only supports RGB? Need investigation.
			FrameFormat: frame.FormatRGBA,
		},
	})
	return properties
}
