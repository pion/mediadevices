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

// FilterID return a filter function to get registered drivers which have given ID
func FilterID(id string) FilterFn {
	return func(d Driver) bool {
		return d.ID() == id
	}
}

// FilterDeviceType returns a filter function to get registered drivers which matches t type
func FilterDeviceType(t DeviceType) FilterFn {
	return func(d Driver) bool {
		return d.Info().DeviceType == t
	}
}

// FilterAnd returns a filter function to take logical conjunction of given filters.
func FilterAnd(filters ...FilterFn) FilterFn {
	return func(d Driver) bool {
		for _, f := range filters {
			if !f(d) {
				return false
			}
		}
		return true
	}
}

// FilterNot returns a filter function to take logical inverse of the given filter.
func FilterNot(filter FilterFn) FilterFn {
	return func(d Driver) bool {
		return !filter(d)
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
func (m *Manager) Register(a Adapter, info Info) error {
	d := wrapAdapter(a, info)
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
