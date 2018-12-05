package homie

import (
	"fmt"
	"log"

	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/data"
	mq "github.com/andig/ingress/pkg/mqtt"
	"github.com/eclipse/paho.mqtt.golang"
)

type Publisher struct {
	*mq.MqttConnector
	rootTopic string
	dev       Device
}

func NewFromOutputConfig(c config.Output) *Publisher {
	return nil
}

func NewPublisher(rootTopic string, dev Device, mqttOptions *mqtt.ClientOptions) *Publisher {
	h := &Publisher{
		MqttConnector: &mq.MqttConnector{},
		rootTopic:     mq.StripTrailingSlash(rootTopic),
		dev:           dev,
	}

	topic := fmt.Sprintf("%s/%s/%s", h.rootTopic, h.dev.Name, propState)
	mqttOptions.SetWill(topic, propStateLost, mqttOptions.WillQos, true)

	// connection lost handler
	mqttOptions.SetOnConnectHandler(h.connectionHandler)
	mqttOptions.SetConnectionLostHandler(h.connectionLostHandler)

	return h
}

func (h *Publisher) connectionHandler(client mqtt.Client) {
	log.Println("mqtt: connected")
	topic := fmt.Sprintf("%s/%s/%s", h.rootTopic, h.dev.Name, propState)
	go h.publish(topic, true, propStateReady)
}

func (h *Publisher) connectionLostHandler(client mqtt.Client, err error) {
	log.Println("mqtt: disconnected")
}

func (h *Publisher) Discover() {
	panic("not implemented")
}

func (h *Publisher) Publish(d data.Data) {
	// h.publishReady()
	for _, node := range h.dev.Nodes {
		for _, property := range node.Properties {
			topic := fmt.Sprintf("%s/%s/%s/%s/%s", h.rootTopic, h.dev.Name, node.Name, property.Name, propDatatype)
			h.publish(topic, true, propDatatypeFloat)
		}
	}
}

func (h *Publisher) publish(topic string, retained bool, message interface{}) {
	token := h.MqttClient.Publish(topic, 1, retained, message)
	h.WaitForToken(token)
}

func (h *Publisher) Run() {
}
