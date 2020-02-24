package microphone

import (
	"errors"
	"golang.org/x/sys/windows"
	"io"
	"time"
	"unsafe"

	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/prop"
)

const (
	// bufferNumber * prop.Audio.Latency is the maximum blockable duration
	// to get data without dropping chunks.
	bufferNumber = 5
)

// Windows APIs
var (
	winmm                 = windows.NewLazySystemDLL("Winmm.dll")
	waveInOpen            = winmm.NewProc("waveInOpen")
	waveInStart           = winmm.NewProc("waveInStart")
	waveInStop            = winmm.NewProc("waveInStop")
	waveInReset           = winmm.NewProc("waveInReset")
	waveInClose           = winmm.NewProc("waveInClose")
	waveInPrepareHeader   = winmm.NewProc("waveInPrepareHeader")
	waveInAddBuffer       = winmm.NewProc("waveInAddBuffer")
	waveInUnprepareHeader = winmm.NewProc("waveInUnprepareHeader")
)

type buffer struct {
	waveHdr
	data []int16
}

func newBuffer(samples int) *buffer {
	b := make([]int16, samples)
	return &buffer{
		waveHdr: waveHdr{
			// Sharing Go memory with Windows C API without reference.
			// Make sure that the lifetime of the buffer struct is longer
			// than the final access from cbWaveIn.
			lpData:         uintptr(unsafe.Pointer(&b[0])),
			dwBufferLength: uint32(samples * 2),
		},
		data: b,
	}
}

type microphone struct {
	hWaveIn windows.Pointer
	buf     map[uintptr]*buffer
	chBuf   chan *buffer
	closed  chan struct{}
}

func init() {
	// TODO: enum devices
	driver.GetManager().Register(&microphone{}, driver.Info{
		Label:      "default",
		DeviceType: driver.Microphone,
	})
}

func (m *microphone) Open() error {
	m.chBuf = make(chan *buffer)
	m.buf = make(map[uintptr]*buffer)
	m.closed = make(chan struct{})
	return nil
}

func (m *microphone) cbWaveIn(hWaveIn windows.Pointer, uMsg uint, dwInstance, dwParam1, dwParam2 *int32) uintptr {
	switch uMsg {
	case MM_WIM_DATA:
		b := m.buf[uintptr(unsafe.Pointer(dwParam1))]
		m.chBuf <- b

	case MM_WIM_OPEN:
	case MM_WIM_CLOSE:
		close(m.chBuf)
	}
	return 0
}

func (m *microphone) Close() error {
	if m.hWaveIn == nil {
		return nil
	}
	close(m.closed)

	ret, _, _ := waveInStop.Call(
		uintptr(unsafe.Pointer(m.hWaveIn)),
	)
	if err := errWinmm[ret]; err != nil {
		return err
	}
	// All enqueued buffers are marked done by waveInReset.
	ret, _, _ = waveInReset.Call(
		uintptr(unsafe.Pointer(m.hWaveIn)),
	)
	if err := errWinmm[ret]; err != nil {
		return err
	}
	for _, buf := range m.buf {
		// Detach buffers from waveIn API.
		ret, _, _ := waveInUnprepareHeader.Call(
			uintptr(unsafe.Pointer(m.hWaveIn)),
			uintptr(unsafe.Pointer(&buf.waveHdr)),
			uintptr(unsafe.Sizeof(buf.waveHdr)),
		)
		if err := errWinmm[ret]; err != nil {
			return err
		}
	}
	// Now, it's ready to free the buffers.
	// As microphone struct still has reference to the buffers,
	// they will be GC-ed once microphone is reopened or unreferenced.

	ret, _, _ = waveInClose.Call(
		uintptr(unsafe.Pointer(m.hWaveIn)),
	)
	if err := errWinmm[ret]; err != nil {
		return err
	}
	<-m.chBuf
	m.hWaveIn = nil

	return nil
}

