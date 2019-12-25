package driver

import (
	uuid "github.com/satori/go.uuid"
)

type manager struct {
	drivers map[string]Driver
}

// Manager is a singleton to manage multiple drivers and their states
var Manager = &manager{
	drivers: make(map[string]Driver),
}

func (m *manager) register(d Driver) {
	id := uuid.NewV4()
	m.drivers[id.String()] = d
}

func (m *manager) Query() []QueryResult {
	results := make([]QueryResult, 0)
	for id, d := range m.drivers {
		results = append(results, QueryResult{
			ID:     id,
			Driver: d,
		})
	}

	return results
}
