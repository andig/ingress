package ingress

import (
	"log"
	"time"
	"sync"

	"github.com/andig/ingress/pkg/homie"
	"github.com/andig/ingress/pkg/volkszaehler"
	"github.com/eclipse/paho.mqtt.golang"
)

type Connectors struct {
	mux sync.Mutex
	Input  map[string]interface{}
	Output map[string]interface{}
}

func Discover(vz *volkszaehler.Api, entities []volkszaehler.Entity) {
	for _, e := range entities {
		log.Printf("%s %s: %s", e.UUID, e.Type, e.Title)
	}
	for _, e := range entities {
		if e.Type == volkszaehler.TypeGroup {
			children := vz.GetEntity(e.UUID).Children
			Discover(vz, children)
		}
	}
}

func NewConnectors(c Config) Connectors {
	conn := Connectors{
		Input: make(map[string]interface{}),
		Output: make(map[string]interface{}),
	}
	
	for _, input := range c.Input {
		conn.createInputConnector(input)
	}
	for _, output := range c.Output {
		conn.createOutputConnector(output)
	}

	return conn
}

func (c *Connectors)createInputConnector(i Input) {
	c.mux.Lock()
	defer c.mux.Unlock()

	var conn interface{}
	switch i.Type {
	case "homie":
		mqttOptions := homie.NewMqttClientOptions(i.Url, i.User, i.Password)
		mqttClient := mqtt.NewClient(mqttOptions)
		homieSubscriber := homie.NewSubscriber("homie")
		homieSubscriber.Connect(mqttClient)
		homieSubscriber.Discover()
		conn = homieSubscriber
		break
	case "default":
		panic("Invalid input type: " + i.Type)
	}

	c.Input[i.Name] = conn
}

func (c *Connectors)createOutputConnector(o Output) {
	c.mux.Lock()
	defer c.mux.Unlock()

	var conn interface{}
	switch o.Type {
	case "volkszaehler":
		vz := volkszaehler.NewAPI(o.Url, 1*time.Second, false)
		Discover(vz, vz.GetPublicEntities())
		conn = vz
		break
	case "default":
		panic("Invalid output type: " + o.Type)
	}

	c.Output[o.Name] = conn
}
