package vaapi

import (
	"io"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/webrtc/v2"
)

// ParamVP8 stores VP8 encoding parameters.
type ParamVP8 struct {
	codec.BaseParams
	Sequence        SequenceParamVP8
	RateControlMode RateControlMode
	RateControl     RateControlParam
}

// SequenceParamVP8 represents VAEncSequenceParameterBufferVP8 and other parameter buffers.
type SequenceParamVP8 struct {
	ErrorResilient  bool
	ClampQindexHigh uint
	ClampQindexLow  uint
}

// NewVP8Param returns default parameters of VP8 codec.
func NewVP8Param() (ParamVP8, error) {
	return ParamVP8{
		Sequence: SequenceParamVP8{
			ClampQindexLow:  9,
			ClampQindexHigh: 127,
		},
		RateControlMode: RateControlVBR,
		RateControl: RateControlParam{
			BitsPerSecond:    400000,
			TargetPercentage: 80,
			WindowSize:       1500,
			InitialQP:        60,
			MinQP:            9,
			MaxQP:            127,
		},
	}, nil
}

// Name represents the codec name
func (p *ParamVP8) Name() string {
	return webrtc.VP8
}

// BuildVideoEncoder builds VP8 encoder with given params
func (p *ParamVP8) BuildVideoEncoder(r video.Reader, property prop.Media) (io.ReadCloser, error) {
	return newVP8Encoder(r, property, *p)
}

// ParamVP9 represents VAEncSequenceParameterBufferVP9 and other parameter buffers.
type ParamVP9 struct {
	codec.BaseParams
	RateControlMode RateControlMode
	RateControl     RateControlParam
}

// RateControlParam represents VAEncMiscParameterRateControl.
type RateControlParam struct {
	// BitsPerSecond is a maximum bit-rate.
	// This parameter overwrites prop.Codec.BitRate.
	BitsPerSecond uint
	// TargetPercentage is a target bit-rate relative to BitsPerSecond.
	TargetPercentage uint
	// WindowSize is a rate control window size in milliseconds.
	WindowSize uint
	InitialQP  uint
	MinQP      uint
	MaxQP      uint
}

// RateControlMode represents rate control mode.
// Note that supported mode depends on the codec and acceleration hardware.
type RateControlMode uint

// List of the RateControlMode.
const (
	RateControlCBR            RateControlMode = 0x00000002
	RateControlVBR            RateControlMode = 0x00000004
	RateControlVCM            RateControlMode = 0x00000008
	RateControlCQP            RateControlMode = 0x00000010
	RateControlVBRConstrained RateControlMode = 0x00000020
	RateControlICQ            RateControlMode = 0x00000040
	RateControlMB             RateControlMode = 0x00000080
	RateControlCFS            RateControlMode = 0x00000100
	RateControlParallel       RateControlMode = 0x00000200
	RateControlQVBR           RateControlMode = 0x00000400
	RateControlAVBR           RateControlMode = 0x00000800
)

// NewVP9Param returns default parameters of VP9 codec.
func NewVP9Param() (ParamVP9, error) {
	return ParamVP9{
		RateControlMode: RateControlVBR,
		RateControl: RateControlParam{
			BitsPerSecond:    400000,
			TargetPercentage: 80,
			WindowSize:       1500,
			InitialQP:        60,
			MinQP:            9,
			MaxQP:            127,
		},
	}, nil
}

// Name represents the codec name
func (p *ParamVP9) Name() string {
	return webrtc.VP9
}

// BuildVideoEncoder builds VP9 encoder with given params
func (p *ParamVP9) BuildVideoEncoder(r video.Reader, property prop.Media) (io.ReadCloser, error) {
	return newVP9Encoder(r, property, *p)
}
