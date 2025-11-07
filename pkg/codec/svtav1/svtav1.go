// Package svtav1 implements AV1 encoder.
// This package requires libSvtAv1Enc headers and libraries to be built.
package svtav1

// #cgo pkg-config: SvtAv1Enc
// #include "bridge.h"
import "C"

import (
	"image"
	"io"
	"sync"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

type encoder struct {
	engine *C.Encoder
	r      video.Reader
	mu     sync.Mutex
	closed bool

	outPool sync.Pool
}

func newEncoder(r video.Reader, p prop.Media, params Params) (codec.ReadCloser, error) {
	var enc *C.Encoder

	if p.FrameRate == 0 {
		p.FrameRate = 30
	}

	if err := errFromC(C.enc_new(&enc)); err != nil {
		return nil, err
	}
	enc.param.source_width = C.uint32_t(p.Width)
	enc.param.source_height = C.uint32_t(p.Height)
	enc.param.profile = C.MAIN_PROFILE
	enc.param.enc_mode = C.int8_t(params.Preset)
	enc.param.rate_control_mode = C.SVT_AV1_RC_MODE_CBR
	enc.param.pred_structure = C.SVT_AV1_PRED_LOW_DELAY_B
	enc.param.target_bit_rate = C.uint32_t(params.BitRate)
	enc.param.frame_rate_numerator = C.uint32_t(p.FrameRate * 1000)
	enc.param.frame_rate_denominator = 1000
	enc.param.intra_refresh_type = C.SVT_AV1_KF_REFRESH
	enc.param.intra_period_length = C.int32_t(params.KeyFrameInterval)
	enc.param.starting_buffer_level_ms = C.int64_t(params.StartingBufferLevel.Milliseconds())
	enc.param.optimal_buffer_level_ms = C.int64_t(params.OptimalBufferLevel.Milliseconds())
	enc.param.maximum_buffer_size_ms = C.int64_t(params.MaximumBufferSize.Milliseconds())

	if err := errFromC(C.enc_init(enc)); err != nil {
		_ = C.enc_free(enc)
		return nil, err
	}

	e := encoder{
		engine: enc,
		r:      video.ToI420(r),
		outPool: sync.Pool{
			New: func() any {
				return []byte(nil)
			},
		},
	}
	return &e, nil
}

func (e *encoder) Read() ([]byte, func(), error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return nil, func() {}, io.EOF
	}

	for {
		img, release, err := e.r.Read()
		if err != nil {
			return nil, func() {}, err
		}
		defer release()
		yuvImg := img.(*image.YCbCr)

		if err := errFromC(C.enc_send_frame(
			e.engine,
			(*C.uchar)(&yuvImg.Y[0]),
			(*C.uchar)(&yuvImg.Cb[0]),
			(*C.uchar)(&yuvImg.Cr[0]),
			C.int(yuvImg.YStride),
			C.int(yuvImg.CStride),
		)); err != nil {
			return nil, func() {}, err
		}

		var buf *C.EbBufferHeaderType
		if err := errFromC(C.enc_get_packet(e.engine, &buf)); err != nil {
			return nil, func() {}, err
		}
		if buf == nil {
			// Feed frames until receiving a packet
			continue
		}

		n := int(buf.n_filled_len)
		outBuf := e.outPool.Get().([]byte)
		if cap(outBuf) < n {
			outBuf = make([]byte, n)
		} else {
			outBuf = outBuf[:n]
		}

		C.memcpy_uint8((*C.uchar)(&outBuf[0]), buf.p_buffer, C.size_t(n))
		C.svt_av1_enc_release_out_buffer(&buf)

		return outBuf, func() {
			e.outPool.Put(outBuf)
		}, err
	}
}

func (e *encoder) ForceKeyFrame() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if err := errFromC(C.enc_force_keyframe(e.engine)); err != nil {
		return err
	}

	return nil
}

func (e *encoder) SetBitRate(bitrate int) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.engine.param.target_bit_rate = C.uint32_t(bitrate)

	if err := errFromC(C.enc_apply_param(e.engine)); err != nil {
		return err
	}

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

	if err := errFromC(C.enc_free(e.engine)); err != nil {
		return err
	}

	e.closed = true
	return nil
}
