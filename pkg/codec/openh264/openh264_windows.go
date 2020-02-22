// +build !dynamic

package openh264

// #cgo LDFLAGS: ${SRCDIR}/../../../cvendor/lib/openh264/libopenh264.x86_64-windows.a -lssp
// #include <stdio.h>
// #include <crtdefs.h>
// #if __MINGW64_VERSION_MAJOR < 6
// // Workaround for the breaking ABI change between MinGW 5 and 6
//   FILE *__cdecl __acrt_iob_func(unsigned index) { return &(__iob_func()[index]); }
//   typedef FILE *__cdecl (*_f__acrt_iob_func)(unsigned index);
//   _f__acrt_iob_func __MINGW_IMP_SYMBOL(__acrt_iob_func) = __acrt_iob_func;
// #endif
import "C"
