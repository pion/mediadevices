package vpx

/*
#cgo pkg-config: vpx
#include <stdlib.h>
#include <stdint.h>
#include <vpx/vpx_decoder.h>
#include <vpx/vpx_codec.h>
#include <vpx/vpx_image.h>
#include <vpx/vp8dx.h>

vpx_codec_iface_t *ifaceVP8Decoder() {
   return vpx_codec_vp8_dx();
}
vpx_codec_iface_t *ifaceVP9Decoder() {
   return vpx_codec_vp9_dx();
}

// Allocates a new decoder context
vpx_codec_ctx_t* newDecoderCtx() {
    return (vpx_codec_ctx_t*)malloc(sizeof(vpx_codec_ctx_t));
}

// Initializes the decoder
vpx_codec_err_t decoderInit(vpx_codec_ctx_t* ctx, vpx_codec_iface_t* iface) {
    return vpx_codec_dec_init_ver(ctx, iface, NULL, 0, VPX_DECODER_ABI_VERSION);
}

// Decodes an encoded frame
vpx_codec_err_t decodeFrame(vpx_codec_ctx_t* ctx, const uint8_t* data, unsigned int data_sz) {
    return vpx_codec_decode(ctx, data, data_sz, NULL, 0);
}

// Creates an iterator
vpx_codec_iter_t* newIter() {
    return (vpx_codec_iter_t*)malloc(sizeof(vpx_codec_iter_t));
}

// Returns the next decoded frame
vpx_image_t* getFrame(vpx_codec_ctx_t* ctx, vpx_codec_iter_t* iter) {
    return vpx_codec_get_frame(ctx, iter);
}

// Frees a decoded frane
void freeFrame(vpx_image_t* f) {
	vpx_img_free(f);
}

// Frees a decoder context
void freeDecoderCtx(vpx_codec_ctx_t* ctx) {
    vpx_codec_destroy(ctx);
    free(ctx);
}

*/
import "C"
import (
	"fmt"
	"image"
	"io"
	"sync"
	"time"
	"unsafe"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/prop"
)

type decoder struct {
	codec      *C.vpx_codec_ctx_t
	raw        *C.vpx_image_t
	cfg        *C.vpx_codec_dec_cfg_t
	iter       C.vpx_codec_iter_t
	frameIndex int
	tStart     time.Time
	tLastFrame time.Time
	reader     io.Reader
	buf        []byte

	mu     sync.Mutex
	closed bool
}

func BuildVideoDecoder(r io.Reader, property prop.Media) (codec.VideoDecoder, error) {
	return NewDecoder(r, property)
}

func NewDecoder(r io.Reader, p prop.Media) (codec.VideoDecoder, error) {
	cfg := &C.vpx_codec_dec_cfg_t{}
	cfg.threads = 1
	cfg.w = C.uint(p.Width)
	cfg.h = C.uint(p.Height)

	codec := C.newDecoderCtx()
	if C.decoderInit(codec, C.ifaceVP8Decoder()) != C.VPX_CODEC_OK {
		return nil, fmt.Errorf("vpx_codec_dec_init failed")
	}

	return &decoder{
		codec:  codec,
		cfg:    cfg,
		iter:   nil, // initialize to NULL to start iteration
		reader: r,
		buf:    make([]byte, 1024*1024),
	}, nil
}

func (d *decoder) Read() (image.Image, func(), error) {
	var input *C.vpx_image_t
	for {
		input = C.getFrame(d.codec, &d.iter)
		if input != nil {
			break
		}
		d.iter = nil
		// Read if there are no remained frames in the decoder
		n, err := d.reader.Read(d.buf)
		if err != nil {
			return nil, nil, err
		}
		status := C.decodeFrame(d.codec, (*C.uint8_t)(&d.buf[0]), C.uint(n))
		if status != C.VPX_CODEC_OK {
			return nil, nil, fmt.Errorf("decode failed: %v", status)
		}
	}
	w := int(input.d_w)
	h := int(input.d_h)
	yStride := int(input.stride[0])
	uStride := int(input.stride[1])
	vStride := int(input.stride[2])

	ySrc := unsafe.Slice((*byte)(unsafe.Pointer(input.planes[0])), yStride*h)
	uSrc := unsafe.Slice((*byte)(unsafe.Pointer(input.planes[1])), uStride*h/2)
	vSrc := unsafe.Slice((*byte)(unsafe.Pointer(input.planes[2])), vStride*h/2)

	dst := image.NewYCbCr(image.Rect(0, 0, w, h), image.YCbCrSubsampleRatio420)

	// copy luma
	for r := 0; r < h; r++ {
		copy(dst.Y[r*dst.YStride:r*dst.YStride+w], ySrc[r*yStride:r*yStride+w])
	}
	// copy chroma
	for r := 0; r < h/2; r++ {
		copy(dst.Cb[r*dst.CStride:r*dst.CStride+w/2], uSrc[r*uStride:r*uStride+w/2])
		copy(dst.Cr[r*dst.CStride:r*dst.CStride+w/2], vSrc[r*vStride:r*vStride+w/2])
	}
	C.freeFrame(input)
	return dst, func() {}, nil
}

func (d *decoder) Close() error {
	C.freeDecoderCtx(d.codec)
	d.closed = true
	return nil
}
