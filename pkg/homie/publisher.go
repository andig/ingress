package homie

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Publisher struct {
	*MqttConnector
	rootTopic string
	dev       Device
}

func NewPublisher(rootTopic string, dev Device, mqttOptions *mqtt.ClientOptions) *Publisher {
	h := &Publisher{
		MqttConnector: &MqttConnector{},
		rootTopic:     stripTrailingSlash(rootTopic),
		dev:           dev,
	}

	topic := fmt.Sprintf("%s/%s/%s", h.rootTopic, h.dev.Name, propState)
	mqttOptions.SetWill(topic, propStateLost, mqttOptions.WillQos, true)

	// connection lost handler
	mqttOptions.SetOnConnectHandler(h.connectionHandler)
	mqttOptions.SetConnectionLostHandler(h.connectionLostHandler)

	return h
}

func (h *Publisher) Publish() {
	// h.publishReady()
	for _, node := range h.dev.Nodes {
		for _, property := range node.Properties {
			topic := fmt.Sprintf("%s/%s/%s/%s/%s", h.rootTopic, h.dev.Name, node.Name, property.Name, propDatatype)
			h.publish(topic, true, propDatatypeFloat)
		}
	}
}

// func (h *Publisher) publishReady() {
// 	topic := fmt.Sprintf("%s/%s/%s", h.rootTopic, h.dev.Name, propState)
// 	go h.publish(topic, true, propStateReady)
// }

func (h *Publisher) connectionHandler(client mqtt.Client) {
	// log.Println("mqtt: connected")
	// h.publishReady()
	topic := fmt.Sprintf("%s/%s/%s", h.rootTopic, h.dev.Name, propState)
	go h.publish(topic, true, propStateReady)
}

func (h *Publisher) connectionLostHandler(client mqtt.Client, err error) {
	// log.Println("mqtt: disconnected")
}

func (h *Publisher) publish(topic string, retained bool, message interface{}) {
	token := h.mqttClient.Publish(topic, 1, retained, message)
	h.WaitForToken(token)
}

func (h *Publisher) Run() {
}
