package mediadevices

import (
	"fmt"
	"math"
	"strings"

	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/prop"
)

var errNotFound = fmt.Errorf("failed to find the best driver that fits the constraints")

// GetDisplayMedia prompts the user to select and grant permission to capture the contents
// of a display or portion thereof (such as a window) as a MediaStream.
// Reference: https://developer.mozilla.org/en-US/docs/Web/API/MediaDevices/getDisplayMedia
func GetDisplayMedia(constraints MediaStreamConstraints) (MediaStream, error) {
	trackers := make([]Track, 0)

	cleanTrackers := func() {
		for _, t := range trackers {
			t.Close()
		}
	}

	var videoConstraints MediaTrackConstraints
	if constraints.Video != nil {
		constraints.Video(&videoConstraints)
		tracker, err := selectScreen(videoConstraints, constraints.Codec)
		if err != nil {
			cleanTrackers()
			return nil, err
		}

		trackers = append(trackers, tracker)
	}

	s, err := NewMediaStream(trackers...)
	if err != nil {
		cleanTrackers()
		return nil, err
	}

	return s, nil
}

// GetUserMedia prompts the user for permission to use a media input which produces a MediaStream
// with tracks containing the requested types of media.
// Reference: https://developer.mozilla.org/en-US/docs/Web/API/MediaDevices/getUserMedia
func GetUserMedia(constraints MediaStreamConstraints) (MediaStream, error) {
	// TODO: It should return media stream based on constraints
	trackers := make([]Track, 0)

	cleanTrackers := func() {
		for _, t := range trackers {
			t.Close()
		}
	}

	var videoConstraints, audioConstraints MediaTrackConstraints
	if constraints.Video != nil {
		constraints.Video(&videoConstraints)
		tracker, err := selectVideo(videoConstraints, constraints.Codec)
		if err != nil {
			cleanTrackers()
			return nil, err
		}

		trackers = append(trackers, tracker)
	}

	if constraints.Audio != nil {
		constraints.Audio(&audioConstraints)
		tracker, err := selectAudio(audioConstraints, constraints.Codec)
		if err != nil {
			cleanTrackers()
			return nil, err
		}

		trackers = append(trackers, tracker)
	}

	s, err := NewMediaStream(trackers...)
	if err != nil {
		cleanTrackers()
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
func selectBestDriver(filter driver.FilterFn, constraints MediaTrackConstraints) (driver.Driver, MediaTrackConstraints, error) {
	var bestDriver driver.Driver
	var bestProp prop.Media
	var foundPropertiesLog []string
	minFitnessDist := math.Inf(1)

	foundPropertiesLog = append(foundPropertiesLog, "\n============ Found Properties ============")
	driverProperties := queryDriverProperties(filter)
	for d, props := range driverProperties {
		priority := float64(d.Info().Priority)
		for _, p := range props {
			foundPropertiesLog = append(foundPropertiesLog, p.String())
			fitnessDist, ok := constraints.MediaConstraints.FitnessDistance(p)
			if !ok {
				continue
			}
			fitnessDist -= priority
			if fitnessDist < minFitnessDist {
				minFitnessDist = fitnessDist
				bestDriver = d
				bestProp = p
			}
		}
	}

	foundPropertiesLog = append(foundPropertiesLog, "=============== Constraints ==============")
	foundPropertiesLog = append(foundPropertiesLog, constraints.String())
	foundPropertiesLog = append(foundPropertiesLog, "================ Best Fit ================")

	if bestDriver == nil {
		foundPropertiesLog = append(foundPropertiesLog, "Not found")
		logger.Debug(strings.Join(foundPropertiesLog, "\n\n"))
		return nil, MediaTrackConstraints{}, errNotFound
	}

	foundPropertiesLog = append(foundPropertiesLog, bestProp.String())
	logger.Debug(strings.Join(foundPropertiesLog, "\n\n"))
	constraints.selectedMedia = prop.Media{}
	constraints.selectedMedia.MergeConstraints(constraints.MediaConstraints)
	constraints.selectedMedia.Merge(bestProp)
	return bestDriver, constraints, nil
}

func selectAudio(constraints MediaTrackConstraints, selector *CodecSelector) (Track, error) {
	typeFilter := driver.FilterAudioRecorder()

	d, c, err := selectBestDriver(typeFilter, constraints)
	if err != nil {
		return nil, err
	}

	return newTrackFromDriver(d, c, selector)
}
func selectVideo(constraints MediaTrackConstraints, selector *CodecSelector) (Track, error) {
	typeFilter := driver.FilterVideoRecorder()
	notScreenFilter := driver.FilterNot(driver.FilterDeviceType(driver.Screen))
	filter := driver.FilterAnd(typeFilter, notScreenFilter)

	d, c, err := selectBestDriver(filter, constraints)
	if err != nil {
		return nil, err
	}

	return newTrackFromDriver(d, c, selector)
}

func selectScreen(constraints MediaTrackConstraints, selector *CodecSelector) (Track, error) {
	typeFilter := driver.FilterVideoRecorder()
	screenFilter := driver.FilterDeviceType(driver.Screen)
	filter := driver.FilterAnd(typeFilter, screenFilter)

	d, c, err := selectBestDriver(filter, constraints)
	if err != nil {
		return nil, err
	}

	return newTrackFromDriver(d, c, selector)
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
