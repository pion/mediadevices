package driver

import (
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/stretchr/testify/assert"
	"testing"
)

func filterTrue(_ Driver) bool {
	return true
}
func filterFalse(_ Driver) bool {
	return false
}

func TestFilterNot(t *testing.T) {
	if FilterNot(filterTrue)(nil) != false {
		t.Error("FilterNot(filterTrue)() must be false")
	}
	if FilterNot(filterFalse)(nil) != true {
		t.Error("FilterNot(filterFalse)() must be true")
	}
}

func TestFilterAnd(t *testing.T) {
	if FilterAnd(filterTrue, filterTrue)(nil) != true {
		t.Error("FilterAnd(filterTrue, filterTrue)() must be true")
	}
	if FilterAnd(filterTrue, filterFalse)(nil) != false {
		t.Error("FilterAnd(filterTrue, filterFalse)() must be false")
	}
	if FilterAnd(filterFalse, filterTrue)(nil) != false {
		t.Error("FilterAnd(filterFalse, filterTrue)() must be false")
	}
	if FilterAnd(filterFalse, filterFalse)(nil) != false {
		t.Error("FilterAnd(filterFalse, filterFalse)() must be false")
	}
	if FilterAnd(filterFalse, filterTrue, filterTrue)(nil) != false {
		t.Error("FilterAnd(filterFalse, filterTrue, filterTrue)() must be false")
	}
	if FilterAnd(filterTrue, filterTrue, filterTrue)(nil) != true {
		t.Error("FilterAnd(filterTrue, filterTrue, filterTrue)() must be true")
	}
}

type fakeVideoAdapter struct{}

func (a *fakeVideoAdapter) Open() error              { return nil }
func (a *fakeVideoAdapter) Close() error             { return nil }
func (a *fakeVideoAdapter) Properties() []prop.Media { return nil }

func (a *fakeVideoAdapter) VideoRecord(_ prop.Media) (r video.Reader, err error) { return nil, nil }

type fakeAudioAdapter struct{}

func (a *fakeAudioAdapter) Open() error              { return nil }
func (a *fakeAudioAdapter) Close() error             { return nil }
func (a *fakeAudioAdapter) Properties() []prop.Media { return nil }

func (a *fakeAudioAdapter) AudioRecord(_ prop.Media) (r audio.Reader, err error) { return nil, nil }

type fakeAdapter struct{}

func (a *fakeAdapter) Open() error              { return nil }
func (a *fakeAdapter) Close() error             { return nil }
func (a *fakeAdapter) Properties() []prop.Media { return nil }

func TestRegister(t *testing.T) {
	m := GetManager()

	va := &fakeVideoAdapter{}
	err := m.Register(va, Info{})
	assert.NoError(t, err, "cannot register video adapter")
	assert.Equal(t, len(m.Query(filterTrue)), 1)

	aa := &fakeAudioAdapter{}
	err = m.Register(aa, Info{})
	assert.NoError(t, err, "cannot register audio adapter")
	assert.Equal(t, len(m.Query(filterTrue)), 2)

	a := &fakeAdapter{}
	assert.Panics(t, func() { m.Register(a, Info{}) }, "should not register adapter that is neither audio nor video")
	assert.Equal(t, len(m.Query(filterTrue)), 2)
}

func TestRegisterSync(t *testing.T) {
	m := GetManager()
	start := make(chan struct{})
	race := func() {
		<-start
		assert.NoError(t, m.Register(&fakeVideoAdapter{}, Info{}))
	}

	go race()
	go race()
	close(start)
}

func TestQuerySync(t *testing.T) {
	m := GetManager()
	start := make(chan struct{})
	race := func() {
		<-start
		m.Query(filterTrue)
	}

	go race()
	go race()
	close(start)
	// write while reading
	assert.NoError(t, m.Register(&fakeVideoAdapter{}, Info{}))
}
