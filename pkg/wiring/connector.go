package wiring

import (
	"log"
	"sync"

	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/data"
	"github.com/andig/ingress/pkg/telemetry"

	"github.com/andig/ingress/pkg/homie"
	"github.com/andig/ingress/pkg/mqtt"
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

	// c.startTelmetry()

	return &c
}

func (c *Connectors) startTelmetry() {
	c.mux.Lock()
	defer c.mux.Unlock()
	telemetry := &telemetry.Telemetry{}
	c.Input["telemetry"] = telemetry
}

func (c *Connectors) createInputConnector(i config.Input) {
	var conn Subscriber
	switch i.Type {
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
		c := make(chan data.Data)

		// start distributor
		go func(name string, c chan data.Data) {
			log.Printf("connector: recv from %s", name)
			for {
				d := <-c
				log.Printf("connector: recv from %s (%s=%f)", name, d.Name, d.Value)
				i := &data.InputData{
					Source: name,
					Data:   &d,
				}

				go mapper.Process(i)
			}
		}(name, c)

		// start subscriber
		go input.Run(c)
	}
}
