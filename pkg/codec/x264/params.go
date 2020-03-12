package x264

// Params stores libx264 specific encoding parameters.
type Params struct {
	// Faster preset has lower CPU usage but lower quality
	Preset Preset
}

// Preset represents a set of default configurations from libx264
type Preset int

const (
	PresetUltrafast Preset = iota
	PresetSuperfast
	PresetVeryfast
	PresetFaster
	PresetFast
	PresetMedium
	PresetSlow
	PresetSlower
	PresetVeryslow
	PresetPlacebo
)
