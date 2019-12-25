package driver

import "fmt"

type manager struct {
	drivers map[string]Driver
}

// Manager is a singleton to manage multiple drivers and their states
var Manager = &manager{
	drivers: make(map[string]Driver),
}

func (m *manager) register(a Adapter) error {
	d := wrapAdapter(a)
	if d == nil {
		return fmt.Errorf("adapter has to be either VideoAdapter/AudioAdapter")
	}

	m.drivers[d.ID()] = d
	return nil
}

func (m *manager) Query() []Driver {
	results := make([]Driver, 0)
	for _, d := range m.drivers {
		results = append(results, d)
	}

	return results
}
