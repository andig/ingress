package wiring

import (
	"log"
	"strings"

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
	sources  *data.Set
	targets  *data.Set
}

// NewConnectors creates the source and output system connectors
func NewConnectors(i []config.Source, o []config.Target) *Connectors {
	c := Connectors{
		sources: data.NewSet(),
		targets: data.NewSet(),
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
		log.Fatal("configuration error: invalid source type: " + conf.Type)
	}

	if !c.sources.Add(conf.Name, conn) {
		log.Fatal("configuration error: cannot redefine source "+ conf.Name)
	}
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

	if !c.targets.Add(conf.Name, conn) {
		log.Fatal("configuration error: cannot redefine target "+ conf.Name)
	}
}

// ApplyTelemetry wires metric providers to the Telemetry instance
func (c *Connectors) ApplyTelemetry() {
	 for _, v := range c.sources.Values() {
		instance, ok := v.(telemetry.Telemetry)
		if !ok {
			continue
		}
	 
		// add metric providers from Source
		for _, source := range c.sources.Values() {
			if metricProvider, ok := source.(telemetry.MetricProvider); ok {
				instance.AddProvider(metricProvider)
			}
		}

		// add metric providers from output
		for _, target := range c.targets.Values() {
			if metricProvider, ok := target.(telemetry.MetricProvider); ok {
				instance.AddProvider(metricProvider)
			}
		}

		// log.Println("connector: activated metrics collection")
		log.Println("enabled metrics collection")
	}
}

// Run starts each Source's Run() function in a gofunc
func (c *Connectors) Run(mapper *Mapper) {
	for _,name := range c.sources.Keys() {
		source := c.sources.Get(name).(api.Source)
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
