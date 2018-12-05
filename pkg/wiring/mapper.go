package wiring

import (
	"log"

	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/data"
)

type MappingMap map[string]config.Mapping

type Mapper struct {
	mappings  MappingMap
	publisher PublisherMap
}

func NewMapper(c []config.Mapping, publisher PublisherMap) *Mapper {
	mappings := make(MappingMap)
	for _, mapping := range c {
		mappings[mapping.Input.Name] = mapping
	}

	mapper := &Mapper{
		mappings:  mappings,
		publisher: publisher,
	}
	return mapper
}

func (m *Mapper) Process(d *data.InputData) {
	// TODO allow source multiple times
	mapping, ok := m.mappings[d.Source]
	if !ok {
		log.Println("mapper: invalid source " + d.Source)
		return
	}

	output, ok := m.publisher[mapping.Output.Name]
	if !ok {
		log.Println("mapper: invalid target " + mapping.Output.Name)
		return
	}

	output.Publish(*d.Data)
}
