package ffmpeg

import (
	"github.com/asticode/go-astiav"
	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

type Params struct {
	codec.BaseParams
	codecName      string
	hardwareDevice string
	pixelFormat    astiav.PixelFormat
	FrameRate      float32
}

type VP8Params struct {
	Params
}

func NewVP8VAAPIParams(hardwareDevice string, pixelFormat astiav.PixelFormat) (VP8Params, error) {
	return VP8Params{
		Params: Params{
			codecName:      "vp8_vaapi",
			hardwareDevice: hardwareDevice,
			pixelFormat:    pixelFormat,
		},
	}, nil
}

func (p *VP8Params) RTPCodec() *codec.RTPCodec {
	return codec.NewRTPVP8Codec(90000)
}

func (p *VP8Params) BuildVideoEncoder(r video.Reader, property prop.Media) (codec.ReadCloser, error) {
	readCloser, err := newHardwareEncoder(r, property, p.Params)
	if err != nil {
		return nil, err
	}
	return readCloser, nil
}

type VP9Params struct {
	Params
}

func NewVP9VAAPIParams(hardwareDevice string, pixelFormat astiav.PixelFormat) (VP8Params, error) {
	return VP8Params{
		Params: Params{
			codecName:      "vp9_vaapi",
			hardwareDevice: hardwareDevice,
			pixelFormat:    pixelFormat,
		},
	}, nil
}

func (p *VP9Params) RTPCodec() *codec.RTPCodec {
	return codec.NewRTPVP9Codec(90000)
}

func (p *VP9Params) BuildVideoEncoder(r video.Reader, property prop.Media) (codec.ReadCloser, error) {
	readCloser, err := newHardwareEncoder(r, property, p.Params)
	if err != nil {
		return nil, err
	}
	return readCloser, nil
}

type H264Params struct {
	Params
}

func NewH264NVENCParams(hardwareDevice string, pixelFormat astiav.PixelFormat) (H264Params, error) {
	return H264Params{
		Params: Params{
			codecName:      "h264_nvenc",
			hardwareDevice: hardwareDevice,
			pixelFormat:    pixelFormat,
		},
	}, nil
}

func NewH264VAAPIParams(hardwareDevice string, pixelFormat astiav.PixelFormat) (H264Params, error) {
	return H264Params{
		Params: Params{
			codecName:      "h264_vaapi",
			hardwareDevice: hardwareDevice,
			pixelFormat:    pixelFormat,
		},
	}, nil
}

// RTPCodec represents the codec metadata
func (p *H264Params) RTPCodec() *codec.RTPCodec {
	return codec.NewRTPH264Codec(90000)
}

func (p *H264Params) BuildVideoEncoder(r video.Reader, property prop.Media) (codec.ReadCloser, error) {
	readCloser, err := newHardwareEncoder(r, property, p.Params)
	if err != nil {
		return nil, err
	}
	return readCloser, nil
}

type H264SoftwareParams struct {
	Params
}

func NewH264X264Params() (H264SoftwareParams, error) {
	return H264SoftwareParams{
		Params: Params{
			codecName: "libx264",
		},
	}, nil
}

func (p *H264SoftwareParams) RTPCodec() *codec.RTPCodec {
	return codec.NewRTPH264Codec(90000)
}

func (p *H264SoftwareParams) BuildVideoEncoder(r video.Reader, property prop.Media) (codec.ReadCloser, error) {
	readCloser, err := newSoftwareEncoder(r, property, p.Params)
	if err != nil {
		return nil, err
	}
	return readCloser, nil
}

type H265Params struct {
	Params
}

func NewH265NVENCParams(hardwareDevice string, pixelFormat astiav.PixelFormat) (H265Params, error) {
	return H265Params{
		Params: Params{
			codecName:      "hevc_nvenc",
			hardwareDevice: hardwareDevice,
			pixelFormat:    pixelFormat,
		},
	}, nil
}

func NewH265VAAPIParams(hardwareDevice string, pixelFormat astiav.PixelFormat) (H265Params, error) {
	return H265Params{
		Params: Params{
			codecName:      "hevc_vaapi",
			hardwareDevice: hardwareDevice,
			pixelFormat:    pixelFormat,
		},
	}, nil
}

func (p *H265Params) RTPCodec() *codec.RTPCodec {
	return codec.NewRTPH265Codec(90000)
}

func (p *H265Params) BuildVideoEncoder(r video.Reader, property prop.Media) (codec.ReadCloser, error) {
	readCloser, err := newHardwareEncoder(r, property, p.Params)
	if err != nil {
		return nil, err
	}
	return readCloser, nil
}

type AV1Params struct {
	Params
}

func NewAV1NVENCParams(hardwareDevice string, pixelFormat astiav.PixelFormat) (AV1Params, error) {
	return AV1Params{
		Params: Params{
			codecName:      "av1_nvenc",
			hardwareDevice: hardwareDevice,
			pixelFormat:    pixelFormat,
		},
	}, nil
}

func (p *AV1Params) RTPCodec() *codec.RTPCodec {
	return codec.NewRTPAV1Codec(90000)
}

func (p *AV1Params) BuildVideoEncoder(r video.Reader, property prop.Media) (codec.ReadCloser, error) {
	readCloser, err := newHardwareEncoder(r, property, p.Params)
	if err != nil {
		return nil, err
	}
	return readCloser, nil
}
