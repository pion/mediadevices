package vaapi

import (
	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

// ParamsVP8 stores VP8 encoding parameters.
type ParamsVP8 struct {
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

// NewVP8Params returns default parameters of VP8 codec.
func NewVP8Params() (ParamsVP8, error) {
	return ParamsVP8{
		BaseParams: codec.BaseParams{
			BitRate:          320000,
			KeyFrameInterval: 30,
		},
		Sequence: SequenceParamVP8{
			ClampQindexLow:  9,
			ClampQindexHigh: 127,
		},
		RateControlMode: RateControlVBR,
		RateControl: RateControlParam{
			TargetPercentage: 80,
			WindowSize:       1500,
			InitialQP:        60,
			MinQP:            9,
			MaxQP:            127,
		},
	}, nil
}

// RTPCodec represents the codec metadata
func (p *ParamsVP8) RTPCodec() *codec.RTPCodec {
	return codec.NewRTPVP8Codec(90000)
}

// BuildVideoEncoder builds VP8 encoder with given params
func (p *ParamsVP8) BuildVideoEncoder(r video.Reader, property prop.Media) (codec.ReadCloser, error) {
	return newVP8Encoder(r, property, *p)
}

// ParamsVP9 represents VAEncSequenceParameterBufferVP9 and other parameter buffers.
type ParamsVP9 struct {
	codec.BaseParams
	RateControlMode RateControlMode
	RateControl     RateControlParam
}

// RateControlParam represents VAEncMiscParameterRateControl.
type RateControlParam struct {
	// bitsPerSecond is a maximum bit-rate.
	// This parameter is calculated from BaseParams.BitRate.
	bitsPerSecond uint
	// TargetPercentage is a target bit-rate relative to the maximum bit-rate.
	// BaseParams.BitRate / (TargetPercentage * 0.01) will be the maximum bit-rate.
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

// NewVP9Params returns default parameters of VP9 codec.
func NewVP9Params() (ParamsVP9, error) {
	return ParamsVP9{
		BaseParams: codec.BaseParams{
			BitRate:          320000,
			KeyFrameInterval: 30,
		},
		RateControlMode: RateControlVBR,
		RateControl: RateControlParam{
			TargetPercentage: 80,
			WindowSize:       1500,
			InitialQP:        60,
			MinQP:            9,
			MaxQP:            127,
		},
	}, nil
}

// RTPCodec represents the codec metadata
func (p *ParamsVP9) RTPCodec() *codec.RTPCodec {
	return codec.NewRTPVP9Codec(90000)
}

// BuildVideoEncoder builds VP9 encoder with given params
func (p *ParamsVP9) BuildVideoEncoder(r video.Reader, property prop.Media) (codec.ReadCloser, error) {
	return newVP9Encoder(r, property, *p)
}
