package driver

import "fmt"

// FilterFn is being used to decide if a driver should be included in the
// query result.
type FilterFn func(Driver) bool

// Manager is a singleton to manage multiple drivers and their states
type Manager struct {
	drivers map[string]Driver
}

var manager = &Manager{
	drivers: make(map[string]Driver),
}

// GetManager gets manager singleton instance
func GetManager() *Manager {
	return manager
}

// Register registers adapter to be discoverable by Query
func (m *Manager) Register(a Adapter) error {
	d := wrapAdapter(a)
	if d == nil {
		return fmt.Errorf("adapter has to be either VideoAdapter/AudioAdapter")
	}

	m.drivers[d.ID()] = d
	return nil
}

// Query queries by using f to filter drivers, and simply return the filtered results.
func (m *Manager) Query(f FilterFn) []Driver {
	results := make([]Driver, 0)
	for _, d := range m.drivers {
		if ok := f(d); ok {
			results = append(results, d)
		}
	}

	return results
}

// VideoDrivers gets a list of registered VideoDriver
func (m *Manager) VideoDrivers() []Driver {
	return m.Query(func(d Driver) bool {
		_, ok := d.(VideoDriver)
		return ok
	})
}

// AudioDrivers gets a list of registered AudioDriver
func (m *Manager) AudioDrivers() []Driver {
	return m.Query(func(d Driver) bool {
		_, ok := d.(AudioDriver)
		return ok
	})
}
