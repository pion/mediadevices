package vaapi

// ParamVP8 stores VP8 encoding parameters.
type ParamVP8 struct {
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

// ParamVP9 represents VAEncSequenceParameterBufferVP9 and other parameter buffers.
type ParamVP9 struct {
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
