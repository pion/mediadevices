package x264

// Params stores libx264 specific encoding parameters.
type Params struct {
	// Quality of the encoding [0-9].
	// Larger value results higher quality and higher CPU usage.
	// It depends on the selected codec.
	Quality int
}
