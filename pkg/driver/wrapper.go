package driver

import (
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
	uuid "github.com/satori/go.uuid"
)

func wrapAdapter(a Adapter) Driver {
	var d Driver
	id := uuid.NewV4().String()
	wrapper := adapterWrapper{Adapter: a, id: id}

	switch v := a.(type) {
	case VideoCapable:
		d = &videoAdapterWrapper{
			adapterWrapper: &wrapper,
			VideoCapable:   v,
		}
	case AudioCapable:
		d = &audioAdapterWrapper{
			adapterWrapper: &wrapper,
			AudioCapable:   v,
		}
	}

	return d
}

type adapterWrapper struct {
	Adapter
	id    string
	state State
}

func (w *adapterWrapper) ID() string {
	return w.id
}

func (w *adapterWrapper) Status() State {
	return w.state
}

func (w *adapterWrapper) Open() error {
	return w.state.Update(StateOpened, w.Adapter.Open)
}

func (w *adapterWrapper) Close() error {
	return w.state.Update(StateClosed, w.Adapter.Close)
}

// TODO: Add state validation
type videoAdapterWrapper struct {
	*adapterWrapper
	VideoCapable
}

func (w *videoAdapterWrapper) Start(prop video.AdvancedProperty) (r video.Reader, err error) {
	w.state.Update(StateStarted, func() error {
		r, err = w.VideoCapable.Start(prop)
		return err
	})
	return
}

func (w *videoAdapterWrapper) Stop() error {
	return w.state.Update(StateStopped, w.VideoCapable.Stop)
}

func (w *videoAdapterWrapper) Properties() []video.AdvancedProperty {
	if w.state == StateClosed {
		return nil
	}

	return w.VideoCapable.Properties()
}

type audioAdapterWrapper struct {
	*adapterWrapper
	AudioCapable
}

func (w *audioAdapterWrapper) Start(prop audio.AdvancedProperty) (r audio.Reader, err error) {
	w.state.Update(StateStarted, func() error {
		r, err = w.AudioCapable.Start(prop)
		return err
	})
	return
}

func (w *audioAdapterWrapper) Stop() error {
	return w.state.Update(StateStopped, w.AudioCapable.Stop)
}

func (w *audioAdapterWrapper) Properties() []audio.AdvancedProperty {
	if w.state == StateClosed {
		return nil
	}

	return w.AudioCapable.Properties()
}
