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
//     long t, ulong dt, long flags, ulong deadline,
//     u_char *y_ptr, u_char *cb_ptr, u_char *cr_ptr) {
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
	"time"
	"unsafe"

	"github.com/pion/mediadevices/pkg/codec"
	mio "github.com/pion/mediadevices/pkg/io"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"

	"github.com/pion/webrtc/v2"
)

type encoder struct {
	codec      *C.vpx_codec_ctx_t
	raw        *C.vpx_image_t
	r          video.Reader
	frameIndex int
	buff       []byte
	tStart     int
	tLastFrame int
}

func init() {
	codec.Register(webrtc.VP8, codec.VideoEncoderBuilder(NewVP8Encoder))
	codec.Register(webrtc.VP9, codec.VideoEncoderBuilder(NewVP9Encoder))
}

// NewVP8Encoder creates new VP8 encoder
func NewVP8Encoder(r video.Reader, p prop.Media) (io.ReadCloser, error) {
	return newEncoder(r, p, C.ifaceVP8())
}

// NewVP9Encoder creates new VP9 encoder
func NewVP9Encoder(r video.Reader, p prop.Media) (io.ReadCloser, error) {
	return newEncoder(r, p, C.ifaceVP9())
}

func newEncoder(r video.Reader, p prop.Media, codecIface *C.vpx_codec_iface_t) (io.ReadCloser, error) {
	if p.BitRate == 0 {
		p.BitRate = 100000
	}

	if p.KeyFrameInterval == 0 {
		p.KeyFrameInterval = 60
	}

	cfg := &C.vpx_codec_enc_cfg_t{}
	if ec := C.vpx_codec_enc_config_default(codecIface, cfg, 0); ec != 0 {
		return nil, fmt.Errorf("vpx_codec_enc_config_default failed (%d)", ec)
	}
	cfg.g_w = C.uint(p.Width)
	cfg.g_h = C.uint(p.Height)
	cfg.g_timebase.num = 1
	cfg.g_timebase.den = 1000
	cfg.rc_target_bitrate = C.uint(p.BitRate) / 1000
	cfg.kf_max_dist = C.uint(p.KeyFrameInterval)

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
		tStart:     t0,
		tLastFrame: t0,
	}, nil
}

func (e *encoder) Read(p []byte) (int, error) {
	if e.buff != nil {
		n, err := mio.Copy(p, e.buff)
		if err == nil {
			e.buff = nil
		}
		return n, err
	}

	img, err := e.r.Read()
	if err != nil {
		return 0, err
	}
	yuvImg := img.(*image.YCbCr)

	e.raw.stride[0] = C.int(yuvImg.YStride)
	e.raw.stride[1] = C.int(yuvImg.CStride)
	e.raw.stride[2] = C.int(yuvImg.CStride)

	t := time.Now().Nanosecond() / 1000000

	var flags int
	if ec := C.encode_wrapper(
		e.codec, e.raw,
		C.long(t-e.tStart), C.ulong(t-e.tLastFrame), C.long(flags), C.VPX_DL_REALTIME,
		(*C.uchar)(&yuvImg.Y[0]), (*C.uchar)(&yuvImg.Cb[0]), (*C.uchar)(&yuvImg.Cr[0]),
	); ec != C.VPX_CODEC_OK {
		return 0, fmt.Errorf("vpx_codec_encode failed (%d)", ec)
	}

	e.frameIndex++
	e.tLastFrame = t

	var frame []byte
	var iter C.vpx_codec_iter_t
	for {
		pkt := C.vpx_codec_get_cx_data(e.codec, &iter)
		if pkt == nil {
			break
		}
		if pkt.kind == C.VPX_CODEC_CX_FRAME_PKT {
			encoded := C.GoBytes(unsafe.Pointer(C.pktBuf(pkt)), C.pktSz(pkt))
			frame = append(frame, encoded...)
		}
	}
	n, err := mio.Copy(p, frame)
	if err != nil {
		e.buff = frame
	}
	return n, err
}

func (e *encoder) Close() error {
	C.free(unsafe.Pointer(e.raw))
	defer C.free(unsafe.Pointer(e.codec))

	if C.vpx_codec_destroy(e.codec) != 0 {
		return errors.New("vpx_codec_destroy failed")
	}
	return nil
}
