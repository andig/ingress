package main

import (
	"time"

	"github.com/andig/ingress/pkg/homie"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func NewMqttClientOptions() *mqtt.ClientOptions {
	mqttOptions := mqtt.NewClientOptions()
	mqttOptions.AddBroker("tcp://localhost:1883")
	// mqttOptions.SetUsername(mqttUser)
	// mqttOptions.SetPassword(mqttPassword)
	// mqttOptions.SetClientID(mqttClientID)
	// mqttOptions.SetCleanSession(mqttCleanSession)
	mqttOptions.SetAutoReconnect(true)
	return mqttOptions
}

func main() {
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

	mqttOptions := NewMqttClientOptions()
	homiePublisher := homie.NewPublisher("homie", *dev, mqttOptions) // refine mqtt client options
	mqttClient := mqtt.NewClient(mqttOptions)
	homiePublisher.Connect(mqttClient)
	homiePublisher.Publish()
	go homiePublisher.Run()

	homieSubscriber := homie.NewSubscriber("homie")
	homieSubscriber.Connect(mqttClient)
	homieSubscriber.Discover()

	go homieSubscriber.Run()

	time.Sleep(3 * time.Second)
}
