package main

import (
	"log"
	"time"

	"github.com/andig/ingress/pkg/config"
	. "github.com/andig/ingress/pkg/mqtt"
	"github.com/andig/ingress/pkg/wiring"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func inject() {
	mqttOptions := NewMqttClientOptions("tcp://localhost:1883", "", "")
	mqttClient := mqtt.NewClient(mqttOptions)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("mqtt: error connecting: ", token.Error())
		panic(token.Error())
	}

	time.Sleep(200 * time.Millisecond)
	token := mqttClient.Publish("input/inject", 0, false, "3.14")
	if token.WaitTimeout(100 * time.Millisecond) {
		log.Println("--> inject done")
	}
}

func main() {
	var c config.Config
	c.LoadConfig("config.yml")

	connectors := wiring.NewConnectors(c.Input, c.Output)
	mapper := wiring.NewMapper(c.Wiring, connectors.Input, connectors.Output)
	go connectors.Run(mapper)

	// test data
	inject()

	time.Sleep(1 * time.Second)
}
