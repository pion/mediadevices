// Package ffmpeg brings libavcodec's encoding capabilities to mediadevices.
// This package requires ffmpeg headers and libraries to be built.
// For more information, see https://github.com/asticode/go-astiav?tab=readme-ov-file#install-ffmpeg-from-source.
//
// Currently, only nvenc, x264, vaapi are implemented, but extending this to other ffmpeg supported codecs should
// be simple.
package ffmpeg

import (
	"errors"
	"io"
	"sync"

	"github.com/asticode/go-astiav"
	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

type hardwareEncoder struct {
	codecCtx       *astiav.CodecContext
	hwFramesCtx    *astiav.HardwareFramesContext
	frame          *astiav.Frame
	hwFrame        *astiav.Frame
	packet         *astiav.Packet
	width          int
	height         int
	r              video.Reader
	nextIsKeyFrame bool

	mu     sync.Mutex
	closed bool
}

type softwareEncoder struct {
	codec          *astiav.Codec
	codecCtx       *astiav.CodecContext
	frame          *astiav.Frame
	packet         *astiav.Packet
	width          int
	height         int
	r              video.Reader
	nextIsKeyFrame bool

	mu     sync.Mutex
	closed bool
}

func newHardwareEncoder(r video.Reader, p prop.Media, params Params) (*hardwareEncoder, error) {
	if p.FrameRate == 0 {
		p.FrameRate = params.FrameRate
	}
	astiav.SetLogLevel(astiav.LogLevel(astiav.LogLevelWarning))

	var hardwareDeviceType astiav.HardwareDeviceType
	switch params.codecName {
	case "h264_nvenc", "hevc_nvenc", "av1_nvenc":
		hardwareDeviceType = astiav.HardwareDeviceType(astiav.HardwareDeviceTypeCUDA)
	case "vp8_vaapi", "vp9_vaapi", "h264_vaapi", "hevc_vaapi":
		hardwareDeviceType = astiav.HardwareDeviceType(astiav.HardwareDeviceTypeVAAPI)
	}

	hwDevice, err := astiav.CreateHardwareDeviceContext(
		hardwareDeviceType,
		params.hardwareDevice,
		nil,
		0,
	)
	if err != nil {
		return nil, errFailedToCreateHwDevice
	}

	codec := astiav.FindEncoderByName(params.codecName)
	if codec == nil {
		return nil, errCodecNotFound
	}

	codecCtx := astiav.AllocCodecContext(codec)
	if codecCtx == nil {
		return nil, errFailedToCreateCodecCtx
	}

	// Configure codec context
	codecCtx.SetWidth(p.Width)
	codecCtx.SetHeight(p.Height)
	codecCtx.SetTimeBase(astiav.NewRational(1, int(p.FrameRate)))
	codecCtx.SetFramerate(codecCtx.TimeBase().Invert())
	codecCtx.SetBitRate(int64(params.BitRate))
	codecCtx.SetGopSize(params.KeyFrameInterval)
	codecCtx.SetMaxBFrames(0)
	switch params.codecName {
	case "h264_nvenc", "hevc_nvenc", "av1_nvenc":
		codecCtx.SetPixelFormat(astiav.PixelFormat(astiav.PixelFormatCuda))
	case "vp8_vaapi", "vp9_vaapi", "h264_vaapi", "hevc_vaapi":
		codecCtx.SetPixelFormat(astiav.PixelFormat(astiav.PixelFormatVaapi))
	}
	codecOptions := codecCtx.PrivateData().Options()
	switch params.codecName {
	case "av1_nvenc":
		codecCtx.SetProfile(astiav.Profile(astiav.ProfileAv1Main))
		codecOptions.Set("tier", "0", 0)
	case "h264_vaapi":
		codecCtx.SetProfile(astiav.Profile(astiav.ProfileH264Main))
		codecOptions.Set("profile", "main", 0)
		codecOptions.Set("level", "1", 0)
	case "hevc_vaapi":
		codecCtx.SetProfile(astiav.Profile(astiav.ProfileHevcMain))
		codecOptions.Set("profile", "main", 0)
		codecOptions.Set("tier", "main", 0)
		codecOptions.Set("level", "1", 0)
	}
	switch params.codecName {
	case "h264_nvenc", "hevc_nvenc", "av1_nvenc":
		codecOptions.Set("forced-idr", "1", 0)
		codecOptions.Set("zerolatency", "1", 0)
		codecOptions.Set("delay", "0", 0)
		codecOptions.Set("tune", "ll", 0)
		codecOptions.Set("preset", "p1", 0)
		codecOptions.Set("rc", "cbr", 0)
	case "vp8_vaapi", "vp9_vaapi", "h264_vaapi", "hevc_vaapi":
		codecOptions.Set("rc_mode", "CBR", 0)
	}

	// Create hardware frames context
	hwFramesCtx := astiav.AllocHardwareFramesContext(hwDevice)
	hwDevice.Free()
	if hwFramesCtx == nil {
		codecCtx.Free()
		return nil, errFailedToCreateHwFramesCtx
	}

	// Set hardware frames context parameters
	hwFramesCtx.SetWidth(p.Width)
	hwFramesCtx.SetHeight(p.Height)
	switch params.codecName {
	case "h264_nvenc", "hevc_nvenc", "av1_nvenc":
		hwFramesCtx.SetHardwarePixelFormat(astiav.PixelFormat(astiav.PixelFormatCuda))
	case "vp8_vaapi", "vp9_vaapi", "h264_vaapi", "hevc_vaapi":
		hwFramesCtx.SetHardwarePixelFormat(astiav.PixelFormat(astiav.PixelFormatVaapi))
	}
	hwFramesCtx.SetSoftwarePixelFormat(params.pixelFormat)

	err = hwFramesCtx.Initialize()
	if err != nil {
		codecCtx.Free()
		hwFramesCtx.Free()
		return nil, errFailedToInitHwFramesCtx
	}
	codecCtx.SetHardwareFramesContext(hwFramesCtx)

	// Open codec context
	if err := codecCtx.Open(codec, nil); err != nil {
		codecCtx.Free()
		hwFramesCtx.Free()
		return nil, errFailedToOpenCodecCtx
	}

	softwareFrame := astiav.AllocFrame()
	if softwareFrame == nil {
		codecCtx.Free()
		hwFramesCtx.Free()
		return nil, errFailedToAllocFrame
	}

	softwareFrame.SetWidth(p.Width)
	softwareFrame.SetHeight(p.Height)
	softwareFrame.SetPixelFormat(params.pixelFormat)

	err = softwareFrame.AllocBuffer(0)
	if err != nil {
		softwareFrame.Free()
		codecCtx.Free()
		hwFramesCtx.Free()
		return nil, errFailedToAllocSwBuf
	}

	hardwareFrame := astiav.AllocFrame()

	err = hardwareFrame.AllocHardwareBuffer(hwFramesCtx)
	if err != nil {
		softwareFrame.Free()
		hardwareFrame.Free()
		codecCtx.Free()
		hwFramesCtx.Free()
		return nil, errFailedToAllocHwBuf
	}

	packet := astiav.AllocPacket()
	if packet == nil {
		softwareFrame.Free()
		hardwareFrame.Free()
		codecCtx.Free()
		hwFramesCtx.Free()
		return nil, errFailedToAllocPacket
	}

	return &hardwareEncoder{
		codecCtx:       codecCtx,
		hwFramesCtx:    hwFramesCtx,
		frame:          softwareFrame,
		hwFrame:        hardwareFrame,
		packet:         packet,
		width:          p.Width,
		height:         p.Height,
		r:              r,
		nextIsKeyFrame: false,
	}, nil
}

func (e *hardwareEncoder) Controller() codec.EncoderController {
	return e
}

func (e *hardwareEncoder) Read() ([]byte, func(), error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return nil, func() {}, io.EOF
	}

	img, release, err := e.r.Read()
	if err != nil {
		return nil, func() {}, err
	}
	defer release()

	if e.nextIsKeyFrame {
		e.frame.SetPictureType(astiav.PictureType(astiav.PictureTypeI))
		e.hwFrame.SetPictureType(astiav.PictureType(astiav.PictureTypeI))
		e.nextIsKeyFrame = false
	} else {
		e.frame.SetPictureType(astiav.PictureType(astiav.PictureTypeNone))
		e.hwFrame.SetPictureType(astiav.PictureType(astiav.PictureTypeNone))
	}

	err = e.frame.Data().FromImage(img)
	if err != nil {
		return nil, func() {}, err
	}

	err = e.frame.TransferHardwareData(e.hwFrame)
	if err != nil {
		return nil, func() {}, err
	}

	// Send frame to encoder
	if err := e.codecCtx.SendFrame(e.hwFrame); err != nil {
		return nil, func() {}, err
	}

	for {
		if err = e.codecCtx.ReceivePacket(e.packet); err != nil {
			if errors.Is(err, astiav.ErrEof) || errors.Is(err, astiav.ErrEagain) {
				continue
			}
			return nil, func() {}, err
		}
		break
	}

	data := make([]byte, e.packet.Size())
	copy(data, e.packet.Data())
	e.packet.Unref()

	return data, func() {}, nil
}

