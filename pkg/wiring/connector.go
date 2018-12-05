package wiring

import (
	"log"
	"sync"

	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/data"

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

func NewConnectors(c config.Config) *Connectors {
	conn := Connectors{
		Input:  make(SubscriberMap),
		Output: make(PublisherMap),
	}

	for _, input := range c.Input {
		conn.createInputConnector(input)
	}
	for _, output := range c.Output {
		conn.createOutputConnector(output)
	}

	return &conn
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
	case "default":
		panic("Invalid input type: " + i.Type)
	}

	c.mux.Lock()
	defer c.mux.Unlock()
	c.Input[i.Name] = conn
}

func (c *Connectors) createOutputConnector(o config.Output) {
	var conn Publisher
	switch o.Type {
	case "volkszaehler":
		conn = volkszaehler.NewFromOutputConfig(o)
		break
	case "default":
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
