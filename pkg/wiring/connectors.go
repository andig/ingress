package wiring

import (
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
	Run()
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
