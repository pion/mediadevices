package mediadevices

import (
	"errors"
	"github.com/pion/interceptor"
	"io"
	"testing"
	"time"
)

func TestOnEnded(t *testing.T) {
	errExpected := errors.New("an error")

	t.Run("ErrorAfterRegister", func(t *testing.T) {
		tr := &baseTrack{}

		called := make(chan error, 1)
		tr.OnEnded(func(error) {
			called <- errExpected
		})
		select {
		case <-called:
			t.Error("OnEnded handler is unexpectedly called")
		case <-time.After(10 * time.Millisecond):
		}

		tr.onError(errExpected)

		select {
		case err := <-called:
			if err != errExpected {
				t.Errorf("Expected to receive error: %v, got: %v", errExpected, err)
			}
		case <-time.After(10 * time.Millisecond):
			t.Error("Timeout")
		}
	})

	t.Run("ErrorBeforeRegister", func(t *testing.T) {
		tr := &baseTrack{}

		tr.onError(errExpected)

		called := make(chan error, 1)
		tr.OnEnded(func(err error) {
			called <- errExpected
		})
		select {
		case err := <-called:
			if err != errExpected {
				t.Errorf("Expected to receive error: %v, got: %v", errExpected, err)
			}
		case <-time.After(10 * time.Millisecond):
			t.Error("Timeout")
		}
	})
}

type fakeRTCPReader struct {
	mockReturn chan []byte
	end        chan struct{}
}

func (mock *fakeRTCPReader) Read(buffer []byte, attributes interceptor.Attributes) (int, interceptor.Attributes, error) {
	select {
	case <-mock.end:
		return 0, nil, io.EOF
	case mockReturn := <-mock.mockReturn:
		if len(buffer) < len(mock.mockReturn) {
			return 0, nil, io.ErrShortBuffer
		}

		return copy(buffer, mockReturn), attributes, nil
	}
}

type fakeKeyFrameController struct {
	called chan struct{}
}

func (mock *fakeKeyFrameController) ForceKeyFrame() error {
	mock.called <- struct{}{}
	return nil
}

func TestRtcpHandler(t *testing.T) {

	t.Run("ShouldStopReading", func(t *testing.T) {
		tr := &baseTrack{}
		stop := make(chan struct{}, 1)
		stopped := make(chan struct{})
		go func() {
			tr.rtcpReadLoop(&fakeRTCPReader{end: stop}, &fakeKeyFrameController{}, stop)
			stopped <- struct{}{}
		}()

		stop <- struct{}{}

		select {
		case <-time.After(100 * time.Millisecond):
			t.Error("Timeout")
		case <-stopped:
		}
	})

	t.Run("ShouldForceKeyFrame", func(t *testing.T) {
		for packetType, packet := range map[string][]byte{
			"PLI": {
				// v=2, p=0, FMT=1, PSFB, len=1
				0x81, 0xce, 0x00, 0x02,
				// ssrc=0x0
				0x00, 0x00, 0x00, 0x00,
				// ssrc=0x4bc4fcb4
				0x4b, 0xc4, 0xfc, 0xb4,
			},
			"FIR": {
				// v=2, p=0, FMT=4, PSFB, len=3
				0x84, 0xce, 0x00, 0x04,
				// ssrc=0x0
				0x00, 0x00, 0x00, 0x00,
				// ssrc=0x4bc4fcb4
				0x4b, 0xc4, 0xfc, 0xb4,
				// ssrc=0x12345678
				0x12, 0x34, 0x56, 0x78,
				// Seqno=0x42
				0x42, 0x00, 0x00, 0x00,
			},
		} {
			t.Run(packetType, func(t *testing.T) {
				tr := &baseTrack{}
				tr.OnEnded(func(err error) {
					if err != io.EOF {
						t.Error(err)
					}
				})
				stop := make(chan struct{}, 1)
				defer func() {
					stop <- struct{}{}
				}()
				mockKeyFrameController := &fakeKeyFrameController{called: make(chan struct{}, 1)}
				mockRTCPReader := &fakeRTCPReader{end: stop, mockReturn: make(chan []byte, 1)}

				go tr.rtcpReadLoop(mockRTCPReader, mockKeyFrameController, stop)

				mockRTCPReader.mockReturn <- packet

				select {
				case <-time.After(1000 * time.Millisecond):
					t.Error("Timeout")
				case <-mockKeyFrameController.called:
				}
			})
		}
	})
}
