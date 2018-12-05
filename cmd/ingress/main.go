package main

import (
	_ "log"
	"time"

	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/homie"
	"github.com/andig/ingress/pkg/wiring"
)

func main() {
	var c config.Config
	c.LoadConfig("config.yml")

	connectors := wiring.NewConnectors(c)
	mapper := wiring.NewMapper(c.Mapper, connectors.Output)
	go connectors.Run(mapper)
	_ = connectors

	dev := &homie.Device{
		Name: "meter1",
		Nodes: []*homie.Node{
			&homie.Node{
				Name: "zaehlwerk1",
				Properties: []*homie.Property{
					&homie.Property{
						Name: "power",
					},
					&homie.Property{
						Name: "zaehlerstand",
					},
				},
			},
		},
	}
	_ = dev

	// mqttOptions := NewMqttClientOptions()
	// homiePublisher := homie.NewPublisher("homie", *dev, mqttOptions) // refine mqtt client options
	// mqttClient := mqtt.NewClient(mqttOptions)
	// homiePublisher.Connect(mqttClient)
	// homiePublisher.Publish()
	// go homiePublisher.Run()

	time.Sleep(3 * time.Second)
}
