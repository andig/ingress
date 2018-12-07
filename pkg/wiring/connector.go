package wiring

import (
	"log"
	"sync"

	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/data"

	"github.com/andig/ingress/pkg/homie"
	"github.com/andig/ingress/pkg/mqtt"
	"github.com/andig/ingress/pkg/telemetry"
	"github.com/andig/ingress/pkg/volkszaehler"
)

type Publisher interface {
	// NewFromOutputConfig(c config.Output)
	Discover()
	Publish(d data.Data)
}

type Subscriber interface {
	// NewFromInputConfig(c config.Input)
	Run(receiver chan data.Data)
}

type SubscriberMap map[string]Subscriber
type PublisherMap map[string]Publisher

type Connectors struct {
	mux    sync.Mutex
	Input  SubscriberMap
	Output PublisherMap
}

// NewConnectors creates the input and output system connectors
func NewConnectors(i []config.Input, o []config.Output) *Connectors {
	c := Connectors{
		Input:  make(SubscriberMap),
		Output: make(PublisherMap),
	}

	for _, input := range i {
		c.createInputConnector(input)
	}
	for _, output := range o {
		c.createOutputConnector(output)
	}

	// activate telemetry if configured
	c.ApplyTelemetry()

	return &c
}

// ApplyTelemetry wires metric providers to the Telemetry instance
func (c *Connectors) ApplyTelemetry() {
	c.mux.Lock()
	defer c.mux.Unlock()
	
	for _, input := range c.Input {
		// find telemetry instance
		if instance, ok := input.(*telemetry.Telemetry); ok {
			// add metric providers from input
			for _, source := range c.Input {
				if metricProvider, ok := source.(telemetry.MetricProvider); ok {
					instance.AddProvider(metricProvider)
				}
			}

			// add metric providers from output
			for _, source := range c.Output {
				if metricProvider, ok := source.(telemetry.MetricProvider); ok {
					instance.AddProvider(metricProvider)
				}
			}

			// log.Println("connector: activated metrics collection")
			log.Println("enabled metrics collection")
			return
		}
	}
}

func (c *Connectors) createInputConnector(i config.Input) {
	var conn Subscriber
	switch i.Type {
	case "telemetry":
		conn = telemetry.NewFromInputConfig(i)
		break
	case "mqtt":
		conn = mqtt.NewFromInputConfig(i)
		break
	case "homie":
		conn = homie.NewFromInputConfig(i)
		break
	default:
		panic("Invalid input type: " + i.Type)
	}

	c.mux.Lock()
	defer c.mux.Unlock()
	c.Input[i.Name] = conn
}

func (c *Connectors) createOutputConnector(o config.Output) {
	var conn Publisher
	switch o.Type {
	case "mqtt":
		conn = mqtt.NewFromOutputConfig(o)
		break
	case "volkszaehler":
		conn = volkszaehler.NewFromOutputConfig(o)
		break
	default:
		panic("Invalid output type: " + o.Type)
	}

	c.mux.Lock()
	defer c.mux.Unlock()
	c.Output[o.Name] = conn
}

// Run starts each subscriber's Run() function in a gofunc
func (c *Connectors) Run(mapper *Mapper) {
	for name, input := range c.Input {
		log.Printf("connector: starting %s", name)
		c := make(chan data.Data)

		// start distributor
		go func(source string, c chan data.Data) {
			log.Printf("connector: recv from %s", name)
			for {
				d := <-c
				log.Printf("connector: recv from %s (%s=%f)", source, d.Name, d.Value)
				go mapper.Process(source, &d)
			}
		}(name, c)

		// start subscriber
		go input.Run(c)
	}
}
