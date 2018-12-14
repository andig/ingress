package mqtt

import (
	"log"

	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/data"

	"github.com/eclipse/paho.mqtt.golang"
)

// Publisher is the MQTT data target
type Publisher struct {
	*MqttConnector
	name  string
	topic string
}

// NewFromTargetConfig creates MQTT data target
func NewFromTargetConfig(c config.Target) api.Target {
	mqttOptions := NewMqttClientOptions(c.URL, c.User, c.Password)
	mqttPublisher := NewPublisher(c.Name, c.Topic, mqttOptions)
	mqttClient := mqtt.NewClient(mqttOptions)
	mqttPublisher.Connect(mqttClient)

	return mqttPublisher
}

// NewPublisher creates MQTT data target
func NewPublisher(name string, topic string, mqttOptions *mqtt.ClientOptions) *Publisher {
	if topic == "" {
		topic = "ingress/%name%"
	}

	h := &Publisher{
		MqttConnector: &MqttConnector{},
		name:          name,
		topic:         topic,
	}

	// connection lost handler
	mqttOptions.SetOnConnectHandler(h.connectionHandler)
	mqttOptions.SetConnectionLostHandler(h.connectionLostHandler)

	return h
}

func (h *Publisher) connectionHandler(client mqtt.Client) {
	log.Println(h.name + ": connected to " + ServerFromClient(client))
}

func (h *Publisher) connectionLostHandler(client mqtt.Client, err error) {
	log.Println(h.name + ": disconnected from " + ServerFromClient(client))
}

// Publish implements api.Source
func (h *Publisher) Publish(d data.Data) {
	topic := d.MatchPattern(h.topic)
	message := d.ValStr()
	log.Printf(h.name+": send (%s=%s)", topic, message)

	token := h.MqttClient.Publish(topic, 1, false, message)
	h.WaitForToken(token, defaultTimeout)
}

// Discover implements api.Source
func (h *Publisher) Discover() {
	log.Println(h.name + ": mqtt does not support discovery")
}
