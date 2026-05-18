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
// vpx_codec_frame_flags_t pktFrameFlags(vpx_codec_cx_pkt_t *pkt) {
//   return pkt->data.frame.flags;
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
//
// // NULL-safe wrapper around vpx_codec_error_detail. libvpx sets err_detail
// // inside its ERROR(...) macro (e.g. vp8_cx_iface.c:108-112) with the
// // specific reason for failures like VPX_CODEC_INVALID_PARAM, but the field
// // is NULL when no detail is available. Returning an empty C string in that
// // case keeps the Go side simple.
// const char *error_detail_safe(vpx_codec_ctx_t *ctx) {
//   if (ctx == NULL) return "";
//   const char *d = vpx_codec_error_detail(ctx);
//   if (d == NULL) return "";
//   return d;
// }
import "C"

import (
	"errors"
	"fmt"
	"image"
	"io"
	"math"
	"sync"
	"time"
	"unsafe"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

type encoder struct {
	codec           *C.vpx_codec_ctx_t
	raw             *C.vpx_image_t
	cfg             *C.vpx_codec_enc_cfg_t
	r               video.Reader
	frameIndex      int
	tStart          time.Time
	tLastFrame      time.Time
	frame           []byte
	deadline        int
	requireKeyFrame bool
	targetBitrate   int
	isKeyFrame      bool

	mu     sync.Mutex
	closed bool
}

const (
	kRateControlThreshold = 0.15
	kMinQuantizer         = 20
	kMaxQuantizer         = 63
)

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
		LagInFrames:                  uint(cfg.g_lag_in_frames),
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
	cfg.g_lag_in_frames = C.uint(params.LagInFrames)
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
		return nil, fmt.Errorf("vpx_codec_enc_init failed (%d): %s", ec, C.GoString(C.error_detail_safe(codec)))
	}
	t0 := time.Now()
	return &encoder{
		r:             video.ToI420(r),
		codec:         codec,
		raw:           rawNoBuffer,
		cfg:           cfg,
		tStart:        t0,
		tLastFrame:    t0,
		deadline:      int(params.Deadline / time.Microsecond),
		frame:         make([]byte, 1024),
		targetBitrate: params.BitRate,
	}, nil
}

