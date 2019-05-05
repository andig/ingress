package mqtt

import (
	"net/url"

	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/log"
	"github.com/andig/ingress/pkg/registry"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func init() {
	registry.RegisterTarget("mqtt", NewFromTargetConfig)
}

// Publisher is the MQTT data target
type Publisher struct {
	*Connector
	name  string
	topic string
}

// NewFromTargetConfig creates MQTT data target
func NewFromTargetConfig(g config.Generic) (t api.Target, err error) {
	var c mqttConfig
	err = config.Decode(g, &c)
	if err != nil {
		return nil, err
	}

	if _, err = url.ParseRequestURI(c.URL); err != nil {
		return t, err
	}

	mqttOptions := NewMqttClientOptions(c.URL, c.User, c.Password)
	mqttPublisher := NewPublisher(c.Name, c.Topic, mqttOptions)
	mqttClient := mqtt.NewClient(mqttOptions)
	mqttPublisher.Connect(mqttClient)

	return mqttPublisher, nil
}

// NewPublisher creates MQTT data target
func NewPublisher(name string, topic string, mqttOptions *mqtt.ClientOptions) *Publisher {
	if topic == "" {
		topic = "ingress/%name%"
	}

	h := &Publisher{
		Connector: &Connector{},
		name:      name,
		topic:     topic,
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
