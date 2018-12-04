package main

import (
	_ "log"
	"time"

	. "github.com/andig/ingress"
	"github.com/andig/ingress/pkg/homie"
)

func main() {
	var c Config
	c.LoadConfig("config.yml")

	connectors := NewConnectors(c)
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
