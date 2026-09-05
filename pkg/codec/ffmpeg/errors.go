package ffmpeg

import (
	"errors"
)

var (
	errFailedToCreateHwDevice    = errors.New("ffmpeg: failed to create device")
	errCodecNotFound             = errors.New("ffmpeg: codec not found")
	errFailedToCreateCodecCtx    = errors.New("ffmpeg: failed to allocate codec context")
	errFailedToCreateHwFramesCtx = errors.New("ffmpeg: failed to create hardware frames context")
	errFailedToInitHwFramesCtx   = errors.New("ffmpeg: failed to initialize hardware frames context")
	errFailedToOpenCodecCtx      = errors.New("ffmpeg: failed to open codec context")
	errFailedToAllocFrame        = errors.New("ffmpeg: failed to allocate frame")
	errFailedToAllocSwBuf        = errors.New("ffmpeg: failed to allocate software buffer")
	errFailedToAllocHwBuf        = errors.New("ffmpeg: failed to allocate hardware buffer")
	errFailedToAllocPacket       = errors.New("ffmpeg: failed to allocate packet")
)
