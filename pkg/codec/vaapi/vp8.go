// +build dragonfly freebsd linux netbsd openbsd solaris

package vaapi

// reference: https://github.com/intel/libva-utils/blob/master/encode/vp8enc.c

// #include <fcntl.h>
// #include <stdint.h>
// #include <stdio.h>
// #include <stdlib.h>
// #include <string.h>
// #include <unistd.h>
//
// #include <va/va.h>
// #include <va/va_drm.h>
// #include <va/va_enc_vp8.h>
//
// #include "helper.h"
//
// void setRefreshEntropyProbsFlagVP8(VAEncPictureParameterBufferVP8 *p, uint32_t f) {
//   p->pic_flags.bits.refresh_entropy_probs = f;
// }
// void setShowFrameFlagVP8(VAEncPictureParameterBufferVP8 *p, uint32_t f) {
//   p->pic_flags.bits.show_frame = f;
// }
// void setForceKFFlagVP8(VAEncPictureParameterBufferVP8 *p, uint32_t f) {
//   p->ref_flags.bits.force_kf = f;
// }
// void setFrameTypeFlagVP8(VAEncPictureParameterBufferVP8 *p, uint32_t f) {
//   p->pic_flags.bits.frame_type = f;
// }
// void setRefreshLastFlagVP8(VAEncPictureParameterBufferVP8 *p, uint32_t f) {
//   p->pic_flags.bits.refresh_last = f;
// }
// void setRefreshGoldenFrameFlagVP8(VAEncPictureParameterBufferVP8 *p, uint32_t f) {
//   p->pic_flags.bits.refresh_golden_frame = f;
// }
// void setRefreshAlternateFrameFlagVP8(VAEncPictureParameterBufferVP8 *p, uint32_t f) {
//   p->pic_flags.bits.refresh_alternate_frame = f;
// }
// void setCopyBufferToGoldenFlagVP8(VAEncPictureParameterBufferVP8 *p, uint32_t f) {
//   p->pic_flags.bits.copy_buffer_to_golden = f;
// }
// void setCopyBufferToAlternateFlagVP8(VAEncPictureParameterBufferVP8 *p, uint32_t f) {
//   p->pic_flags.bits.copy_buffer_to_alternate = f;
// }
// void setNoRefLastFrameFlagVP8(VAEncPictureParameterBufferVP8 *p, uint32_t f) {
//   p->ref_flags.bits.no_ref_last= f;
// }
// void setNoRefGoldenFrameFlagVP8(VAEncPictureParameterBufferVP8 *p, uint32_t f) {
//   p->ref_flags.bits.no_ref_gf= f;
// }
// void setNoRefAlternateFrameFlagVP8(VAEncPictureParameterBufferVP8 *p, uint32_t f) {
//   p->ref_flags.bits.no_ref_arf= f;
// }
import "C"

