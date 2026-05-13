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
	LagInFrames                  uint
	ErrorResilient               ErrorResilientMode
	// EncodingThreads sets libvpx's g_threads. 0 leaves the libvpx default
	// (single-thread). Multi-threaded encoding can substantially improve
	// throughput on multi-core hosts; the effective ceiling depends on
	// resolution and codec (VP8 typically benefits up to ~4 threads, VP9 more).
	EncodingThreads uint
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
