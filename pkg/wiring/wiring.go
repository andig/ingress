package wiring

import (
	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/log"
)

// Wire connects source and target with associated mapping
type Wire struct {
	Source  string
	Target  string
	Actions []api.Action
}

// Wires is a list of wires
type Wires struct {
	wires []Wire
}

// NewWiring creates a system wiring, validatated against available connectors
func NewWiring(c []config.Wire, conn *Connectors, actions *Actions) *Wires {
	wires := &Wires{
		wires: make([]Wire, 0),
	}
	for _, wire := range c {
		if wire.Source == "" {
			log.Fatalf("configuration error: missing wire source (%+v)", wire)
		}
		if _, err := conn.SourceForName(wire.Source); err != nil {
			log.Fatalf("cannot wire source %s -> *, source not defined", wire.Source)
		}

		if wire.Target == "" {
			log.Fatalf("configuration error: missing wire target (%+v)", wire)
		}
		if _, err := conn.TargetForName(wire.Target); err != nil {
			log.Fatalf("cannot wire target * -> %s, target not defined", wire.Target)
		}

		wireActions := make([]api.Action, 0)
		for _, action := range wire.Actions {
			wireAction, err := actions.ActionForName(action)
			if err != nil {
				log.Fatalf("cannot wire %s -> %s with action %s, action not defined", wire.Source, wire.Target, action)
			}

			wireActions = append(wireActions, wireAction)
		}

		log.Context(
			log.SRC, wire.Source,
			log.TGT, wire.Target,
		).Printf("creating wire")

		w := Wire{
			Source:  wire.Source,
			Target:  wire.Target,
			Actions: wireActions,
		}
		wires.wires = append(wires.wires, w)
	}

	if len(wires.wires) == 0 {
		log.Fatal("no wires created - please check your configuration")
	}

	return wires
}

// WiresForSource returns all wires connected to given source
func (w *Wires) WiresForSource(source string) []Wire {
	res := make([]Wire, 0)

	for _, wire := range w.wires {
		if wire.Source == source {
			res = append(res, wire)
		}
	}

	return res
}
