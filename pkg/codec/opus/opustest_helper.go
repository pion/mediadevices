//go:build opustest

package opus

/*
#include <opus.h>
*/
import "C"

// newTestEncoder creates a raw OpusEncoder engine for tests that need to
// bypass the production reader pipeline (which buffers audio to valid frame
// sizes). This file is excluded from normal builds via the "opustest" build
// tag. It lives in a non-test file because Go forbids the use of cgo in
// _test.go files.
func newTestEncoder(sampleRate, channels int) *C.OpusEncoder {
	var cerror C.int
	engine := C.opus_encoder_create(
		C.opus_int32(sampleRate),
		C.int(channels),
		C.OPUS_APPLICATION_VOIP,
		&cerror,
	)
	if engine == nil || cerror != C.OPUS_OK {
		return nil
	}
	return engine
}

func destroyTestEncoder(engine *C.OpusEncoder) {
	if engine != nil {
		C.opus_encoder_destroy(engine)
	}
}
