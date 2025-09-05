package vpx

/*
#cgo pkg-config: vpx
#include <vpx/vpx_image.h>
*/
import "C"
import "unsafe"

type VpxImage struct {
	img *C.vpx_image_t
}

func NewImageFromPtr(ptr *C.vpx_image_t) *VpxImage {
	return &VpxImage{img: ptr}
}

func (i *VpxImage) Width() int {
	return int(i.img.d_w)
}

func (i *VpxImage) Height() int {
	return int(i.img.d_h)
}

func (i *VpxImage) YStride() int {
	return int(i.img.stride[0])
}

func (i *VpxImage) UStride() int {
	return int(i.img.stride[1])
}

func (i *VpxImage) VStride() int {
	return int(i.img.stride[2])
}

func (i *VpxImage) Plane(n int) unsafe.Pointer {
	return unsafe.Pointer(i.img.planes[n])
}
