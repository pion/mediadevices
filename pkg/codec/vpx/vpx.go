// Package vpx implements VP8 and VP9 encoder.
// This package requires libvpx headers and libraries to be built.
package vpx

// #cgo pkg-config: vpx
// #include <stdlib.h>
// #include <vpx/vpx_encoder.h>
// #include <vpx/vpx_image.h>
// #include <vpx/vp8cx.h>
//
// // C function pointers
// vpx_codec_iface_t *ifaceVP8() {
//   return vpx_codec_vp8_cx();
// }
// vpx_codec_iface_t *ifaceVP9() {
//   return vpx_codec_vp9_cx();
// }
//
// // C union helpers
// void *pktBuf(vpx_codec_cx_pkt_t *pkt) {
//   return pkt->data.frame.buf;
// }
// int pktSz(vpx_codec_cx_pkt_t *pkt) {
//   return pkt->data.frame.sz;
// }
//
// // Alloc helpers
// vpx_codec_ctx_t *newCtx() {
//   return malloc(sizeof(vpx_codec_ctx_t));
// }
// vpx_image_t *newImage() {
//   return malloc(sizeof(vpx_image_t));
// }
//
// // Wrap encode function to keep Go memory safe
// vpx_codec_err_t encode_wrapper(
//     vpx_codec_ctx_t* codec, vpx_image_t* raw,
//     long t, unsigned long dt, long flags, unsigned long deadline,
//     unsigned char *y_ptr, unsigned char *cb_ptr, unsigned char *cr_ptr) {
//   raw->planes[0] = y_ptr;
//   raw->planes[1] = cb_ptr;
//   raw->planes[2] = cr_ptr;
//   vpx_codec_err_t ret = vpx_codec_encode(codec, raw, t, dt, flags, deadline);
//   raw->planes[0] = raw->planes[1] = raw->planes[2] = 0;
//   return ret;
// }
import "C"

