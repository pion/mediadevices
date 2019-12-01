package mediadevices

type MediaStreamConstraints struct {
	Audio MediaTrackConstraints
	Video MediaTrackConstraints
}

type MediaTrackConstraints bool
