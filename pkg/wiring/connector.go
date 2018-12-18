package wiring

import (
	"errors"
	"log"
	"strings"
	"sync"

	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/data"

	"github.com/andig/ingress/pkg/http"
	"github.com/andig/ingress/pkg/homie"
	"github.com/andig/ingress/pkg/mqtt"
	"github.com/andig/ingress/pkg/telemetry"
	"github.com/andig/ingress/pkg/volkszaehler"
)

// Connectors manages data sources and targets
type Connectors struct {
	mux      sync.Mutex
	sources  map[string]api.Source
	targets  map[string]api.Target
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

	var conn api.Source
	switch strings.ToLower(conf.Type) {
	case "telemetry":
		conn = telemetry.NewFromSourceConfig(conf)
		break
	case "mqtt":
		conn = mqtt.NewFromSourceConfig(conf)
		break
	case "homie":
		conn = homie.NewFromSourceConfig(conf)
		break
	default:
		log.Fatal("connectors: invalid source type: " + conf.Type)
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	if _, err := c.SourceForName(conf.Name); err == nil {
		log.Fatal("configuration error: cannot redefine source "+ conf.Name)
	}
	c.sources[conf.Name] = conn
}

func (c *Connectors) createTargetConnector(conf config.Target) {
	if conf.Name == "" {
		log.Fatal("configuration error: missing target name")
	}

	var conn api.Target
	switch conf.Type {
	case "http":
		conn = http.NewFromTargetConfig(conf)
		break
	case "mqtt":
		conn = mqtt.NewFromTargetConfig(conf)
		break
	case "volkszaehler":
		conn = volkszaehler.NewFromTargetConfig(conf)
		break
	default:
		log.Fatal("Invalid output type: " + conf.Type)
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	if _, err := c.TargetForName(conf.Name); err == nil {
		log.Fatal("configuration error: cannot redefine target "+ conf.Name)
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

			// log.Println("connector: activated metrics collection")
			log.Println("enabled metrics collection")
			return
		}
	}
}

// SourceForName returns a data source identified by source name
func (c *Connectors) SourceForName(name string) (api.Source, error) {
	source, ok := c.sources[name]
	if !ok {
		return nil, errors.New("configuration error: undefined source "+name)
	}
	return source, nil
}

// TargetForName returns a data target identified by target name
func (c *Connectors) TargetForName(name string) (api.Target, error) {
	target, ok := c.targets[name]
	if !ok {
		return nil, errors.New("configuration error: undefined target "+name)
	}
	return target, nil
}

// Run starts each Source's Run() function in a gofunc
func (c *Connectors) Run(mapper *Mapper) {
	for name, source := range c.sources {
		log.Printf("connector: starting %s", name)
		c := make(chan data.Data)

		// start distributor
		go func(name string, c chan data.Data) {
			for {
				d := <-c
				log.Printf("connector: recv from %s (%s=%f)", name, d.Name, d.Value)
				go mapper.Process(name, d)
			}
		}(name, c)

		// start source connector
		go source.Run(c)
	}
}
