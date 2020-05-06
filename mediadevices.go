package mediadevices

import (
	"fmt"
	"math"

	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/prop"
)

var errNotFound = fmt.Errorf("failed to find the best driver that fits the constraints")

// GetDisplayMedia prompts the user to select and grant permission to capture the contents
// of a display or portion thereof (such as a window) as a MediaStream.
// Reference: https://developer.mozilla.org/en-US/docs/Web/API/MediaDevices/getDisplayMedia
func GetDisplayMedia(constraints MediaStreamConstraints) (MediaStream, error) {
	tracks := make([]Track, 0)

	cleanTracks := func() {
		for _, t := range tracks {
			t.Stop()
		}
	}

	if constraints.Video != nil {
		var p prop.Media
		constraints.Video(&p)
		track, err := selectScreen(p)
		if err != nil {
			cleanTracks()
			return nil, err
		}

		tracks = append(tracks, track)
	}

	s, err := NewMediaStream(tracks...)
	if err != nil {
		cleanTracks()
		return nil, err
	}

	return s, nil
}

// GetUserMedia prompts the user for permission to use a media input which produces a MediaStream
// with tracks containing the requested types of media.
// Reference: https://developer.mozilla.org/en-US/docs/Web/API/MediaDevices/getUserMedia
func GetUserMedia(constraints MediaStreamConstraints) (MediaStream, error) {
	tracks := make([]Track, 0)

	cleanTracks := func() {
		for _, t := range tracks {
			t.Stop()
		}
	}

	if constraints.Video != nil {
		var p prop.Media
		constraints.Video(&p)
		track, err := selectVideo(p)
		if err != nil {
			cleanTracks()
			return nil, err
		}

		tracks = append(tracks, track)
	}

	if constraints.Audio != nil {
		var p prop.Media
		constraints.Audio(&p)
		track, err := selectAudio(p)
		if err != nil {
			cleanTracks()
			return nil, err
		}

		tracks = append(tracks, track)
	}

	s, err := NewMediaStream(tracks...)
	if err != nil {
		cleanTracks()
		return nil, err
	}

	return s, nil
}

func queryDriverProperties(filter driver.FilterFn) map[driver.Driver][]prop.Media {
	var needToClose []driver.Driver
	drivers := driver.GetManager().Query(filter)
	m := make(map[driver.Driver][]prop.Media)

	for _, d := range drivers {
		if d.Status() == driver.StateClosed {
			err := d.Open()
			if err != nil {
				// Skip this driver if we failed to open because we can't get the properties
				continue
			}
			needToClose = append(needToClose, d)
		}

		m[d] = d.Properties()
	}

	for _, d := range needToClose {
		// Since it was closed, we should close it to avoid a leak
		d.Close()
	}

	return m
}

// select implements SelectSettings algorithm.
// Reference: https://w3c.github.io/mediacapture-main/#dfn-selectsettings
func selectBestDriver(filter driver.FilterFn, constraints prop.Media) (driver.Driver, prop.Media, error) {
	var bestDriver driver.Driver
	var bestProp prop.Media
	minFitnessDist := math.Inf(1)

	driverProperties := queryDriverProperties(filter)
	for d, props := range driverProperties {
		priority := float64(d.Info().Priority)
		for _, p := range props {
			fitnessDist := constraints.FitnessDistance(p) - priority
			if fitnessDist < minFitnessDist {
				minFitnessDist = fitnessDist
				bestDriver = d
				bestProp = p
			}
		}
	}

	if bestDriver == nil {
		return nil, prop.Media{}, errNotFound
	}

	constraints.Merge(bestProp)
	return bestDriver, constraints, nil
}

func selectAudio(constraints prop.Media) (Track, error) {
	typeFilter := driver.FilterAudioRecorder()
	filter := typeFilter
	if constraints.DeviceID != "" {
		idFilter := driver.FilterID(constraints.DeviceID)
		filter = driver.FilterAnd(typeFilter, idFilter)
	}

	d, c, err := selectBestDriver(filter, constraints)
	if err != nil {
		return nil, err
	}

	return newAudioTrack(d, c)
}

func selectVideo(constraints prop.Media) (Track, error) {
	typeFilter := driver.FilterVideoRecorder()
	notScreenFilter := driver.FilterNot(driver.FilterDeviceType(driver.Screen))
	filter := driver.FilterAnd(typeFilter, notScreenFilter)
	if constraints.DeviceID != "" {
		idFilter := driver.FilterID(constraints.DeviceID)
		filter = driver.FilterAnd(typeFilter, notScreenFilter, idFilter)
	}

	d, c, err := selectBestDriver(filter, constraints)
	if err != nil {
		return nil, err
	}

	return newVideoTrack(d, c)
}

func selectScreen(constraints prop.Media) (Track, error) {
	typeFilter := driver.FilterVideoRecorder()
	screenFilter := driver.FilterDeviceType(driver.Screen)
	filter := driver.FilterAnd(typeFilter, screenFilter)
	if constraints.DeviceID != "" {
		idFilter := driver.FilterID(constraints.DeviceID)
		filter = driver.FilterAnd(typeFilter, screenFilter, idFilter)
	}

	d, c, err := selectBestDriver(filter, constraints)
	if err != nil {
		return nil, err
	}

	return newVideoTrack(d, c)
}

func EnumerateDevices() []MediaDeviceInfo {
	drivers := driver.GetManager().Query(
		driver.FilterFn(func(driver.Driver) bool { return true }))
	info := make([]MediaDeviceInfo, 0, len(drivers))
	for _, d := range drivers {
		var kind MediaDeviceType
		switch {
		case driver.FilterVideoRecorder()(d):
			kind = VideoInput
		case driver.FilterAudioRecorder()(d):
			kind = AudioInput
		default:
			continue
		}
		driverInfo := d.Info()
		info = append(info, MediaDeviceInfo{
			DeviceID:   d.ID(),
			Kind:       kind,
			Label:      driverInfo.Label,
			DeviceType: driverInfo.DeviceType,
		})
	}
	return info
}
