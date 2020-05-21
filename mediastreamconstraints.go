package mediadevices

import (
	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

type MediaStreamConstraints struct {
	Audio MediaOption
	Video MediaOption
}

// MediaTrackConstraints represents https://w3c.github.io/mediacapture-main/#dom-mediatrackconstraints
type MediaTrackConstraints struct {
	prop.MediaConstraints
	Enabled bool
	// VideoEncoderBuilders are codec builders that are used for encoding the video
	// and later being used for sending the appropriate RTP payload type.
	//
	// If one encoder builder fails to build the codec, the next builder will be used,
	// repeating until a codec builds. If no builders build successfully, an error is returned.
	VideoEncoderBuilders []codec.VideoEncoderBuilder
	// AudioEncoderBuilders are codec builders that are used for encoding the audio
	// and later being used for sending the appropriate RTP payload type.
	//
	// If one encoder builder fails to build the codec, the next builder will be used,
	// repeating until a codec builds. If no builders build successfully, an error is returned.
	AudioEncoderBuilders []codec.AudioEncoderBuilder
	// VideoTransform will be used to transform the video that's coming from the driver.
	// So, basically it'll look like following: driver -> VideoTransform -> codec
	VideoTransform video.TransformFunc
	// AudioTransform will be used to transform the audio that's coming from the driver.
	// So, basically it'll look like following: driver -> AudioTransform -> code
	AudioTransform audio.TransformFunc

	selectedMedia prop.Media
}

type MediaOption func(*MediaTrackConstraints)
