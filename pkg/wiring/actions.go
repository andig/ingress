package wiring

import (
	"errors"

	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/log"
	"github.com/andig/ingress/pkg/registry"
)

type Actions struct {
	actions map[string]api.Action
}

func NewActions(c []config.Generic) *Actions {
	a := &Actions{
		actions: make(map[string]api.Action),
	}

	for _, generic := range c {
		var conf config.Action
		err := config.PartialDecode(generic, &conf)
		if err != nil {
			log.Context(conf.Name).Fatal(err)
		}

		provider, ok := registry.ActionProviders[conf.Type]
		if !ok {
			log.Fatalf("invalid action type: %s", conf.Type)
		}
		if _, ok := a.actions[conf.Name]; ok {
			log.Fatal("configuration error: cannot redefine action " + conf.Name)
		}
		action, err := provider(generic)
		if err != nil {
			log.Context(log.ACT, conf.Name).Fatal(err)
		}

		a.actions[conf.Name] = action
	}

	return a
}

func (a *Actions) ActionForName(name string) (api.Action, error) {
	action, ok := a.actions[name]
	if !ok {
		return nil, errors.New("configuration error: undefined action " + name)
	}
	return action, nil
}
