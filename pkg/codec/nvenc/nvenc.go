package nvenc

import (
	"image"
	"io"
	"sync"

	bdandyNvenc "github.com/bdandy/go-nvenc/v8"
	"github.com/bdandy/go-nvenc/v8/guid"
	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

type encoder struct {
	engine *bdandyNvenc.Encoder
	r      video.Reader
	mu     sync.Mutex
	closed bool
}

func newEncoder(r video.Reader, p prop.Media, params Params) (codec.ReadCloser, error) {
	if params.KeyFrameInterval == 0 {
		params.KeyFrameInterval = 60
	}
	enc, err := bdandyNvenc.NewEncoder(10000)
	if err != nil {
		return nil, err
	}
	enc.SetCodec(guid.CodecH264Guid)
	enc.SetPreset(guid.PresetLowLatencyDefaultGuid)
	enc.SetResolution(uint32(p.Width), uint32(p.Height))
	enc.SetFrameRate(uint32(p.FrameRate), 1)
	enc.Config().SetGOPLen(0xffffffff)
	err = enc.InitializeEncoder(0x100, 0x100) // 0x100 is BufferFormatYUV420
	if err != nil {
		return nil, err
	}
	e := encoder{
		engine: enc,
		r:      video.ToI420(r),
	}
	return &e, nil
}

func (e *encoder) Read() ([]byte, func(), error) {
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
	yuvImg := img.(*image.YCbCr)
	rawFrame := make([]byte, len(yuvImg.Y)+len(yuvImg.Cb)+len(yuvImg.Cr))
	rawFrame = append(rawFrame, yuvImg.Y...)
	rawFrame = append(rawFrame, yuvImg.Cb...)
	rawFrame = append(rawFrame, yuvImg.Cr...)
	encodedFrame, err := e.engine.Encode(rawFrame)
	if err != nil {
		return nil, func() {}, err
	}
	return encodedFrame, func() {}, nil
}

func (e *encoder) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return nil
	}
	e.closed = true
	return e.engine.Destroy()
}

func (e *encoder) Controller() codec.EncoderController {
	return e
}

func (e *encoder) SetBitRate(bitrate int) error {
	// TODO
	return nil
}

func (e *encoder) ForceKeyFrame() error {
	// TODO
	return nil
}
