package actions

import (
	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/log"
	"github.com/andig/ingress/pkg/registry"
)

func init() {
	registry.RegisterAction("map", NewMapping)
}

type MappingRemainder string

const (
	pass MappingRemainder = "pass"
)

type mappingConfig struct {
	config.Action `yaml:",squash"`
	Matches       map[string]string
	Other         MappingRemainder
}

// Mappings is a list of mappings identified by mapping name
type Mapping struct {
	mappingConfig
}

// NewMapping creates a mapping action
func NewMapping(g config.Generic) (a api.Action, err error) {
	var conf mappingConfig
	err = config.Decode(g, &conf)
	if err != nil {
		return nil, err
	}

	a = &Mapping{
		conf,
	}
	return a, nil
}

// Process implements the mapping's Action interface
func (a *Mapping) Process(d api.Data) api.Data {
	dataName := d.Name()
	for from, to := range a.Matches {
		if dataName == from {
			log.Context(
				log.EV, d.Name(),
				log.ACT, a.Name,
			).Debugf("mapping %s -> %s ", d.Name(), to)

			d.SetName(to)
			return d
		}
	}

	// not mapped
	return a.remainder(d)
}

// remainder implements the mapping actions "Other" config setting behaviour
func (a *Mapping) remainder(d api.Data) api.Data {
	if a.Other == pass {
		return d
	}

	log.Context(
		log.EV, d.Name(),
		log.ACT, a.Name,
	).Debugf("dropped")

	return nil
}
