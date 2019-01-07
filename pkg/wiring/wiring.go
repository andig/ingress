package wiring

import (
	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/log"
)

// Wire connects source and target with associated mapping
type Wire struct {
	Source   string
	Target   string
	Mappings [][]Mapping
	Actions  []api.Action
}

// Wiring is a list of wires
type Wiring struct {
	wires []Wire
}

// NewWiring creates a system wiring, validatated against available connectors
func NewWiring(c []config.Wire, conn *Connectors, mappings *Mappings, actions *Actions) *Wiring {
	wires := make([]Wire, 0)
	for _, wire := range c {
		for _, source := range wire.Sources {
			if _, err := conn.SourceForName(source); err != nil {
				log.Fatalf("cannot wire %s -> *, source not defined", source)
			}

			for _, target := range wire.Targets {
				if _, err := conn.TargetForName(target); err != nil {
					log.Fatalf("cannot wire %s -> %s, target not defined", source, target)
				}

				wireMappings := make([][]Mapping, 0)
				for _, mapping := range wire.Mappings {
					wireMapping, err := mappings.MappingsForName(mapping)
					if err != nil {
						log.Fatalf("cannot wire %s -> %s, undefined mapping %s", source, target, mapping)
					}

					wireMappings = append(wireMappings, wireMapping)
				}

				log.Context(
					log.SRC, source,
					log.TGT, target,
				).Printf("creating wire")

				wire := Wire{
					Source:   source,
					Target:   target,
					Mappings: wireMappings,
				}
				wires = append(wires, wire)
			}
		}
	}

	if len(wires) == 0 {
		log.Println("no wires created - please check your configuration")
	}

	wiring := &Wiring{
		wires: wires,
	}
	return wiring
}

// WiresForSource returns all wires connected to given source
func (w *Wiring) WiresForSource(source string) []Wire {
	res := make([]Wire, 0)

	for _, wire := range w.wires {
		if wire.Source == source {
			res = append(res, wire)
		}
	}

	return res
}
