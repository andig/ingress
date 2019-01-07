package mqtt

import (
	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
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
	log.Context(log.TGT, h.name).Println("connected to " + ServerFromClient(client))
}

func (h *Publisher) connectionLostHandler(client mqtt.Client, err error) {
	log.Context(log.TGT, h.name).Warnf("disconnected from " + ServerFromClient(client))
}

// Publish implements api.Source
func (h *Publisher) Publish(d api.Data) {
	topic := d.MatchPattern(h.topic)
	message := d.ValStr()
	log.Context(
		log.TGT, h.name,
		log.EV, topic,
		log.VAL, message,
	).Debugf("send")

	token := h.MqttClient.Publish(topic, 1, false, message)
	h.WaitForToken(token, defaultTimeout)
}
