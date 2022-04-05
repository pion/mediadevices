// +build !dynamic

package opus

//#cgo CFLAGS: -I${SRCDIR}/include
//#cgo CXXFLAGS: -I${SRCDIR}/include
//#cgo linux,arm LDFLAGS: ${SRCDIR}/lib/libopus-linux-armv7.a -lm
//#cgo linux,arm64 LDFLAGS: ${SRCDIR}/lib/libopus-linux-arm64.a -lm
//#cgo linux,amd64 LDFLAGS: ${SRCDIR}/lib/libopus-linux-x64.a -lm
//#cgo darwin,amd64 LDFLAGS: ${SRCDIR}/lib/libopus-darwin-x64.a
//#cgo darwin,arm64 LDFLAGS: ${SRCDIR}/lib/libopus-darwin-arm64.a
//#cgo windows,amd64 LDFLAGS: ${SRCDIR}/lib/libopus-windows-x64.a
import "C"
