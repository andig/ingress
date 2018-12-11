package wiring

import (
	"log"

	"github.com/andig/ingress/pkg/config"
)

type Wire struct {
	Source   string
	Target   string
	Mappings [][]Mapping
}

type Wiring struct {
	wires []Wire
}

// NewWiring creates a system wiring, validatated against available connectors
func NewWiring(c []config.Wiring, mappings *Mappings, conn *Connectors) *Wiring {
	wires := make([]Wire, 0)
	for _, wiring := range c {
		for _, source := range wiring.Sources {
			if _, err := conn.SourceForName(source); err != nil {
				log.Fatalf("wiring: cannot wire %s -> *, source not defined", source)
			}

			for _, target := range wiring.Targets {
				if _, err := conn.TargetForName(target); err != nil {
					log.Fatalf("wiring: cannot wire %s -> %s, target not defined", source, target)
				}

				wireMappings := make([][]Mapping, 0)
				for _, mapping := range wiring.Mappings {
					wireMapping, err := mappings.MappingsForName(mapping)
					if err != nil {
						log.Fatalf("wiring: cannot wire %s -> %s, undefined mapping %s", source, target, mapping)
					}

					wireMappings = append(wireMappings, wireMapping)
				}

				log.Printf("wiring: wiring %s -> %s ", source, target)

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
		log.Println("wiring: no wires created - please check your configuration")
	}

	wiring := &Wiring{
		wires: wires,
	}
	return wiring
}

func (w *Wiring) WiresForSource(source string) []Wire {
	res := make([]Wire, 0)

	for _, wire := range w.wires {
		if wire.Source == source {
			res = append(res, wire)
		}
	}

	return res
}