func (m *microphone) AudioRecord(p prop.Media) (audio.Reader, error) {
	for i := 0; i < bufferNumber; i++ {
		b := newBuffer(
			int(uint64(p.Latency) * uint64(p.SampleRate) / uint64(time.Second)),
		)
		// Map the buffer by its data head address to restore access to the Go struct
		// in callback function. Don't resize the buffer after it.
		m.buf[uintptr(unsafe.Pointer(&b.waveHdr))] = b
	}

	waveFmt := &waveFormatEx{
		wFormatTag:      WAVE_FORMAT_PCM,
		nChannels:       uint16(p.ChannelCount),
		nSamplesPerSec:  uint32(p.SampleRate),
		nAvgBytesPerSec: uint32(p.SampleRate * p.ChannelCount * 2),
		nBlockAlign:     uint16(p.ChannelCount * 2),
		wBitsPerSample:  16,
	}
	ret, _, _ := waveInOpen.Call(
		uintptr(unsafe.Pointer(&m.hWaveIn)),
		WAVE_MAPPER,
		uintptr(unsafe.Pointer(waveFmt)),
		windows.NewCallback(m.cbWaveIn),
		0,
		CALLBACK_FUNCTION,
	)
	if err := errWinmm[ret]; err != nil {
		return nil, err
	}

	for _, buf := range m.buf {
		// Attach buffers to waveIn API.
		ret, _, _ := waveInPrepareHeader.Call(
			uintptr(unsafe.Pointer(m.hWaveIn)),
			uintptr(unsafe.Pointer(&buf.waveHdr)),
			uintptr(unsafe.Sizeof(buf.waveHdr)),
		)
		if err := errWinmm[ret]; err != nil {
			return nil, err
		}
	}
	for _, buf := range m.buf {
		// Enqueue buffers.
		ret, _, _ := waveInAddBuffer.Call(
			uintptr(unsafe.Pointer(m.hWaveIn)),
			uintptr(unsafe.Pointer(&buf.waveHdr)),
			uintptr(unsafe.Sizeof(buf.waveHdr)),
		)
		if err := errWinmm[ret]; err != nil {
			return nil, err
		}
	}

	ret, _, _ = waveInStart.Call(
		uintptr(unsafe.Pointer(m.hWaveIn)),
	)
	if err := errWinmm[ret]; err != nil {
		return nil, err
	}

	// TODO: detect microphone device disconnection and return EOF

	var bi int
	reader := audio.ReaderFunc(func(samples [][2]float32) (n int, err error) {
		var b *buffer
		for i := range samples {
			// if we don't have anything left in buff, we'll wait until we receive
			// more samples
			if b == nil || bi == int(b.waveHdr.dwBytesRecorded/2) {
				var more bool
				b, more = <-m.chBuf
				if !more {
					return i, io.EOF
				}

				select {
				case <-m.closed:
				default:
					// Re-enqueue used buffer.
					ret, _, _ := waveInAddBuffer.Call(
						uintptr(unsafe.Pointer(m.hWaveIn)),
						uintptr(unsafe.Pointer(&b.waveHdr)),
						uintptr(unsafe.Sizeof(b.waveHdr)),
					)
					if err := errWinmm[ret]; err != nil {
						return 0, err
					}
				}

				bi = 0
			}

			samples[i][0] = float32(b.data[bi]) / 0x8000
			if p.ChannelCount == 2 {
				samples[i][1] = float32(b.data[bi+1]) / 0x8000
				bi++
			}
			bi++
		}
		return len(samples), nil
	})
	return reader, nil
}

func (m *microphone) Properties() []prop.Media {
	// TODO: Get actual properties
	monoProp := prop.Media{
		Audio: prop.Audio{
			SampleRate:   48000,
			Latency:      time.Millisecond * 20,
			ChannelCount: 1,
		},
	}

	stereoProp := monoProp
	stereoProp.ChannelCount = 2

	return []prop.Media{monoProp, stereoProp}
}

// Windows API structures

type waveFormatEx struct {
	wFormatTag      uint16
	nChannels       uint16
	nSamplesPerSec  uint32
	nAvgBytesPerSec uint32
	nBlockAlign     uint16
	wBitsPerSample  uint16
	cbSize          uint16
}

