// +build !dynamic

package openh264

//#cgo CFLAGS: -I${SRCDIR}/include
//#cgo CXXFLAGS: -I${SRCDIR}/include
//#cgo linux,arm LDFLAGS: ${SRCDIR}/lib/libopenh264_linux_armv7.a
//#cgo linux,arm64 LDFLAGS: ${SRCDIR}/lib/libopenh264_linux_arm64.a
//#cgo linux,amd64 LDFLAGS: ${SRCDIR}/lib/libopenh264_linux_x64.a
//#cgo darwin,amd64 LDFLAGS: ${SRCDIR}/lib/libopenh264_darwin_x64.a
//#cgo windows,amd64 LDFLAGS: ${SRCDIR}/lib/libopenh264_windows_x64.a -lssp -DGOOS_WINDOWS=1
// #ifdef GOOS_WINDOWS
// #include <stdio.h>
// #include <crtdefs.h>
// #if __MINGW64_VERSION_MAJOR < 6
// // Workaround for the breaking ABI change between MinGW 5 and 6
//   FILE *__cdecl __acrt_iob_func(unsigned index) { return &(__iob_func()[index]); }
//   typedef FILE *__cdecl (*_f__acrt_iob_func)(unsigned index);
//   _f__acrt_iob_func __MINGW_IMP_SYMBOL(__acrt_iob_func) = __acrt_iob_func;
// #endif
// #endif
import "C"