// ForceKeyFrame forces the next frame to be encoded as a keyframe
func (e *hardwareEncoder) ForceKeyFrame() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.nextIsKeyFrame = true
	return nil
}

func (e *hardwareEncoder) SetBitRate(bitrate int) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.codecCtx.SetBitRate(int64(bitrate))
	return nil
}

func (e *hardwareEncoder) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.packet != nil {
		e.packet.Free()
	}
	if e.frame != nil {
		e.frame.Free()
	}
	if e.hwFrame != nil {
		e.hwFrame.Free()
	}
	if e.codecCtx != nil {
		e.codecCtx.Free()
	}
	if e.hwFramesCtx != nil {
		e.hwFramesCtx.Free()
	}

	e.closed = true
	return nil
}

func newSoftwareEncoder(r video.Reader, p prop.Media, params Params) (*softwareEncoder, error) {
	if p.FrameRate == 0 {
		p.FrameRate = params.FrameRate
	}
	astiav.SetLogLevel(astiav.LogLevel(astiav.LogLevelWarning))

	codec := astiav.FindEncoderByName(params.codecName)
	if codec == nil {
		return nil, errCodecNotFound
	}

	codecCtx := astiav.AllocCodecContext(codec)
	if codecCtx == nil {
		return nil, errFailedToCreateCodecCtx
	}

	// Configure codec context
	codecCtx.SetWidth(p.Width)
	codecCtx.SetHeight(p.Height)
	codecCtx.SetTimeBase(astiav.NewRational(1, int(p.FrameRate)))
	codecCtx.SetFramerate(codecCtx.TimeBase().Invert())
	codecCtx.SetPixelFormat(astiav.PixelFormat(astiav.PixelFormatYuv420P))
	codecCtx.SetBitRate(int64(params.BitRate))
	codecCtx.SetGopSize(params.KeyFrameInterval)
	codecCtx.SetMaxBFrames(0)
	codecOptions := codecCtx.PrivateData().Options()
	codecOptions.Set("preset", "ultrafast", 0)
	codecOptions.Set("tune", "zerolatency", 0)
	codecCtx.SetFlags(astiav.CodecContextFlags(astiav.CodecContextFlagLowDelay))

	// Open codec context
	if err := codecCtx.Open(codec, nil); err != nil {
		codecCtx.Free()
		return nil, errFailedToOpenCodecCtx
	}

	softwareFrame := astiav.AllocFrame()
	if softwareFrame == nil {
		codecCtx.Free()
		return nil, errFailedToAllocFrame
	}

	softwareFrame.SetWidth(p.Width)
	softwareFrame.SetHeight(p.Height)
	softwareFrame.SetPixelFormat(astiav.PixelFormat(astiav.PixelFormatYuv420P))

	err := softwareFrame.AllocBuffer(0)
	if err != nil {
		softwareFrame.Free()
		codecCtx.Free()
		return nil, errFailedToAllocSwBuf
	}

	packet := astiav.AllocPacket()
	if packet == nil {
		softwareFrame.Free()
		codecCtx.Free()
		return nil, errFailedToAllocPacket
	}

	return &softwareEncoder{
		codecCtx:       codecCtx,
		frame:          softwareFrame,
		packet:         packet,
		width:          p.Width,
		height:         p.Height,
		r:              video.ToI420(r),
		nextIsKeyFrame: false,
	}, nil
}

