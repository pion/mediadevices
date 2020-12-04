package vpx

import (
	"time"

	"github.com/pion/mediadevices/pkg/codec"
)

// Params stores libvpx specific encoding parameters.
// Value range is codec (VP8/VP9) specific.
type Params struct {
	codec.BaseParams
	Deadline                     time.Duration
	RateControlEndUsage          RateControlMode
	RateControlUndershootPercent uint
	RateControlOvershootPercent  uint
	RateControlMinQuantizer      uint
	RateControlMaxQuantizer      uint
	ErrorResilient               ErrorResilientMode
}

// RateControlMode represents rate control mode.
type RateControlMode int

// RateControlMode values.
const (
	RateControlVBR RateControlMode = iota
	RateControlCBR
	RateControlCQ
)

// ErrorResilientMode represents error resilient mode.
type ErrorResilientMode int

// ErrorResilientMode values.
const (
	ErrorResilientDefault    ErrorResilientMode = 0x01
	ErrorResilientPartitions ErrorResilientMode = 0x02
)
