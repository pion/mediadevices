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

// Frees a decoder context
void freeDecoderCtx(vpx_codec_ctx_t* ctx) {
    vpx_codec_destroy(ctx);
    free(ctx);
}

*/
import "C"
import (
	"fmt"
	"sync"
	"time"

	"github.com/pion/mediadevices/pkg/prop"
)

type Decoder struct {
	codec      *C.vpx_codec_ctx_t
	raw        *C.vpx_image_t
	cfg        *C.vpx_codec_dec_cfg_t
	frameIndex int
	tStart     time.Time
	tLastFrame time.Time

	mu     sync.Mutex
	closed bool
}

func NewDecoder(p prop.Media) (Decoder, error) {
	cfg := &C.vpx_codec_dec_cfg_t{}
	cfg.threads = 1
	cfg.w = C.uint(p.Width)
	cfg.h = C.uint(p.Height)

	codec := C.newDecoderCtx()
	if C.decoderInit(codec, C.ifaceVP8Decoder()) != C.VPX_CODEC_OK {
		return Decoder{}, fmt.Errorf("vpx_codec_dec_init failed")
	}

	return Decoder{
		codec: codec,
		cfg:   cfg,
	}, nil
}

func (d *Decoder) Decode(data []byte) {
	if len(data) == 0 {
		return
	}
	status := C.decodeFrame(d.codec, (*C.uint8_t)(&data[0]), C.uint(len(data)))
	if status != C.VPX_CODEC_OK {
		fmt.Println("Decode failed", status)
		panic("Decode failed")
	}
}

func (d *Decoder) GetFrame() *VpxImage {
	var iter C.vpx_codec_iter_t = nil // initialize to NULL to start iteration
	img := C.getFrame(d.codec, &iter)
	if img == nil {
		return nil
	}
	return &VpxImage{img: img}
}

func (d *Decoder) FreeDecoderCtx() {
	C.freeDecoderCtx(d.codec)
}
