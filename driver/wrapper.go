package driver

import uuid "github.com/satori/go.uuid"

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
