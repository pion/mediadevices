// +build dragonfly freebsd linux netbsd openbsd solaris

package vaapi

// reference: https://github.com/intel/libva-utils/blob/master/encode/vp9enc.c

// #include <fcntl.h>
// #include <stdint.h>
// #include <stdio.h>
// #include <stdlib.h>
// #include <string.h>
//
// #include <va/va.h>
// #include <va/va_enc_vp9.h>
//
// #include "helper.h"
//
// void setShowFrameFlag9(VAEncPictureParameterBufferVP9 *p, uint32_t f) {
//   p->pic_flags.bits.show_frame = f;
// }
// void setForceKFFlag9(VAEncPictureParameterBufferVP9 *p, uint32_t f) {
//   p->ref_flags.bits.force_kf = f;
// }
// void setFrameTypeFlagVP9(VAEncPictureParameterBufferVP9 *p, uint32_t f) {
//   p->pic_flags.bits.frame_type = f;
// }
// void setRefFrameCtrlL0VP9(VAEncPictureParameterBufferVP9 *p, uint32_t f) {
//   p->ref_flags.bits.ref_frame_ctrl_l0 = f;
// }
// void setRefLastIndexVP9(VAEncPictureParameterBufferVP9 *p, uint32_t f) {
//   p->ref_flags.bits.ref_last_idx = f;
// }
// void setRefGFIndexVP9(VAEncPictureParameterBufferVP9 *p, uint32_t f) {
//   p->ref_flags.bits.ref_gf_idx = f;
// }
// void setRefARFIndexVP9(VAEncPictureParameterBufferVP9 *p, uint32_t f) {
//   p->ref_flags.bits.ref_arf_idx = f;
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
	surfaceVP9Ref0 = iota
	surfaceVP9Ref1
	surfaceVP9Ref2
	surfaceVP9Ref3
	surfaceVP9Ref4
	surfaceVP9Ref5
	surfaceVP9Ref6
	surfaceVP9Ref7
	surfaceVP9Input
	surfaceVP9Num
)

type encoderVP9 struct {
	r     video.Reader
	frame []byte

	fdDRI    C.int
	display  C.VADisplay
	confID   C.VAConfigID
	surfs    [surfaceVP9Num]C.VASurfaceID
	ctxID    C.VAContextID
	qMat     C.VAEncMiscParameterTypeVP9PerSegmantParam
	seqParam C.VAEncSequenceParameterBufferVP9
	picParam C.VAEncPictureParameterBufferVP9
	hrdParam hrdParam
	frParam  frParam
	rcParam  rcParam

	slotCurr int
	slotLast int
	slotGF   int
	slotARF  int

	frameCnt int
	prop     prop.Media
	params   ParamsVP9

	rate *framerateDetector

	mu     sync.Mutex
	closed bool
}

