package registry

import (
	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/log"
)

type SourceProvider func(config.Source) (api.Source, error)
type TargetProvider func(config.Target) (api.Target, error)

var SourceProviders map[string]SourceProvider
var TargetProviders map[string]TargetProvider

func init() {
	SourceProviders = make(map[string]SourceProvider)
	TargetProviders = make(map[string]TargetProvider)
}

func RegisterSource(name string, provider SourceProvider) {
	// var once sync.Once
	// once.Do(func() {
	// 	SourceProviders = make(map[string]api.Source)
	// })

	if _, ok := SourceProviders[name]; ok {
		log.Fatalf("Source %s already defined", name)
	}

	SourceProviders[name] = provider
}

func RegisterTarget(name string, provider TargetProvider) {
	// var once sync.Once
	// once.Do(func() {
	// 	TargetProviders = make(map[string]api.Target)
	// })

	if _, ok := TargetProviders[name]; ok {
		log.Fatalf("Target %s already defined", name)
	}

	TargetProviders[name] = provider
}
