//go:build cgo && !darwin && !linux

package cgoheap

// HeapInUse returns the number of live bytes in the process default malloc
// zone. Native heap measurement is only implemented on darwin and linux;
// other platforms return 0.
func HeapInUse() uint64 {
	return 0
}

// Supported reports whether HeapInUse returns meaningful data on this
// platform.
func Supported() bool {
	return false
}
