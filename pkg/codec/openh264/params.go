package openh264

// #include <openh264/codec_api.h>
import "C"

import (
	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

// Params stores libopenh264 specific encoding parameters.
type Params struct {
	codec.BaseParams
	UsageType           UsageTypeEnum
	RCMode              RCModeEnum
	EnableFrameSkip     bool
	MaxNalSize          uint
	IntraPeriod         uint
	MultipleThreadIdc   int
	SliceNum            uint
	SliceMode           SliceModeEnum
	SliceSizeConstraint uint
}

type UsageTypeEnum int

const (
	CameraVideoRealTime      UsageTypeEnum = C.CAMERA_VIDEO_REAL_TIME   ///< camera video for real-time communication
	ScreenContentRealTime    UsageTypeEnum = C.SCREEN_CONTENT_REAL_TIME ///< screen content signal
	CameraVideoNonRealTime   UsageTypeEnum = C.CAMERA_VIDEO_NON_REAL_TIME
	ScreenContentNonRealTime UsageTypeEnum = C.SCREEN_CONTENT_NON_REAL_TIME
	InputContentTypeAll      UsageTypeEnum = C.INPUT_CONTENT_TYPE_ALL
)

type RCModeEnum int

const (
	RCQualityMode         RCModeEnum = C.RC_QUALITY_MODE           ///< quality mode
	RCBitrateMode         RCModeEnum = C.RC_BITRATE_MODE           ///< bitrate mode
	RCBufferbaseedMode    RCModeEnum = C.RC_BUFFERBASED_MODE       ///< no bitrate control,only using buffer status,adjust the video quality
	RCTimestampMode       RCModeEnum = C.RC_TIMESTAMP_MODE         //rate control based timestamp
	RCBitrateModePostSkip RCModeEnum = C.RC_BITRATE_MODE_POST_SKIP ///< this is in-building RC MODE, WILL BE DELETED after algorithm tuning!
	RCOffMode             RCModeEnum = C.RC_OFF_MODE               ///< rate control off mode
)

type SliceModeEnum uint

const (
	SMSingleSlice      SliceModeEnum = C.SM_SINGLE_SLICE      ///< | SliceNum==1
	SMFixedslcnumSlice SliceModeEnum = C.SM_FIXEDSLCNUM_SLICE ///< | according to SliceNum        | enabled dynamic slicing for multi-thread
	SMRasterSlice      SliceModeEnum = C.SM_RASTER_SLICE      ///< | according to SlicesAssign    | need input of MB numbers each slice. In addition, if other constraint in SSliceArgument is presented, need to follow the constraints. Typically if MB num and slice size are both constrained, re-encoding may be involved.
	SMSizelimitedSlice SliceModeEnum = C.SM_SIZELIMITED_SLICE ///< | according to SliceSize       | slicing according to size, the slicing will be dynamic(have no idea about slice_nums until encoding current frame)
)

// NewParams returns default openh264 codec specific parameters.
func NewParams() (Params, error) {
	return Params{
		BaseParams: codec.BaseParams{
			BitRate: 100000,
		},
		UsageType:           CameraVideoRealTime,
		RCMode:              RCBitrateMode,
		EnableFrameSkip:     true,
		MaxNalSize:          0,
		IntraPeriod:         30,
		MultipleThreadIdc:   0, // Defaults to 0, so that it'll automatically use multi threads when needed
		SliceNum:            1, // Defaults to single NAL unit mode
		SliceMode:           SMSizelimitedSlice,
		SliceSizeConstraint: 12800,
	}, nil
}

// RTPCodec represents the codec metadata
func (p *Params) RTPCodec() *codec.RTPCodec {
	return codec.NewRTPH264Codec(90000)
}

// BuildVideoEncoder builds openh264 encoder with given params
func (p *Params) BuildVideoEncoder(r video.Reader, property prop.Media) (codec.ReadCloser, error) {
	return newEncoder(r, property, *p)
}
