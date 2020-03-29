package screen

// #cgo pkg-config: x11 xext
// #include <stdint.h>
// #include <sys/shm.h>
// #include <X11/Xlib.h>
// #define XUTIL_DEFINE_FUNCTIONS
// #include <X11/Xutil.h>
// #include <X11/extensions/XShm.h>
//
// void copyBGR24(void *dst, char *src, size_t l) { // 64bit aligned copy
//   uint64_t *d = (uint64_t*)dst;
//   uint64_t *s = (uint64_t*)src;
//   l /= 8;
//   for (size_t i = 0; i < l; i ++) {
//     uint64_t v = *s;
//     // Reorder BGR to RGB
//     *d = 0xFF000000FF000000 |
//          ((v >> 16) & 0xFF00000000) | (v & 0xFF0000000000) | ((v & 0xFF00000000) << 16) |
//          ((v >> 16) & 0xFF) | (v & 0xFF00) | ((v & 0xFF) << 16);
//     d++;
//     s++;
//   }
// }
//
// void copyBGR16(void *dst, char *src, size_t l) { // 64bit aligned copy
//   uint64_t *d = (uint64_t*)dst;
//   uint32_t *s = (uint32_t*)src;
//   l /= 8;
//   for (size_t i = 0; i < l; i ++) {
//     uint64_t v = *s;
//     // Reorder BGR to RGB
//     *d = 0xFF000000FF000000 |
//          ((v & 0xF8000000) << 8)  | ((v & 0x7E00000) << 21) | ((v & 0x1F0000) << 35) |
//          ((v & 0xF800) >> 8) | ((v & 0x7E0) << 5) | ((v & 0x1F) << 19);
//     d++;
//     s++;
//   }
// }
//
// char *align64(char *ptr) { // return 64bit aligned pointer
//   if (((size_t)ptr & 0x07) == 0) {
//     return ptr;
//   }
//   // Clear lower 3bits to align the address to 8bytes.
//   return (char*)(((size_t)ptr & (~(size_t)0x07)) + 0x08);
// }
// size_t align64ForTest(size_t ptr) {
//   return (size_t)align64((char*)ptr);
// }
import "C"

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"unsafe"
)

const shmaddrInvalid = ^uintptr(0)

type display C.Display

type pixelFormat int

const (
	pixFmtBGR24 pixelFormat = iota
	pixFmtRGB24
	pixFmtBGR16
	pixFmtRGB16
)

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
	dp     *C.Display
	img    *C.XImage
	shm    C.XShmSegmentInfo
	b      []byte
	pixFmt pixelFormat
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
	return image.Rect(0, 0, int(s.img.width), int(s.img.height))
}

type colorFunc func() (r, g, b, a uint32)

func (c colorFunc) RGBA() (r, g, b, a uint32) {
	return c()
}

func (s *shmImage) At(x, y int) color.Color {
	switch s.pixFmt {
	case pixFmtBGR24:
		addr := (x + y*int(s.img.width)) * 4
		b := uint32(s.b[addr]) * 0x100
		g := uint32(s.b[addr+1]) * 0x100
		r := uint32(s.b[addr+2]) * 0x100
		return colorFunc(func() (_, _, _, _ uint32) {
			return r, g, b, 0xFFFF
		})
	case pixFmtBGR16:
		addr := (x + y*int(s.img.width)) * 2
		b1, b2 := s.b[addr], s.b[addr+1]
		b := uint32(b1>>3) * 0x100
		g := uint32((b1&0x7)<<3|(b2&0xE0)>>5) * 0x100
		r := uint32(b2&0x1F) * 0x100
		return colorFunc(func() (_, _, _, _ uint32) {
			return r, g, b, 0xFFFF
		})
	default:
		panic("unsupported pixel format")
	}
}

func (s *shmImage) RGBAAt(x, y int) color.RGBA {
	switch s.pixFmt {
	case pixFmtBGR24:
		addr := (x + y*int(s.img.width)) * 4
		b := s.b[addr]
		g := s.b[addr+1]
		r := s.b[addr+2]
		return color.RGBA{R: r, G: g, B: b, A: 0xFF}
	case pixFmtBGR16:
		addr := (x + y*int(s.img.width)) * 2
		b1, b2 := s.b[addr], s.b[addr+1]
		b := b1 >> 3
		g := (b1&0x7)<<3 | (b2&0xE0)>>5
		r := b2 & 0x1F
		return color.RGBA{R: r, G: g, B: b, A: 0xFF}
	default:
		panic("unsupported pixel format")
	}
}

func (s *shmImage) ToRGBA(dst *image.RGBA) *image.RGBA {
	dst.Rect = s.Bounds()
	dst.Stride = int(s.img.width) * 4
	l := int(4 * s.img.width * s.img.height)
	if len(dst.Pix) < l {
		if cap(dst.Pix) < l {
			dst.Pix = make([]uint8, l)
		}
		dst.Pix = dst.Pix[:l]
	}
	switch s.pixFmt {
	case pixFmtBGR24:
		C.copyBGR24(unsafe.Pointer(&dst.Pix[0]), s.img.data, C.ulong(len(dst.Pix)))
		return dst
	case pixFmtBGR16:
		C.copyBGR16(unsafe.Pointer(&dst.Pix[0]), s.img.data, C.ulong(len(dst.Pix)))
		return dst
	default:
		panic("unsupported pixel format")
	}
}

func newShmImage(dp *C.Display, screen int) (*shmImage, error) {
	cScreen := C.int(screen)
	w := int(C.XDisplayWidth(dp, cScreen))
	h := int(C.XDisplayHeight(dp, cScreen))
	v := C.XDefaultVisual(dp, cScreen)
	depth := int(C.XDefaultDepth(dp, cScreen))

	s := &shmImage{dp: dp}

	switch {
	case v.red_mask == 0xFF0000 && v.green_mask == 0xFF00 && v.blue_mask == 0xFF:
		s.pixFmt = pixFmtBGR24
	case v.red_mask == 0xF800 && v.green_mask == 0x7E0 && v.blue_mask == 0x1F:
		s.pixFmt = pixFmtBGR16
	default:
		fmt.Printf("x11capture: unsupported pixel format (R: %0x, G: %0x, B: %0x)\n",
			v.red_mask, v.green_mask, v.blue_mask)
		return nil, errors.New("unsupported pixel format")
	}

	s.shm.shmid = C.shmget(C.IPC_PRIVATE, C.ulong(w*h*4+8), C.IPC_CREAT|0600)
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
		dp, v, C.uint(depth), C.ZPixmap, C.align64(s.shm.shmaddr), &s.shm, C.uint(w), C.uint(h))
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

// cAlign64 is fot testing
func cAlign64(ptr uintptr) uintptr {
	return uintptr(C.align64ForTest(C.ulong(uintptr(ptr))))
}