import (
	"errors"
	"fmt"
	"image"
	"io"
	"sync"
	"unsafe"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

const (
	surfaceVP8Ref0 = iota
	surfaceVP8Ref1
	surfaceVP8Ref2
	surfaceVP8Ref3
	surfaceVP8Input
	surfaceVP8Num
)

type encoderVP8 struct {
	r     video.Reader
	frame []byte

	fdDRI    C.int
	display  C.VADisplay
	confID   C.VAConfigID
	surfs    [surfaceVP8Num]C.VASurfaceID
	ctxID    C.VAContextID
	seqParam C.VAEncSequenceParameterBufferVP8
	picParam C.VAEncPictureParameterBufferVP8
	qMat     C.VAQMatrixBufferVP8
	hrdParam hrdParam
	frParam  frParam
	rcParam  rcParam

	frameCnt int
	prop     prop.Media
	params   ParamsVP8

	rate *framerateDetector

	mu     sync.Mutex
	closed bool
}

// newVP8Encoder creates new VP8 encoder
func newVP8Encoder(r video.Reader, p prop.Media, params ParamsVP8) (codec.ReadCloser, error) {
	if p.Width%16 != 0 || p.Width == 0 {
		return nil, errors.New("width must be 16*n")
	}
	if p.Height%16 != 0 || p.Height == 0 {
		return nil, errors.New("height must be 16*n")
	}
	if params.KeyFrameInterval == 0 {
		params.KeyFrameInterval = 30
	}
	if p.FrameRate == 0 {
		p.FrameRate = 30
	}

	params.RateControl.bitsPerSecond =
		uint(float32(params.BitRate) / (0.01 * float32(params.RateControl.TargetPercentage)))

	var error_resilient_value uint
	if params.Sequence.ErrorResilient {
		error_resilient_value = 1
	}

	// Parameters are from https://github.com/intel/libva-utils/blob/master/encode/vp8enc.c
	e := &encoderVP8{
		r:      video.ToI420(r),
		prop:   p,
		params: params,
		rate:   newFramerateDetector(uint32(p.FrameRate)),
		seqParam: C.VAEncSequenceParameterBufferVP8{
			frame_width:     C.uint(p.Width),
			frame_height:    C.uint(p.Height),
			error_resilient: C.uint(error_resilient_value),
			bits_per_second: C.uint(params.RateControl.bitsPerSecond),
			intra_period:    C.uint(params.KeyFrameInterval),
			kf_max_dist:     C.uint(params.KeyFrameInterval),
			reference_frames: [4]C.VASurfaceID{
				C.VA_INVALID_ID,
				C.VA_INVALID_ID,
				C.VA_INVALID_ID,
				C.VA_INVALID_ID,
			},
		},
		picParam: C.VAEncPictureParameterBufferVP8{
			ref_last_frame:      C.VA_INVALID_SURFACE,
			ref_gf_frame:        C.VA_INVALID_SURFACE,
			ref_arf_frame:       C.VA_INVALID_SURFACE,
			reconstructed_frame: C.VA_INVALID_SURFACE,
			clamp_qindex_low:    C.uint8_t(params.Sequence.ClampQindexLow),
			clamp_qindex_high:   C.uint8_t(params.Sequence.ClampQindexHigh),
			loop_filter_level: [4]C.int8_t{
				19, 19, 19, 19,
			},
		},
		qMat: C.VAQMatrixBufferVP8{
			quantization_index: [4]C.uint16_t{
				60, 60, 60, 60,
			},
			quantization_index_delta: [5]C.int16_t{
				0, 0, 0, 0, 0,
			},
		},
		hrdParam: hrdParam{
			hdr: C.VAEncMiscParameterBuffer{
				_type: C.VAEncMiscParameterTypeHRD,
			},
			data: C.VAEncMiscParameterHRD{
				initial_buffer_fullness: C.uint(params.RateControl.bitsPerSecond *
					params.RateControl.WindowSize / 2000),
				buffer_size: C.uint(params.RateControl.bitsPerSecond *
					params.RateControl.WindowSize / 1000),
			},
		},
		frParam: frParam{
			hdr: C.VAEncMiscParameterBuffer{
				_type: C.VAEncMiscParameterTypeFrameRate,
			},
			data: C.VAEncMiscParameterFrameRate{
				framerate: C.uint(p.FrameRate),
			},
		},
		rcParam: rcParam{
			hdr: C.VAEncMiscParameterBuffer{
				_type: C.VAEncMiscParameterTypeRateControl,
			},
			data: C.VAEncMiscParameterRateControl{
				window_size:       C.uint(params.RateControl.WindowSize),
				initial_qp:        C.uint(params.RateControl.InitialQP),
				min_qp:            C.uint(params.RateControl.MinQP),
				max_qp:            C.uint(params.RateControl.MaxQP),
				bits_per_second:   C.uint(params.RateControl.bitsPerSecond),
				target_percentage: C.uint(params.RateControl.TargetPercentage),
			},
		},
	}
	C.setShowFrameFlagVP8(&e.picParam, 1)

	// Try using dri
	var err error
	e.display, e.fdDRI, err = openDisplay("/dev/dri/card0")
	if err != nil {
		// TODO: try another graphic card and display via X11
		return nil, err
	}

	var vaMajor, vaMinor C.int
	if s := C.vaInitialize(e.display, &vaMajor, &vaMinor); s != C.VA_STATUS_SUCCESS {
		return nil, fmt.Errorf("failed to init libva: %s", C.GoString(C.vaErrorStr(s)))
	}

	var numEntrypoints C.int
	entrypoints := make([]C.VAEntrypoint, int(C.vaMaxNumEntrypoints(e.display)))

	if s := C.vaQueryConfigEntrypoints(
		e.display,
		C.VAProfileVP8Version0_3,
		&entrypoints[0],
		&numEntrypoints,
	); s != C.VA_STATUS_SUCCESS {
		return nil, fmt.Errorf("failed to query libva entrypoints: %s", C.GoString(C.vaErrorStr(s)))
	}

	var epFound bool
	for i := 0; i < int(numEntrypoints); i++ {
		if entrypoints[i] == C.VAEntrypointEncSlice {
			epFound = true
		}
	}
	if !epFound {
		return nil, errors.New("libva entrypoint not found")
	}

	confAttrs := []C.VAConfigAttrib{
		{_type: C.VAConfigAttribRTFormat},
		{_type: C.VAConfigAttribRateControl},
	}
	if s := C.vaGetConfigAttributes(
		e.display,
		C.VAProfileVP8Version0_3,
		C.VAEntrypointEncSlice,
		&confAttrs[0], 2,
	); s != C.VA_STATUS_SUCCESS {
		return nil, fmt.Errorf("failed to get config attrs: %s", C.GoString(C.vaErrorStr(s)))
	}
	if (confAttrs[0].value & C.VA_RT_FORMAT_YUV420) == 0 {
		return nil, errors.New("the hardware encoder doesn't support YUV420")
	}
	if (confAttrs[1].value & C.uint(params.RateControlMode)) == 0 {
		return nil, errors.New("the hardware encoder doesn't support specified rate control mode")
	}
	confAttrs[0].value = C.VA_RT_FORMAT_YUV420
	confAttrs[1].value = C.uint(params.RateControlMode)

	if s := C.vaCreateConfig(
		e.display,
		C.VAProfileVP8Version0_3,
		C.VAEntrypointEncSlice,
		&confAttrs[0], 2,
		&e.confID,
	); s != C.VA_STATUS_SUCCESS {
		return nil, fmt.Errorf("failed to create config: %s", C.GoString(C.vaErrorStr(s)))
	}

	surfAttr := C.VASurfaceAttrib{
		_type: C.VASurfaceAttribPixelFormat,
		flags: C.VA_SURFACE_ATTRIB_SETTABLE,
		value: C.genValInt(C.VA_FOURCC_I420),
	}
	if s := C.vaCreateSurfaces(
		e.display,
		C.VA_RT_FORMAT_YUV420,
		C.uint(p.Width), C.uint(p.Height),
		&e.surfs[0], surfaceVP8Num,
		&surfAttr, 1,
	); s != C.VA_STATUS_SUCCESS {
		return nil, fmt.Errorf("failed to create surfaces: %s", C.GoString(C.vaErrorStr(s)))
	}

	if s := C.vaCreateContext(
		e.display,
		e.confID,
		C.int(p.Width), C.int(p.Height),
		C.VA_PROGRESSIVE,
		&e.surfs[0], surfaceVP8Num,
		&e.ctxID,
	); s != C.VA_STATUS_SUCCESS {
		return nil, fmt.Errorf("failed to create context: %s", C.GoString(C.vaErrorStr(s)))
	}

	return e, nil
}

func (e *encoderVP8) Read() ([]byte, func(), error) {
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

	kf := e.frameCnt%e.params.KeyFrameInterval == 0
	e.frameCnt++

	e.frParam.data.framerate = C.uint(e.rate.Calc())

	if kf {
		// Key frame
		C.setForceKFFlagVP8(&e.picParam, 1)
		C.setFrameTypeFlagVP8(&e.picParam, 0)
		C.setRefreshLastFlagVP8(&e.picParam, 0)
		C.setRefreshGoldenFrameFlagVP8(&e.picParam, 0)
		C.setCopyBufferToGoldenFlagVP8(&e.picParam, 0)
		C.setRefreshAlternateFrameFlagVP8(&e.picParam, 0)
		C.setCopyBufferToAlternateFlagVP8(&e.picParam, 0)

		if e.params.Sequence.ErrorResilient {
			C.setRefreshEntropyProbsFlagVP8(&e.picParam, 1)
		} else {
			C.setRefreshEntropyProbsFlagVP8(&e.picParam, 0)
		}

		if e.picParam.reconstructed_frame == C.VA_INVALID_SURFACE {
			e.picParam.reconstructed_frame = e.surfs[surfaceVP8Ref0]
			e.picParam.ref_last_frame = e.surfs[surfaceVP8Ref0]
			e.picParam.ref_gf_frame = e.surfs[surfaceVP8Ref0]
			e.picParam.ref_arf_frame = e.surfs[surfaceVP8Ref0]
		}
	} else {
		// Select available surface
		e.picParam.reconstructed_frame = C.VA_INVALID_SURFACE
		for i := surfaceVP8Ref0; i <= surfaceVP8Ref3; i++ {
			s := e.surfs[i]
			if s != e.picParam.ref_last_frame && s != e.picParam.ref_gf_frame && s != e.picParam.ref_arf_frame {
				e.picParam.reconstructed_frame = s
				break
			}
		}
		if e.picParam.reconstructed_frame == C.VA_INVALID_SURFACE {
			return nil, func() {}, errors.New("no available surface")
		}

		C.setForceKFFlagVP8(&e.picParam, 0)
		C.setFrameTypeFlagVP8(&e.picParam, 1)
		C.setNoRefLastFrameFlagVP8(&e.picParam, 0)
		C.setNoRefGoldenFrameFlagVP8(&e.picParam, 0)
		C.setNoRefAlternateFrameFlagVP8(&e.picParam, 1)
		C.setRefreshEntropyProbsFlagVP8(&e.picParam, 0)
	}

	// Prepare buffers
	buffs := make([]C.VABufferID, 0, bufferNum)
	type buffParam struct {
		typ  C.VABufferType
		n    uint
		num  uint
		src  unsafe.Pointer
		hook func()
	}
	buffParams := []buffParam{
		{
			typ: C.VAEncCodedBufferType,
			n:   uint(e.prop.Width * e.prop.Height), num: 1, src: nil,
		},
		{
			typ: C.VAEncSequenceParameterBufferType,
			n:   uint(unsafe.Sizeof(e.seqParam)), num: 1, src: unsafe.Pointer(&e.seqParam),
		},
		{
			typ: C.VAEncPictureParameterBufferType,
			n:   uint(unsafe.Sizeof(e.picParam)), num: 1, src: unsafe.Pointer(&e.picParam),
			hook: func() {
				e.picParam.coded_buf = buffs[0]
			},
		},
		{
			typ: C.VAEncMiscParameterBufferType,
			n:   uint(unsafe.Sizeof(e.hrdParam)), num: 1, src: unsafe.Pointer(&e.hrdParam),
		},
		{
			typ: C.VAQMatrixBufferType,
			n:   uint(unsafe.Sizeof(e.qMat)), num: 1, src: unsafe.Pointer(&e.qMat),
		},
	}
	if kf {
		buffParams = append(buffParams,
			buffParam{
				typ: C.VAEncMiscParameterBufferType,
				n:   uint(unsafe.Sizeof(e.frParam)), num: 1, src: unsafe.Pointer(&e.frParam),
			},
			buffParam{
				typ: C.VAEncMiscParameterBufferType,
				n:   uint(unsafe.Sizeof(e.rcParam)), num: 1, src: unsafe.Pointer(&e.rcParam),
			},
		)
	}
	for _, p := range buffParams {
		if p.hook != nil {
			p.hook()
		}
		var id C.VABufferID
		if s := C.vaCreateBufferPtr(
			e.display, e.ctxID,
			p.typ, C.uint(p.n), C.uint(p.num),
			C.size_t(uintptr(p.src)),
			&id,
		); s != C.VA_STATUS_SUCCESS {
			return nil, func() {}, fmt.Errorf("failed to create buffer: %s", C.GoString(C.vaErrorStr(s)))
		}
		buffs = append(buffs, id)
	}

	// Render picture
	if s := C.vaBeginPicture(
		e.display, e.ctxID,
		e.surfs[surfaceVP8Input],
	); s != C.VA_STATUS_SUCCESS {
		return nil, func() {}, fmt.Errorf("failed to begin picture: %s", C.GoString(C.vaErrorStr(s)))
	}

	// Upload image
	var vaImg C.VAImage
	var rawBuf unsafe.Pointer
	if s := C.vaDeriveImage(e.display, e.surfs[surfaceVP8Input], &vaImg); s != C.VA_STATUS_SUCCESS {
		return nil, func() {}, fmt.Errorf("failed to derive image: %s", C.GoString(C.vaErrorStr(s)))
	}
	if s := C.vaMapBuffer(e.display, vaImg.buf, &rawBuf); s != C.VA_STATUS_SUCCESS {
		return nil, func() {}, fmt.Errorf("failed to map buffer: %s", C.GoString(C.vaErrorStr(s)))
	}
	// TODO: use vaImg.pitches to support padding
	C.memcpy(
		unsafe.Pointer(uintptr(rawBuf)+uintptr(vaImg.offsets[0])),
		unsafe.Pointer(&yuvImg.Y[0]), C.size_t(len(yuvImg.Y)),
	)
	C.memcpy(
		unsafe.Pointer(uintptr(rawBuf)+uintptr(vaImg.offsets[1])),
		unsafe.Pointer(&yuvImg.Cb[0]), C.size_t(len(yuvImg.Cb)),
	)
	C.memcpy(
		unsafe.Pointer(uintptr(rawBuf)+uintptr(vaImg.offsets[2])),
		unsafe.Pointer(&yuvImg.Cr[0]), C.size_t(len(yuvImg.Cr)),
	)
	if s := C.vaUnmapBuffer(e.display, vaImg.buf); s != C.VA_STATUS_SUCCESS {
		return nil, func() {}, fmt.Errorf("failed to unmap buffer: %s", C.GoString(C.vaErrorStr(s)))
	}
	if s := C.vaDestroyImage(e.display, vaImg.image_id); s != C.VA_STATUS_SUCCESS {
		return nil, func() {}, fmt.Errorf("failed to destroy image: %s", C.GoString(C.vaErrorStr(s)))
	}

	if s := C.vaRenderPicture(
		e.display, e.ctxID,
		&buffs[1], // 0 is for ouput
		C.int(len(buffs)-1),
	); s != C.VA_STATUS_SUCCESS {
		return nil, func() {}, fmt.Errorf("failed to render picture: %s", C.GoString(C.vaErrorStr(s)))
	}
	if s := C.vaEndPicture(
		e.display, e.ctxID,
	); s != C.VA_STATUS_SUCCESS {
		return nil, func() {}, fmt.Errorf("failed to end picture: %s", C.GoString(C.vaErrorStr(s)))
	}

	// Load encoded data
	for retry := 3; retry >= 0; retry-- {
		if s := C.vaSyncSurface(e.display, e.picParam.reconstructed_frame); s != C.VA_STATUS_SUCCESS {
			return nil, func() {}, fmt.Errorf("failed to sync surface: %s", C.GoString(C.vaErrorStr(s)))
		}
		var surfStat C.VASurfaceStatus
		if s := C.vaQuerySurfaceStatus(
			e.display, e.picParam.reconstructed_frame, &surfStat,
		); s != C.VA_STATUS_SUCCESS {
			return nil, func() {}, fmt.Errorf("failed to query surface status: %s", C.GoString(C.vaErrorStr(s)))
		}
		if surfStat == C.VASurfaceReady {
			break
		}
		if retry == 0 {
			return nil, func() {}, fmt.Errorf("failed to sync surface: %d", surfStat)
		}
	}
	var seg *C.VACodedBufferSegment
	if s := C.vaMapBufferSeg(e.display, buffs[0], &seg); s != C.VA_STATUS_SUCCESS {
		return nil, func() {}, fmt.Errorf("failed to map buffer: %s", C.GoString(C.vaErrorStr(s)))
	}
	if seg.status&C.VA_CODED_BUF_STATUS_SLICE_OVERFLOW_MASK != 0 {
		return nil, func() {}, errors.New("buffer size too small")
	}

	if cap(e.frame) < int(seg.size) {
		e.frame = make([]byte, int(seg.size))
	}
	e.frame = e.frame[:int(seg.size)]
	C.memcpy(
		unsafe.Pointer(&e.frame[0]),
		unsafe.Pointer(seg.buf), C.size_t(seg.size),
	)

	if s := C.vaUnmapBuffer(e.display, buffs[0]); s != C.VA_STATUS_SUCCESS {
		return nil, func() {}, fmt.Errorf("failed to unmap buffer: %s", C.GoString(C.vaErrorStr(s)))
	}

	// Destroy buffers
	for _, b := range buffs {
		if s := C.vaDestroyBuffer(e.display, b); s != C.VA_STATUS_SUCCESS {
			return nil, func() {}, fmt.Errorf("failed to destroy buffer: %s", C.GoString(C.vaErrorStr(s)))
		}
	}

	// vp8enc_update_reference_list
	if kf {
		e.picParam.ref_gf_frame = e.picParam.reconstructed_frame
		e.picParam.ref_arf_frame = e.picParam.reconstructed_frame
		C.setRefreshGoldenFrameFlagVP8(&e.picParam, 1)
		C.setCopyBufferToGoldenFlagVP8(&e.picParam, 0)
		C.setRefreshAlternateFrameFlagVP8(&e.picParam, 1)
		C.setCopyBufferToAlternateFlagVP8(&e.picParam, 0)
	} else {
		C.setRefreshGoldenFrameFlagVP8(&e.picParam, 0)
		C.setCopyBufferToGoldenFlagVP8(&e.picParam, 0)
		C.setRefreshAlternateFrameFlagVP8(&e.picParam, 0)
		C.setCopyBufferToAlternateFlagVP8(&e.picParam, 0)
		// TODO: proper golden frame update
		//       (ffmpeg vaapi_encode_vp8 also don't use golden frame)
	}
	e.picParam.ref_last_frame = e.picParam.reconstructed_frame
	C.setRefreshLastFlagVP8(&e.picParam, 1)

	encoded := make([]byte, len(e.frame))
	copy(encoded, e.frame)
	return encoded, func() {}, err
}

func (e *encoderVP8) SetBitRate(b int) error {
	panic("SetBitRate is not implemented")
}

func (e *encoderVP8) ForceKeyFrame() error {
	panic("ForceKeyFrame is not implemented")
}

func (e *encoderVP8) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	C.vaDestroySurfaces(e.display, &e.surfs[0], C.int(len(e.surfs)))
	C.vaDestroyContext(e.display, e.ctxID)
	C.vaDestroyConfig(e.display, e.confID)
	C.vaTerminate(e.display)
	C.close(e.fdDRI)

	e.closed = true
	return nil
}