func (e *softwareEncoder) Read() ([]byte, func(), error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return nil, func() {}, io.EOF
	}
	img, release, err := e.r.Read()
	if err != nil {
		return nil, func() {}, err
	}
	defer release()
	if e.nextIsKeyFrame {
		e.frame.SetPictureType(astiav.PictureType(astiav.PictureTypeI))
		e.nextIsKeyFrame = false
	} else {
		e.frame.SetPictureType(astiav.PictureType(astiav.PictureTypeNone))
	}
	err = e.frame.Data().FromImage(img)
	if err != nil {
		return nil, func() {}, err
	}
	if err := e.codecCtx.SendFrame(e.frame); err != nil {
		return nil, func() {}, err
	}
	for {
		if err = e.codecCtx.ReceivePacket(e.packet); err != nil {
			if errors.Is(err, astiav.ErrEof) || errors.Is(err, astiav.ErrEagain) {
				continue
			}
			return nil, func() {}, err
		}
		break
	}
	data := make([]byte, e.packet.Size())
	copy(data, e.packet.Data())
	e.packet.Unref()
	return data, func() {}, nil
}

func (e *softwareEncoder) Controller() codec.EncoderController {
	return e
}

func (e *softwareEncoder) ForceKeyFrame() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.nextIsKeyFrame = true
	return nil
}

func (e *softwareEncoder) SetBitRate(bitrate int) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.codecCtx.SetBitRate(int64(bitrate))
	return nil
}

func (e *softwareEncoder) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.packet != nil {
		e.packet.Free()
	}
	if e.frame != nil {
		e.frame.Free()
	}
	if e.codecCtx != nil {
		e.codecCtx.Free()
	}

	e.closed = true
	return nil
}
