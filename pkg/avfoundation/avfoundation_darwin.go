// Package avfoundation provides AVFoundation binding for Go
package avfoundation

// #cgo CFLAGS: -x objective-c
// #cgo LDFLAGS: -framework AVFoundation -framework Foundation -framework CoreMedia -framework CoreVideo
// #include "AVFoundationBind/AVFoundationBind.h"
// #include "AVFoundationBind/AVFoundationBind.m"
// extern void onData(void*, void*, int);
// void onDataBridge(void *userData, void *buf, int len) {
// 	onData(userData, buf, len);
// }
import "C"
import (
	"context"
	"fmt"
	"io"
	"sync"
	"unsafe"

	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/prop"
)

type MediaType C.AVBindMediaType

const (
	Video = MediaType(C.AVBindMediaTypeVideo)
	Audio = MediaType(C.AVBindMediaTypeAudio)
)

// Device represents a metadata that later can be used to retrieve back the
// underlying device given by AVFoundation
type Device struct {
	// UID is a unique identifier for a device
	UID     string
	cDevice C.AVBindDevice
	Name    string
}

func frameFormatToAVBind(f frame.Format) (C.AVBindFrameFormat, bool) {
	switch f {
	case frame.FormatI420:
		return C.AVBindFrameFormatI420, true
	case frame.FormatNV21:
		return C.AVBindFrameFormatNV21, true
	case frame.FormatNV12:
		return C.AVBindFrameFormatNV12, true
	case frame.FormatYUYV:
		return C.AVBindFrameFormatYUYV, true
	case frame.FormatUYVY:
		return C.AVBindFrameFormatUYVY, true
	default:
		return 0, false
	}
}

func frameFormatFromAVBind(f C.AVBindFrameFormat) (frame.Format, bool) {
	switch f {
	case C.AVBindFrameFormatI420:
		return frame.FormatI420, true
	case C.AVBindFrameFormatNV21:
		return frame.FormatNV21, true
	case C.AVBindFrameFormatNV12:
		return frame.FormatNV12, true
	case C.AVBindFrameFormatYUYV:
		return frame.FormatYUYV, true
	case C.AVBindFrameFormatUYVY:
		return frame.FormatUYVY, true
	default:
		return "", false
	}
}

// Devices uses AVFoundation to query a list of devices based on the media type
func Devices(mediaType MediaType) ([]Device, error) {
	var cDevicesPtr C.PAVBindDevice
	var cDevicesLen C.int

	status := C.AVBindDevices(C.AVBindMediaType(mediaType), &cDevicesPtr, &cDevicesLen)
	if status != nil {
		return nil, fmt.Errorf("%s", C.GoString(status))
	}

	// https://github.com/golang/go/wiki/cgo#turning-c-arrays-into-go-slices
	cDevices := (*[1 << 28]C.AVBindDevice)(unsafe.Pointer(cDevicesPtr))[:cDevicesLen:cDevicesLen]
	devices := make([]Device, cDevicesLen)

	for i := range devices {
		devices[i].UID = C.GoString(&cDevices[i].uid[0])
		devices[i].cDevice = cDevices[i]
		devices[i].Name = C.GoString(&cDevices[i].name[0])
	}

	return devices, nil
}

// ReadCloser is a wrapper around the data callback from AVFoundation. The data received from the
// the underlying callback can be retrieved by calling Read.
type ReadCloser struct {
	dataChan   chan []byte
	id         handleID
	onClose    func()
	cancelCtx  context.Context
	cancelFunc func()
	closeWG    sync.WaitGroup
	lock       sync.Mutex
}

func newReadCloser(onClose func()) *ReadCloser {
	var rc ReadCloser
	rc.dataChan = make(chan []byte, 1)
	rc.onClose = onClose
	rc.id = register(rc.dataCb)
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	rc.cancelCtx = cancelCtx
	rc.cancelFunc = cancelFunc
	return &rc
}

