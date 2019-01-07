package wiring

import (
	"errors"
	"sync"

	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/log"
)

// Mapping maps data entity from name to name
type Mapping struct {
	From string
	To   string
}

// Mappings is a list of mappings identified by mapping name
type Mappings struct {
	mux      sync.Mutex
	mappings map[string][]Mapping
}

// NewMappings creates a system wiring, validatated against available connectors
func NewMappings(c []config.Mapping, conn *Connectors) *Mappings {
	m := &Mappings{
		mappings: make(map[string][]Mapping, 0),
	}

	for _, mapping := range c {
		m.createMapping(mapping, conn)
	}

	return m
}

func (m *Mappings) createMapping(conf config.Mapping, conn *Connectors) {
	if conf.Name == "" {
		log.Fatal("configuration error - missing mapping name")
	}

	m.mux.Lock()
	defer m.mux.Unlock()

	if _, ok := m.mappings[conf.Name]; ok {
		log.Fatal("configuration error - cannot redefine mapping " + conf.Name)
	}

	mapEntries := make([]Mapping, 0)
	for _, mapEntry := range conf.Entries {
		e := Mapping{
			From: mapEntry.From,
			To:   mapEntry.To,
		}
		mapEntries = append(mapEntries, e)
	}

	m.mappings[conf.Name] = mapEntries
}

// MappingsForName returns a list of mappings identified by mapping name
func (m *Mappings) MappingsForName(name string) ([]Mapping, error) {
	target, ok := m.mappings[name]
	if !ok {
		return nil, errors.New("Undefined mapping " + name)
	}
	return target, nil
}
