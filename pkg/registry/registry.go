package registry

import (
	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/log"
)

type SourceProvider func(config.Generic) (api.Source, error)
type TargetProvider func(config.Generic) (api.Target, error)
type ActionProvider func(config.Generic) (api.Action, error)

var SourceProviders = make(map[string]SourceProvider)
var TargetProviders = make(map[string]TargetProvider)
var ActionProviders = make(map[string]ActionProvider)

func RegisterSource(name string, provider SourceProvider) {
	if _, ok := SourceProviders[name]; ok {
		log.Fatalf("Source %s already defined", name)
	}

	SourceProviders[name] = provider
}

func RegisterTarget(name string, provider TargetProvider) {
	if _, ok := TargetProviders[name]; ok {
		log.Fatalf("Target %s already defined", name)
	}

	TargetProviders[name] = provider
}

func RegisterAction(name string, provider ActionProvider) {
	if _, ok := ActionProviders[name]; ok {
		log.Fatalf("Action %s already defined", name)
	}

	ActionProviders[name] = provider
}
