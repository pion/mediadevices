package codec

import (
	"io"

	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
)

type VideoEncoderBuilder func(r video.Reader, prop video.AdvancedProperty) (io.ReadCloser, error)
type AudioEncoderBuilder func(r audio.Reader, inProp, outProp audio.AdvancedProperty) (io.ReadCloser, error)
