//go:build nomicrophone
// +build nomicrophone

package microphone

// This stub file is used when building with the 'nomicrophone' build tag.
// Use this when cross-compiling or when malgo (miniaudio) dependencies are not available.
//
// To build without microphone support:
//   go build -tags nomicrophone
//
// This is particularly useful for:
// - Cross-compilation where CGO dependencies are unavailable
// - Environments without audio system development libraries
// - Minimal builds that only need camera/video support