type waveHdr struct {
	lpData          uintptr
	dwBufferLength  uint32
	dwBytesRecorded uint32
	dwUser          *uint32
	dwFlags         uint32
	dwLoops         uint32
	lpNext          *waveHdr
	reserved        *uint32
}

// Windows consts

const (
	MMSYSERR_NOERROR      = 0
	MMSYSERR_ERROR        = 1
	MMSYSERR_BADDEVICEID  = 2
	MMSYSERR_NOTENABLED   = 3
	MMSYSERR_ALLOCATED    = 4
	MMSYSERR_INVALHANDLE  = 5
	MMSYSERR_NODRIVER     = 6
	MMSYSERR_NOMEM        = 7
	MMSYSERR_NOTSUPPORTED = 8
	MMSYSERR_BADERRNUM    = 9
	MMSYSERR_INVALFLAG    = 10
	MMSYSERR_INVALPARAM   = 11
	MMSYSERR_HANDLEBUSY   = 12
	MMSYSERR_INVALIDALIAS = 13
	MMSYSERR_BADDB        = 14
	MMSYSERR_KEYNOTFOUND  = 15
	MMSYSERR_READERROR    = 16
	MMSYSERR_WRITEERROR   = 17
	MMSYSERR_DELETEERROR  = 18
	MMSYSERR_VALNOTFOUND  = 19
	MMSYSERR_NODRIVERCB   = 20

	WAVERR_BADFORMAT    = 32
	WAVERR_STILLPLAYING = 33
	WAVERR_UNPREPARED   = 34
	WAVERR_SYNC         = 35

	WAVE_MAPPER     = 0xFFFF
	WAVE_FORMAT_PCM = 1

	CALLBACK_NULL     = 0
	CALLBACK_WINDOW   = 0x10000
	CALLBACK_TASK     = 0x20000
	CALLBACK_FUNCTION = 0x30000
	CALLBACK_THREAD   = CALLBACK_TASK
	CALLBACK_EVENT    = 0x50000

	MM_WIM_OPEN  = 0x3BE
	MM_WIM_CLOSE = 0x3BF
	MM_WIM_DATA  = 0x3C0
)

var errWinmm = map[uintptr]error{
	MMSYSERR_NOERROR:      nil,
	MMSYSERR_ERROR:        errors.New("error"),
	MMSYSERR_BADDEVICEID:  errors.New("bad device id"),
	MMSYSERR_NOTENABLED:   errors.New("not enabled"),
	MMSYSERR_ALLOCATED:    errors.New("already allocated"),
	MMSYSERR_INVALHANDLE:  errors.New("invalid handler"),
	MMSYSERR_NODRIVER:     errors.New("no driver"),
	MMSYSERR_NOMEM:        errors.New("no memory"),
	MMSYSERR_NOTSUPPORTED: errors.New("not supported"),
	MMSYSERR_BADERRNUM:    errors.New("band error number"),
	MMSYSERR_INVALFLAG:    errors.New("invalid flag"),
	MMSYSERR_INVALPARAM:   errors.New("invalid param"),
	MMSYSERR_HANDLEBUSY:   errors.New("handle busy"),
	MMSYSERR_INVALIDALIAS: errors.New("invalid alias"),
	MMSYSERR_BADDB:        errors.New("bad db"),
	MMSYSERR_KEYNOTFOUND:  errors.New("key not found"),
	MMSYSERR_READERROR:    errors.New("read error"),
	MMSYSERR_WRITEERROR:   errors.New("write error"),
	MMSYSERR_DELETEERROR:  errors.New("delete error"),
	MMSYSERR_VALNOTFOUND:  errors.New("value not found"),
	MMSYSERR_NODRIVERCB:   errors.New("no driver cb"),
	WAVERR_BADFORMAT:      errors.New("bad format"),
	WAVERR_STILLPLAYING:   errors.New("still playing"),
	WAVERR_UNPREPARED:     errors.New("unprepared"),
	WAVERR_SYNC:           errors.New("sync"),
}