import (
	"errors"
	"fmt"
	"image"
	"io"
	"sync"
	"time"
	"unsafe"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

type encoder struct {
	codec      *C.vpx_codec_ctx_t
	raw        *C.vpx_image_t
	cfg        *C.vpx_codec_enc_cfg_t
	r          video.Reader
	frameIndex int
	tStart     int
	tLastFrame int
	frame      []byte
	deadline   int

	mu     sync.Mutex
	closed bool
}

// VP8Params is codec specific paramaters
type VP8Params struct {
	Params
}

// NewVP8Params returns default VP8 codec specific parameters.
func NewVP8Params() (VP8Params, error) {
	p, err := newParams(C.ifaceVP8())
	if err != nil {
		return VP8Params{}, err
	}

	return VP8Params{
		Params: p,
	}, nil
}

// RTPCodec represents the codec metadata
func (p *VP8Params) RTPCodec() *codec.RTPCodec {
	return codec.NewRTPVP8Codec(90000)
}

// BuildVideoEncoder builds VP8 encoder with given params
func (p *VP8Params) BuildVideoEncoder(r video.Reader, property prop.Media) (codec.ReadCloser, error) {
	return newEncoder(r, property, p.Params, C.ifaceVP8())
}

// VP9Params is codec specific paramaters
type VP9Params struct {
	Params
}

// NewVP9Params returns default VP9 codec specific parameters.
func NewVP9Params() (VP9Params, error) {
	p, err := newParams(C.ifaceVP9())
	if err != nil {
		return VP9Params{}, err
	}

	return VP9Params{
		Params: p,
	}, nil
}

// RTPCodec represents the codec metadata
func (p *VP9Params) RTPCodec() *codec.RTPCodec {
	return codec.NewRTPVP9Codec(90000)
}

// BuildVideoEncoder builds VP9 encoder with given params
func (p *VP9Params) BuildVideoEncoder(r video.Reader, property prop.Media) (codec.ReadCloser, error) {
	return newEncoder(r, property, p.Params, C.ifaceVP9())
}

func newParams(codecIface *C.vpx_codec_iface_t) (Params, error) {
	cfg := &C.vpx_codec_enc_cfg_t{}
	if ec := C.vpx_codec_enc_config_default(codecIface, cfg, 0); ec != 0 {
		return Params{}, fmt.Errorf("vpx_codec_enc_config_default failed (%d)", ec)
	}
	return Params{
		Deadline:                     time.Microsecond * time.Duration(C.VPX_DL_REALTIME),
		RateControlEndUsage:          RateControlMode(cfg.rc_end_usage),
		RateControlUndershootPercent: uint(cfg.rc_undershoot_pct),
		RateControlOvershootPercent:  uint(cfg.rc_overshoot_pct),
		RateControlMinQuantizer:      uint(cfg.rc_min_quantizer),
		RateControlMaxQuantizer:      uint(cfg.rc_max_quantizer),
		ErrorResilient:               ErrorResilientMode(cfg.g_error_resilient),
	}, nil
}

func newEncoder(r video.Reader, p prop.Media, params Params, codecIface *C.vpx_codec_iface_t) (codec.ReadCloser, error) {
	if params.BitRate == 0 {
		params.BitRate = 100000
	}

	if params.KeyFrameInterval == 0 {
		params.KeyFrameInterval = 60
	}

	cfg := &C.vpx_codec_enc_cfg_t{}
	if ec := C.vpx_codec_enc_config_default(codecIface, cfg, 0); ec != 0 {
		return nil, fmt.Errorf("vpx_codec_enc_config_default failed (%d)", ec)
	}

	cfg.rc_end_usage = uint32(params.RateControlEndUsage)
	cfg.rc_undershoot_pct = C.uint(params.RateControlUndershootPercent)
	cfg.rc_overshoot_pct = C.uint(params.RateControlOvershootPercent)
	cfg.rc_min_quantizer = C.uint(params.RateControlMinQuantizer)
	cfg.rc_max_quantizer = C.uint(params.RateControlMaxQuantizer)

	cfg.g_error_resilient = C.uint32_t(params.ErrorResilient)

	cfg.g_w = C.uint(p.Width)
	cfg.g_h = C.uint(p.Height)
	cfg.g_timebase.num = 1
	cfg.g_timebase.den = 1000
	cfg.rc_target_bitrate = C.uint(params.BitRate) / 1000
	cfg.kf_max_dist = C.uint(params.KeyFrameInterval)

	cfg.rc_resize_allowed = 0
	cfg.g_pass = C.VPX_RC_ONE_PASS

	raw := &C.vpx_image_t{}
	if C.vpx_img_alloc(raw, C.VPX_IMG_FMT_I420, cfg.g_w, cfg.g_h, 1) == nil {
		return nil, errors.New("vpx_img_alloc failed")
	}
	rawNoBuffer := C.newImage()
	*rawNoBuffer = *raw // Copy only parameters
	C.vpx_img_free(raw) // Pointers will be overwritten by the raw buffer

	codec := C.newCtx()
	if ec := C.vpx_codec_enc_init_ver(
		codec, codecIface, cfg, 0, C.VPX_ENCODER_ABI_VERSION,
	); ec != 0 {
		return nil, fmt.Errorf("vpx_codec_enc_init failed (%d)", ec)
	}
	t0 := time.Now().Nanosecond() / 1000000
	return &encoder{
		r:          video.ToI420(r),
		codec:      codec,
		raw:        rawNoBuffer,
		cfg:        cfg,
		tStart:     t0,
		tLastFrame: t0,
		deadline:   int(params.Deadline / time.Microsecond),
		frame:      make([]byte, 1024),
	}, nil
}

func (e *encoder) Read() ([]byte, func(), error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return nil, func() {}, io.EOF
	}

	img, _, err := e.r.Read()
	if err != nil {
		return nil, func() {}, err
	}
	yuvImg := img.(*image.YCbCr)
	bounds := yuvImg.Bounds()
	height := C.int(bounds.Dy())
	width := C.int(bounds.Dx())

	e.raw.stride[0] = C.int(yuvImg.YStride)
	e.raw.stride[1] = C.int(yuvImg.CStride)
	e.raw.stride[2] = C.int(yuvImg.CStride)

	t := time.Now().Nanosecond() / 1000000

	if e.cfg.g_w != C.uint(width) || e.cfg.g_h != C.uint(height) {
		e.cfg.g_w, e.cfg.g_h = C.uint(width), C.uint(height)
		if ec := C.vpx_codec_enc_config_set(e.codec, e.cfg); ec != C.VPX_CODEC_OK {
			return nil, func() {}, fmt.Errorf("vpx_codec_enc_config_set failed (%d)", ec)
		}
		e.raw.w, e.raw.h = C.uint(width), C.uint(height)
		e.raw.r_w, e.raw.r_h = C.uint(width), C.uint(height)
		e.raw.d_w, e.raw.d_h = C.uint(width), C.uint(height)
	}

	duration := t - e.tLastFrame
	// VPX doesn't allow 0 duration. If 0 is given, vpx_codec_encode will fail with VPX_CODEC_INVALID_PARAM.
	// 0 duration is possible because mediadevices first gets the frame meta data by reading from the source,
	// and consequently the codec will read the first frame from the buffer. This means the first frame won't
	// have a pause to the second frame, which means if the delay is <1 ms (vpx duration resolution), duration
	// is going to be 0.
	if duration == 0 {
		duration = 1
	}
	var flags int
	if ec := C.encode_wrapper(
		e.codec, e.raw,
		C.long(t-e.tStart), C.ulong(duration), C.long(flags), C.ulong(e.deadline),
		(*C.uchar)(&yuvImg.Y[0]), (*C.uchar)(&yuvImg.Cb[0]), (*C.uchar)(&yuvImg.Cr[0]),
	); ec != C.VPX_CODEC_OK {
		return nil, func() {}, fmt.Errorf("vpx_codec_encode failed (%d)", ec)
	}

	e.frameIndex++
	e.tLastFrame = t

	e.frame = e.frame[:0]
	var iter C.vpx_codec_iter_t
	for {
		pkt := C.vpx_codec_get_cx_data(e.codec, &iter)
		if pkt == nil {
			break
		}
		if pkt.kind == C.VPX_CODEC_CX_FRAME_PKT {
			encoded := C.GoBytes(unsafe.Pointer(C.pktBuf(pkt)), C.pktSz(pkt))
			e.frame = append(e.frame, encoded...)
		}
	}

	encoded := make([]byte, len(e.frame))
	copy(encoded, e.frame)
	return encoded, func() {}, err
}

func (e *encoder) SetBitRate(b int) error {
	panic("SetBitRate is not implemented")
}

func (e *encoder) ForceKeyFrame() error {
	panic("ForceKeyFrame is not implemented")
}

func (e *encoder) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.closed = true

	C.free(unsafe.Pointer(e.raw))
	defer C.free(unsafe.Pointer(e.codec))

	if C.vpx_codec_destroy(e.codec) != 0 {
		return errors.New("vpx_codec_destroy failed")
	}
	return nil
}
