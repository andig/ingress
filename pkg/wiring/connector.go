package wiring

import (
	"context"
	"errors"
	"sync"

	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/log"
	"github.com/andig/ingress/pkg/registry"

	"github.com/andig/ingress/pkg/telemetry"
)

// Connectors manages data sources and targets
type Connectors struct {
	mux     sync.Mutex
	sources map[string]api.Source
	targets map[string]api.Target
}

// NewConnectors creates the source and output system connectors
func NewConnectors(i []config.Source, o []config.Target) *Connectors {
	c := Connectors{
		sources: make(map[string]api.Source),
		targets: make(map[string]api.Target),
	}

	for _, source := range i {
		c.createSourceConnector(source)
	}
	for _, target := range o {
		c.createTargetConnector(target)
	}

	// activate telemetry if configured
	c.ApplyTelemetry()

	return &c
}

func (c *Connectors) createSourceConnector(conf config.Source) {
	if conf.Name == "" {
		log.Fatal("configuration error: missing source name")
	}

	provider, ok := registry.SourceProviders[conf.Type]
	if !ok {
		log.Fatalf("Invalid source type: %s", conf.Type)
	}

	conn, err := provider(conf)
	if err != nil {
		log.Context(log.TGT, conf.Name).Fatal(err)
	}

	if err != nil {
		log.Context(log.SRC, conf.Name).Fatal(err)
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	if _, err := c.SourceForName(conf.Name); err == nil {
		log.Fatal("configuration error: cannot redefine source " + conf.Name)
	}
	c.sources[conf.Name] = conn
}

func (c *Connectors) createTargetConnector(conf config.Target) {
	if conf.Name == "" {
		log.Fatal("configuration error: missing target name")
	}

	provider, ok := registry.TargetProviders[conf.Type]
	if !ok {
		log.Fatalf("Invalid target type: %s", conf.Type)
	}

	conn, err := provider(conf)
	if err != nil {
		log.Context(log.TGT, conf.Name).Fatal(err)
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	if _, err := c.TargetForName(conf.Name); err == nil {
		log.Fatal("configuration error: cannot redefine target " + conf.Name)
	}
	c.targets[conf.Name] = conn
}

// ApplyTelemetry wires metric providers to the Telemetry instance
func (c *Connectors) ApplyTelemetry() {
	c.mux.Lock()
	defer c.mux.Unlock()

	for _, Source := range c.sources {
		// find telemetry instance
		if instance, ok := Source.(*telemetry.Telemetry); ok {
			// add metric providers from Source
			for _, source := range c.sources {
				if metricProvider, ok := source.(telemetry.MetricProvider); ok {
					instance.AddProvider(metricProvider)
				}
			}

			// add metric providers from output
			for _, target := range c.targets {
				if metricProvider, ok := target.(telemetry.MetricProvider); ok {
					instance.AddProvider(metricProvider)
				}
			}

			// log.Println("activated metrics collection")
			log.Println("enabled metrics collection")
			return
		}
	}
}

// SourceForName returns a data source identified by source name
func (c *Connectors) SourceForName(name string) (api.Source, error) {
	source, ok := c.sources[name]
	if !ok {
		return nil, errors.New("configuration error: undefined source " + name)
	}
	return source, nil
}

// TargetForName returns a data target identified by target name
func (c *Connectors) TargetForName(name string) (api.Target, error) {
	target, ok := c.targets[name]
	if !ok {
		return nil, errors.New("configuration error: undefined target " + name)
	}
	return target, nil
}

// Run starts each Source's Run() function in a gofunc
func (c *Connectors) Run(ctx context.Context, mapper *Mapper) {
	for name, source := range c.sources {
		log.Context(log.SRC, name).Printf("starting event loop")
		c := make(chan api.Data)

		// start distributor
		go func(name string, c chan api.Data) {
			for {
				select {
				case <-ctx.Done():
					return
				case d := <-c:
					log.Context(
						log.SRC, name,
						log.EV, d.GetName(),
					).Debugf("processing")
					go mapper.Process(name, d)
				}
			}
		}(name, c)

		// start source connector
		go source.Run(c)
	}
}
