package driver

import (
	"fmt"
	"github.com/pion/mediadevices/pkg/io/video"
	uuid "github.com/satori/go.uuid"
)

func wrapAdapter(a Adapter) Driver {
	var d Driver
	id := uuid.NewV4().String()

	switch v := a.(type) {
	case VideoAdapter:
		d = &videoAdapterWrapper{
			VideoAdapter: v,
			id:           id,
		}
	case AudioAdapter:
		d = &audioAdapterWrapper{
			AudioAdapter: v,
			id:           id,
		}
	}

	return d
}

// TODO: Add state validation
type videoAdapterWrapper struct {
	VideoAdapter
	id    string
	state State
}

func (w *videoAdapterWrapper) ID() string {
	return w.id
}

func (w *videoAdapterWrapper) Status() State {
	return w.state
}

func (w *videoAdapterWrapper) Open() error {
	if w.state != StateClosed {
		return fmt.Errorf("invalid state: driver is already opened")
	}

	err := w.VideoAdapter.Open()
	if err == nil {
		w.state = StateOpened
	}
	return err
}

func (w *videoAdapterWrapper) Close() error {
	err := w.VideoAdapter.Close()
	if err == nil {
		w.state = StateClosed
	}
	return err
}

func (w *videoAdapterWrapper) Start(setting VideoSetting) (video.Reader, error) {
	if w.state == StateClosed {
		return nil, fmt.Errorf("invalid state: driver hasn't been opened")
	}

	if w.state == StateStarted {
		return nil, fmt.Errorf("invalid state: driver has been started")
	}

	r, err := w.VideoAdapter.Start(setting)
	if err == nil {
		w.state = StateStarted
	}
	return r, err
}

func (w *videoAdapterWrapper) Stop() error {
	if w.state != StateStarted {
		return fmt.Errorf("invalid state: driver hasn't been started")
	}

	err := w.VideoAdapter.Stop()
	if err == nil {
		w.state = StateStopped
	}
	return err
}

func (w *videoAdapterWrapper) Settings() []VideoSetting {
	if w.state == StateClosed {
		return nil
	}

	return w.VideoAdapter.Settings()
}

// TODO: Add state validation
type audioAdapterWrapper struct {
	AudioAdapter
	id    string
	state State
}

func (w *audioAdapterWrapper) ID() string {
	return w.id
}

func (w *audioAdapterWrapper) Status() State {
	return w.state
}