// newVP9Encoder creates new VP9 encoder
func newVP9Encoder(r video.Reader, p prop.Media, params ParamsVP9) (codec.ReadCloser, error) {
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

	// Parameters are from https://github.com/intel/libva-utils/blob/master/encode/vp9enc.c
	e := &encoderVP9{
		r:      video.ToI420(r),
		prop:   p,
		params: params,
		rate:   newFramerateDetector(uint32(p.FrameRate)),
		seqParam: C.VAEncSequenceParameterBufferVP9{
			max_frame_width:  8192,
			max_frame_height: 8192,
			bits_per_second:  C.uint(params.RateControl.bitsPerSecond),
			intra_period:     C.uint(params.KeyFrameInterval),
			kf_min_dist:      1,
			kf_max_dist:      C.uint(params.KeyFrameInterval),
		},
		picParam: C.VAEncPictureParameterBufferVP9{
			reference_frames: [8]C.VASurfaceID{
				C.VA_INVALID_ID,
				C.VA_INVALID_ID,
				C.VA_INVALID_ID,
				C.VA_INVALID_ID,
				C.VA_INVALID_ID,
				C.VA_INVALID_ID,
				C.VA_INVALID_ID,
				C.VA_INVALID_ID,
			},
			reconstructed_frame:    C.VA_INVALID_SURFACE,
			frame_width_src:        C.uint(p.Width),
			frame_height_src:       C.uint(p.Height),
			frame_width_dst:        C.uint(p.Width),
			frame_height_dst:       C.uint(p.Height),
			luma_ac_qindex:         60,
			luma_dc_qindex_delta:   1,
			chroma_ac_qindex_delta: 1,
			chroma_dc_qindex_delta: 1,
			filter_level:           10,
			ref_lf_delta: [4]C.int8_t{
				1, 1, 1, 1,
			},
			mode_lf_delta: [2]C.int8_t{
				1, 1,
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
	C.setShowFrameFlag9(&e.picParam, 1)

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
		C.VAProfileVP9Profile0,
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
		C.VAProfileVP9Profile0,
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
		C.VAProfileVP9Profile0,
		C.VAEntrypointEncSlice,
		&confAttrs[0], 2,
		&e.confID,
	); s != C.VA_STATUS_SUCCESS {
		return nil, fmt.Errorf("failed to create config: %s", C.GoString(C.vaErrorStr(s)))
	}

	surfAttr := C.VASurfaceAttrib{
		_type: C.VASurfaceAttribPixelFormat,
		flags: C.VA_SURFACE_ATTRIB_SETTABLE,
		value: C.genValInt(C.VA_FOURCC_NV12), // libva VP9 seems not supporting I420 surface
	}
	if s := C.vaCreateSurfaces(
		e.display,
		C.VA_RT_FORMAT_YUV420,
		C.uint(p.Width), C.uint(p.Height),
		&e.surfs[0], surfaceVP9Num,
		&surfAttr, 1,
	); s != C.VA_STATUS_SUCCESS {
		return nil, fmt.Errorf("failed to create surfaces: %s", C.GoString(C.vaErrorStr(s)))
	}

	if s := C.vaCreateContext(
		e.display,
		e.confID,
		C.int(p.Width), C.int(p.Height),
		C.VA_PROGRESSIVE,
		&e.surfs[0], surfaceVP9Num,
		&e.ctxID,
	); s != C.VA_STATUS_SUCCESS {
		return nil, fmt.Errorf("failed to create context: %s", C.GoString(C.vaErrorStr(s)))
	}

	return e, nil
}

func (e *encoderVP9) Read() ([]byte, func(), error) {
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
		C.setForceKFFlag9(&e.picParam, 1)
		C.setFrameTypeFlagVP9(&e.picParam, 0)
		e.picParam.refresh_frame_flags = 0
		for i := range e.picParam.reference_frames {
			e.picParam.reference_frames[i] = C.VA_INVALID_ID
		}
	} else {
		C.setForceKFFlag9(&e.picParam, 0)
		C.setFrameTypeFlagVP9(&e.picParam, 1)
		e.picParam.refresh_frame_flags = 1 << uint(e.slotCurr)
		C.setRefFrameCtrlL0VP9(&e.picParam, 0x7)
		C.setRefLastIndexVP9(&e.picParam, C.uint(e.slotLast))
		C.setRefGFIndexVP9(&e.picParam, C.uint(e.slotGF))
		C.setRefARFIndexVP9(&e.picParam, C.uint(e.slotARF))
	}
	e.picParam.reconstructed_frame = e.surfs[e.slotCurr]

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
		e.surfs[surfaceVP9Input],
	); s != C.VA_STATUS_SUCCESS {
		return nil, func() {}, fmt.Errorf("failed to begin picture: %s", C.GoString(C.vaErrorStr(s)))
	}

	// Upload image
	var vaImg C.VAImage
	var rawBuf unsafe.Pointer
	if s := C.vaDeriveImage(e.display, e.surfs[surfaceVP9Input], &vaImg); s != C.VA_STATUS_SUCCESS {
		return nil, func() {}, fmt.Errorf("failed to derive image: %s", C.GoString(C.vaErrorStr(s)))
	}
	if s := C.vaMapBuffer(e.display, vaImg.buf, &rawBuf); s != C.VA_STATUS_SUCCESS {
		return nil, func() {}, fmt.Errorf("failed to map buffer: %s", C.GoString(C.vaErrorStr(s)))
	}
	// TODO: use vaImg.pitches to support padding
	C.copyI420toNV12(
		rawBuf,
		(*C.uchar)(&yuvImg.Y[0]),
		(*C.uchar)(&yuvImg.Cb[0]),
		(*C.uchar)(&yuvImg.Cr[0]),
		C.uint(len(yuvImg.Y)),
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
	if s := C.vaSyncSurface(e.display, e.picParam.reconstructed_frame); s != C.VA_STATUS_SUCCESS {
		return nil, func() {}, fmt.Errorf("failed to sync surface: %s", C.GoString(C.vaErrorStr(s)))
	}
	var surfStat C.VASurfaceStatus
	if s := C.vaQuerySurfaceStatus(
		e.display, e.picParam.reconstructed_frame, &surfStat,
	); s != C.VA_STATUS_SUCCESS {
		return nil, func() {}, fmt.Errorf("failed to query surface status: %s", C.GoString(C.vaErrorStr(s)))
	}
	var seg *C.VACodedBufferSegment
	if s := C.vaMapBufferSeg(e.display, buffs[0], &seg); s != C.VA_STATUS_SUCCESS {
		return nil, func() {}, fmt.Errorf("failed to map buffer: %s", C.GoString(C.vaErrorStr(s)))
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

	// Update reference list
	e.picParam.reference_frames[e.slotCurr] = e.picParam.reconstructed_frame
	e.slotLast, e.slotGF, e.slotARF = e.slotCurr, e.slotLast, e.slotGF
	e.slotCurr++
	if e.slotCurr >= len(e.picParam.reference_frames) {
		e.slotCurr = 0
	}

	encoded := make([]byte, len(e.frame))
	copy(encoded, e.frame)
	return encoded, func() {}, err
}

func (e *encoderVP9) SetBitRate(b int) error {
	panic("SetBitRate is not implemented")
}

func (e *encoderVP9) ForceKeyFrame() error {
	panic("ForceKeyFrame is not implemented")
}

func (e *encoderVP9) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	C.vaDestroySurfaces(e.display, &e.surfs[0], C.int(len(e.surfs)))
	C.vaDestroyContext(e.display, e.ctxID)
	C.vaDestroyConfig(e.display, e.confID)
	closeDisplay(e.display, e.fdDRI)

	e.closed = true
	return nil
}
