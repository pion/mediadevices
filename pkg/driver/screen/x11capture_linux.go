package screen

// #cgo pkg-config: x11 xext
// #include <sys/shm.h>
// #include <X11/Xlib.h>
// #define XUTIL_DEFINE_FUNCTIONS
// #include <X11/Xutil.h>
// #include <X11/extensions/XShm.h>
import "C"

import (
	"errors"
	"image"
	"image/color"
	"unsafe"
)

const shmaddrInvalid = ^uintptr(0)

type display C.Display

func openDisplay() (*display, error) {
	dp := C.XOpenDisplay(nil)
	if dp == nil {
		return nil, errors.New("failed to open display")
	}
	return (*display)(dp), nil
}

func (d *display) c() *C.Display {
	return (*C.Display)(d)
}

func (d *display) Close() {
	C.XCloseDisplay(d.c())
}

func (d *display) NumScreen() int {
	return int(C.XScreenCount(d.c()))
}

type shmImage struct {
	dp  *C.Display
	img *C.XImage
	shm C.XShmSegmentInfo
	b   []byte
}

func (s *shmImage) Free() {
	if s.img != nil {
		C.XShmDetach(s.dp, &s.shm)
		C.XDestroyImage(s.img)
	}
	if uintptr(unsafe.Pointer(s.shm.shmaddr)) != shmaddrInvalid {
		C.shmdt(unsafe.Pointer(s.shm.shmaddr))
	}
}

func (s *shmImage) ColorModel() color.Model {
	return color.RGBAModel
}

func (s *shmImage) Bounds() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{X: int(s.img.width), Y: int(s.img.height)},
	}
}

type colorFunc func() (r, g, b, a uint32)

func (c colorFunc) RGBA() (r, g, b, a uint32) {
	return c()
}

func (s *shmImage) At(x, y int) color.Color {
	addr := (x + y*int(s.img.width)) * 4
	b := uint32(s.b[addr]) * 0x100
	g := uint32(s.b[addr+1]) * 0x100
	r := uint32(s.b[addr+2]) * 0x100
	a := uint32(s.b[addr+3]) * 0x100
	return colorFunc(func() (_, _, _, _ uint32) {
		return r, g, b, a
	})
}

func newShmImage(dp *C.Display, screen int) (*shmImage, error) {
	cScreen := C.int(screen)
	w := int(C.XDisplayWidth(dp, cScreen))
	h := int(C.XDisplayHeight(dp, cScreen))
	v := C.XDefaultVisual(dp, cScreen)
	depth := int(C.XDefaultDepth(dp, cScreen))

	s := &shmImage{dp: dp}

	s.shm.shmid = C.shmget(C.IPC_PRIVATE, C.ulong(w*h*4), C.IPC_CREAT|0600)
	if s.shm.shmid == -1 {
		return nil, errors.New("failed to get shared memory")
	}
	s.shm.shmaddr = (*C.char)(C.shmat(s.shm.shmid, unsafe.Pointer(nil), 0))
	if uintptr(unsafe.Pointer(s.shm.shmaddr)) == shmaddrInvalid {
		s.shm.shmaddr = nil
		return nil, errors.New("failed to get shared memory address")
	}
	s.shm.readOnly = 0
	C.shmctl(s.shm.shmid, C.IPC_RMID, nil)

	s.img = C.XShmCreateImage(
		dp, v, C.uint(depth), C.ZPixmap, s.shm.shmaddr, &s.shm, C.uint(w), C.uint(h))
	if s.img == nil {
		s.Free()
		return nil, errors.New("failed to create XShm image")
	}
	C.XShmAttach(dp, &s.shm)
	C.XSync(dp, 0)

	return s, nil
}

type reader struct {
	dp  *C.Display
	img *shmImage
}

func newReader(screen int) (*reader, error) {
	dp := C.XOpenDisplay(nil)
	if dp == nil {
		return nil, errors.New("failed to open display")
	}
	if C.XShmQueryExtension(dp) == 0 {
		return nil, errors.New("no XShm support")
	}

	img, err := newShmImage(dp, screen)
	if err != nil {
		C.XCloseDisplay(dp)
		return nil, err
	}

	return &reader{
		dp:  dp,
		img: img,
	}, nil
}

func (r *reader) Size() (int, int) {
	return int(r.img.img.width), int(r.img.img.height)
}

func (r *reader) Read() *shmImage {
	C.XShmGetImage(r.dp, C.XDefaultRootWindow(r.dp), r.img.img, 0, 0, C.AllPlanes)
	r.img.b = C.GoBytes(
		unsafe.Pointer(r.img.img.data),
		C.int(r.img.img.width*r.img.img.height*4),
	)
	return r.img
}

func (r *reader) Close() {
	r.img.Free()
	C.XCloseDisplay(r.dp)
}