func (e *encoder) Read() ([]byte, func(), error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return nil, func() {}, io.EOF
	}

	img, release, err := e.r.Read()
	if err != nil {
		return nil, func() {}, err
	}
	defer release()
	yuvImg := img.(*image.YCbCr)
	bounds := yuvImg.Bounds()
	height := C.int(bounds.Dy())
	width := C.int(bounds.Dx())

	e.raw.stride[0] = C.int(yuvImg.YStride)
	e.raw.stride[1] = C.int(yuvImg.CStride)
	e.raw.stride[2] = C.int(yuvImg.CStride)

	t := time.Now()

	if e.cfg.g_w != C.uint(width) || e.cfg.g_h != C.uint(height) {
		e.cfg.g_w, e.cfg.g_h = C.uint(width), C.uint(height)

		newCodec := C.newCtx()
		if ec := C.vpx_codec_enc_init_ver(
			newCodec, e.codec.iface, e.cfg, 0, C.VPX_ENCODER_ABI_VERSION,
		); ec != 0 {
			return nil, func() {}, fmt.Errorf("vpx_codec_enc_init failed (%d): %s", ec, C.GoString(C.error_detail_safe(newCodec)))
		}
		C.free(unsafe.Pointer(e.codec))
		e.codec = newCodec

		e.raw.w, e.raw.h = C.uint(width), C.uint(height)
		e.raw.r_w, e.raw.r_h = C.uint(width), C.uint(height)
		e.raw.d_w, e.raw.d_h = C.uint(width), C.uint(height)
	}

	if ec := C.vpx_codec_enc_config_set(e.codec, e.cfg); ec != 0 {
		return nil, func() {}, fmt.Errorf("vpx_codec_enc_config_set failed (%d): %s", ec, C.GoString(C.error_detail_safe(e.codec)))
	}

	duration := t.Sub(e.tLastFrame).Microseconds()
	// VPX doesn't allow 0 duration. If 0 is given, vpx_codec_encode will fail with VPX_CODEC_INVALID_PARAM.
	// 0 duration is possible because mediadevices first gets the frame meta data by reading from the source,
	// and consequently the codec will read the first frame from the buffer. This means the first frame won't
	// have a pause to the second frame, which means if the delay is <1 ms (vpx duration resolution), duration
	// is going to be 0.
	//
	// Also clamp implausibly large durations. tLastFrame is only updated after a successful encode, so when
	// an encoder sits idle for a long time (waiting for a remote subscriber, paused upstream, etc.) the next
	// frame's duration can exceed UINT32_MAX microseconds. libvpx 1.15.0+ rejects any duration strictly
	// greater than UINT32_MAX with VPX_CODEC_INVALID_PARAM — see libvpx/vpx/src/vpx_encoder.c:206 (added in
	// libvpx commit 7fb8ceccf, "Restrict ranges of duration,deadline to UINT32_MAX", 2024-03-14):
	//
	//   else if (duration > UINT32_MAX || deadline > UINT32_MAX)
	//       res = VPX_CODEC_INVALID_PARAM;
	//
	// UINT32_MAX microseconds is about 71m35s, so any encoder idle for longer than that fails its next
	// encode. Older libvpx (≤ 1.14.x) silently accepted huge durations, which is why this surfaced as a
	// new failure mode after a libvpx upgrade to 1.15+. Once the error fires, the failure path below also
	// does not update tLastFrame, so every subsequent encode sees an even larger duration and fails
	// identically — the encoder is permanently unusable until recreated.
	//
	// Substituting a sane synthetic frame interval keeps the encoder healthy across idle gaps; pts
	// continues monotonically from tStart so libvpx's internal pts ordering remains correct.
	const maxFrameDurationUs = 1_000_000   // 1s — well below libvpx 1.15+'s UINT32_MAX μs (~71 min) limit
	const fallbackFrameDurationUs = 33_333 // ~30fps frame interval, a sensible "this is a fresh frame"
	if duration <= 0 {
		duration = 1
	} else if duration > maxFrameDurationUs {
		duration = fallbackFrameDurationUs
	}

	targetVpxBitrate := C.uint(float32(e.targetBitrate / 1000)) // convert to kilobits / second
	if e.cfg.rc_target_bitrate != targetVpxBitrate && targetVpxBitrate >= 1 {
		e.cfg.rc_target_bitrate = targetVpxBitrate
		rc := C.vpx_codec_enc_config_set(e.codec, e.cfg)
		if rc != C.VPX_CODEC_OK {
			return nil, func() {}, fmt.Errorf("vpx_codec_enc_config_set failed (%d): %s", rc, C.GoString(C.error_detail_safe(e.codec)))
		}
	}

	var flags int
	if e.requireKeyFrame {
		flags = flags | C.VPX_EFLAG_FORCE_KF
	}
	if ec := C.encode_wrapper(
		e.codec, e.raw,
		C.long(t.Sub(e.tStart).Microseconds()), C.ulong(duration), C.long(flags), C.ulong(e.deadline),
		(*C.uchar)(&yuvImg.Y[0]), (*C.uchar)(&yuvImg.Cb[0]), (*C.uchar)(&yuvImg.Cr[0]),
	); ec != C.VPX_CODEC_OK {
		return nil, func() {}, fmt.Errorf("vpx_codec_encode failed (%d): %s", ec, C.GoString(C.error_detail_safe(e.codec)))
	}

	e.requireKeyFrame = false
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
			e.isKeyFrame = C.pktFrameFlags(pkt)&C.VPX_FRAME_IS_KEY == C.VPX_FRAME_IS_KEY
			encoded := C.GoBytes(unsafe.Pointer(C.pktBuf(pkt)), C.pktSz(pkt))
			e.frame = append(e.frame, encoded...)
		}
	}

	encoded := make([]byte, len(e.frame))
	copy(encoded, e.frame)
	return encoded, func() {}, err
}

func (e *encoder) ForceKeyFrame() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.requireKeyFrame = true
	return nil
}

func (e *encoder) SetBitRate(bitrate int) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.targetBitrate = bitrate
	return nil
}

func (e *encoder) DynamicQPControl(currentBitrate int, targetBitrate int) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	bitrateDiff := math.Abs(float64(currentBitrate - targetBitrate))
	if bitrateDiff <= float64(currentBitrate)*kRateControlThreshold {
		return nil
	}
	currentMax := e.cfg.rc_max_quantizer

	if targetBitrate < currentBitrate {
		e.cfg.rc_max_quantizer = min(currentMax+1, kMaxQuantizer)
	} else {
		e.cfg.rc_max_quantizer = max(currentMax-1, kMinQuantizer)
	}
	e.cfg.rc_min_quantizer = e.cfg.rc_max_quantizer
	return nil
}

func (e *encoder) Controller() codec.EncoderController {
	return e
}

func (e *encoder) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return nil
	}

	e.closed = true

	C.free(unsafe.Pointer(e.raw))
	defer C.free(unsafe.Pointer(e.codec))

	if C.vpx_codec_destroy(e.codec) != 0 {
		return errors.New("vpx_codec_destroy failed")
	}
	return nil
}
