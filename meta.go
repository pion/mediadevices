package mediadevices

import (
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

// detectCurrentVideoProp is a small helper to get current video property
func detectCurrentVideoProp(broadcaster *video.Broadcaster) (prop.Media, error) {
	var currentProp prop.Media

	// Since broadcaster has a ring buffer internally, a new reader will either read the last
	// buffered frame or a new frame from the source. This also implies that no frame will be lost
	// in any case.
	metaReader := broadcaster.NewReader(false)
	metaReader = video.DetectChanges(0, func(p prop.Media) { currentProp = p })(metaReader)
	_, _, err := metaReader.Read()

	return currentProp, err
}

// detectCurrentAudioProp is a small helper to get current audio property
func detectCurrentAudioProp(broadcaster *audio.Broadcaster) (prop.Media, error) {
	var currentProp prop.Media

	// Since broadcaster has a ring buffer internally, a new reader will either read the last
	// buffered frame or a new frame from the source. This also implies that no frame will be lost
	// in any case.
	metaReader := broadcaster.NewReader(false)
	metaReader = audio.DetectChanges(0, func(p prop.Media) { currentProp = p })(metaReader)
	_, _, err := metaReader.Read()

	return currentProp, err
}
