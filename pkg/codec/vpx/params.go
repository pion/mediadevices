package vpx

// Params stores libvpx specific encoding parameters.
// Value range is codec (VP8/VP9) specific.
type Params struct {
	RateControlEndUsage          RateControlMode
	RateControlUndershootPercent uint
	RateControlOvershootPercent  uint
	RateControlMinQuantizer      uint
	RateControlMaxQuantizer      uint
}

// RateControlMode represents rate control mode.
type RateControlMode int

// RateControlMode values.
const (
	RateControlVBR RateControlMode = iota
	RateControlCBR
	RateControlCQ
)
