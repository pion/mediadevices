package driver

// FilterFn is being used to decide if a driver should be included in the
// query result.
type FilterFn func(Driver) bool

// FilterVideoRecorder return a filter function to get a list of registered VideoRecorders
func FilterVideoRecorder() FilterFn {
	return func(d Driver) bool {
		_, ok := d.(VideoRecorder)
		return ok
	}
}

// FilterAudioRecorder return a filter function to get a list of registered AudioRecorders
func FilterAudioRecorder() FilterFn {
	return func(d Driver) bool {
		_, ok := d.(AudioRecorder)
		return ok
	}
}

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
func (m *Manager) Register(a Adapter, label string) error {
	d := wrapAdapter(a, label)
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
