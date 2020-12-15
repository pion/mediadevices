// +build !dynamic

package openh264

//#cgo CFLAGS: -I${SRCDIR}/include
//#cgo CXXFLAGS: -I${SRCDIR}/include
//#cgo linux,arm LDFLAGS: ${SRCDIR}/lib/libopenh264_linux_armv7.a
//#cgo linux,arm64 LDFLAGS: ${SRCDIR}/lib/libopenh264_linux_arm64.a
//#cgo linux,amd64 LDFLAGS: ${SRCDIR}/lib/libopenh264_linux_x64.a
//#cgo darwin,amd64 LDFLAGS: ${SRCDIR}/lib/libopenh264_darwin_x64.a
//#cgo windows,amd64 LDFLAGS: ${SRCDIR}/lib/libopenh264_windows_x64.a -lssp
import "C"
