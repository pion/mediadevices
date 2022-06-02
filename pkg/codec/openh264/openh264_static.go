// +build !dynamic

package openh264

//#cgo CFLAGS: -I${SRCDIR}/include
//#cgo CXXFLAGS: -I${SRCDIR}/include
//#cgo linux,arm LDFLAGS: ${SRCDIR}/lib/libopenh264-linux-armv7.a
//#cgo linux,arm64 LDFLAGS: ${SRCDIR}/lib/libopenh264-linux-arm64.a
//#cgo linux,amd64 LDFLAGS: ${SRCDIR}/lib/libopenh264-linux-x64.a
//#cgo darwin,amd64 LDFLAGS: ${SRCDIR}/lib/libopenh264-darwin-x64.a
//#cgo darwin,arm64 LDFLAGS: ${SRCDIR}/lib/libopenh264-darwin-arm64.a
//#cgo windows,amd64 LDFLAGS: ${SRCDIR}/lib/libopenh264-windows-x64.a -lssp
import "C"
