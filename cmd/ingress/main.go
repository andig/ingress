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

	token = mqttClient.Publish("homie/meter1/zaehlwerk1/power", 0, false, "4711")
	if token.WaitTimeout(100 * time.Millisecond) {
		log.Println("--> inject done")
	}
}

func main() {
	var c config.Config
	c.Load("config.yml")
	c.Dump()

	connectors := wiring.NewConnectors(c.Sources, c.Targets)
	mappings := wiring.NewMappings(c.Mappings, connectors)
	wires := wiring.NewWiring(c.Wires, mappings, connectors)
	mapper := wiring.NewMapper(wires, connectors)
	go connectors.Run(mapper)

	// test data
	inject()

	time.Sleep(3 * time.Second)
}
