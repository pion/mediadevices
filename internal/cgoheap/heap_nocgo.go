//go:build !cgo

package cgoheap

// HeapInUse returns the number of live bytes in the process default malloc
// zone. Without cgo there is no native heap measurement; 0 is returned.
func HeapInUse() uint64 {
	return 0
}

// Supported reports whether HeapInUse returns meaningful data on this
// platform.
func Supported() bool {
	return false
}