func (rc *ReadCloser) dataCb(data []byte) {
	rc.closeWG.Add(1)
	defer rc.closeWG.Done()

	// TODO: add a policy for slow reader
	if rc.cancelCtx.Err() != nil {
		return
	}
	select {
	// Use the Done channel to avoid waiting for new data from closed camera
	case <-rc.cancelCtx.Done():
	case rc.dataChan <- data:
	}
}

// Read reads raw data, the format is determined by the media type and property:
//   - For video, each call will return a frame.
//   - For audio, each call will return a chunk which its size configured by Latency
func (rc *ReadCloser) Read() ([]byte, func(), error) {
	data, ok := <-rc.dataChan
	if !ok {
		return nil, func() {}, io.EOF
	}
	return data, func() {}, nil
}

// Close closes the capturing session, and no data will flow anymore
func (rc *ReadCloser) Close() {
	rc.lock.Lock()
	defer rc.lock.Unlock()

	if rc.cancelCtx.Err() != nil {
		return // already closed
	}

	if rc.onClose != nil {
		rc.onClose()
	}
	rc.cancelFunc()
	unregister(rc.id)
	rc.closeWG.Wait()
	close(rc.dataChan)
}

// Session represents a capturing session.
type Session struct {
	device   Device
	cSession C.PAVBindSession
	lock     sync.Mutex
	closed   bool
}

// NewSession creates a new capturing session
func NewSession(device Device) (*Session, error) {
	var session Session

	status := C.AVBindSessionInit(device.cDevice, &session.cSession)
	if status != nil {
		return nil, fmt.Errorf("%s", C.GoString(status))
	}

	session.device = device
	return &session, nil
}

// Close stops capturing session and frees up resources
func (session *Session) Close() error {
	session.lock.Lock()
	defer session.lock.Unlock()
	if session.closed {
		return nil
	}
	session.closed = true

	if session.cSession == nil {
		return nil
	}

	status := C.AVBindSessionFree(&session.cSession)
	if status != nil {
		return fmt.Errorf("%s", C.GoString(status))
	}
	return nil
}

// Open start capturing session. As soon as it returns successfully, the data will start
// flowing. The raw data can be retrieved by using ReadCloser's Read method.
func (session *Session) Open(property prop.Media) (*ReadCloser, error) {
	frameFormat, ok := frameFormatToAVBind(property.FrameFormat)
	if !ok {
		return nil, fmt.Errorf("Unsupported frame format")
	}

	cProperty := C.AVBindMediaProperty{
		width:       C.int(property.Width),
		height:      C.int(property.Height),
		frameFormat: frameFormat,
	}

	rc := newReadCloser(func() {
		C.AVBindSessionClose(session.cSession)
	})
	status := C.AVBindSessionOpen(
		session.cSession,
		cProperty,
		C.AVBindDataCallback(unsafe.Pointer(C.onDataBridge)),
		unsafe.Pointer(&rc.id),
	)
	if status != nil {
		return nil, fmt.Errorf("%s", C.GoString(status))
	}
	return rc, nil
}

// Properties queries a list of properties that device supports
func (session *Session) Properties() []prop.Media {
	var cPropertiesPtr C.PAVBindMediaProperty
	var cPropertiesLen C.int

	status := C.AVBindSessionProperties(session.cSession, &cPropertiesPtr, &cPropertiesLen)
	if status != nil {
		return nil
	}

	// https://github.com/golang/go/wiki/cgo#turning-c-arrays-into-go-slices
	cProperties := (*[1 << 28]C.AVBindMediaProperty)(unsafe.Pointer(cPropertiesPtr))[:cPropertiesLen:cPropertiesLen]
	var properties []prop.Media
	for _, cProperty := range cProperties {
		frameFormat, ok := frameFormatFromAVBind(cProperty.frameFormat)
		if ok {
			properties = append(properties, prop.Media{
				Video: prop.Video{
					Width:       int(cProperty.width),
					Height:      int(cProperty.height),
					FrameFormat: frameFormat,
				},
			})
		}
	}
	return properties
}
