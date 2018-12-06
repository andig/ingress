package mqtt

import (
	"fmt"
	"log"
	"strings"

	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/data"
	"github.com/eclipse/paho.mqtt.golang"
)

const patternDelimiter = "%"

type Publisher struct {
	*MqttConnector
	name         string
	topicPattern string
}

func NewFromOutputConfig(c config.Output) *Publisher {
	mqttOptions := NewMqttClientOptions(c.URL, c.User, c.Password)
	mqttPublisher := NewPublisher(c.Name, c.Topic, mqttOptions)
	mqttClient := mqtt.NewClient(mqttOptions)
	mqttPublisher.Connect(mqttClient)

	return mqttPublisher
}

func NewPublisher(name string, topicPattern string, mqttOptions *mqtt.ClientOptions) *Publisher {
	h := &Publisher{
		MqttConnector: &MqttConnector{},
		name:          name,
		topicPattern:  topicPattern,
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

func (h *Publisher) Publish(d data.Data) {
	topic := h.topicPattern
	topic = h.replacePattern(topic, "id", d.ID)
	topic = h.replacePattern(topic, "name", d.Name)

	message := fmt.Sprintf("%.4f", d.Value)
	log.Printf(h.name+": send (%s=%s)", topic, message)

	token := h.MqttClient.Publish(topic, 1, false, message)
	h.WaitForToken(token)
}

func (h *Publisher) replacePattern(input string, pattern string, value string) string {
	pattern = patternDelimiter + pattern + patternDelimiter
	return strings.Replace(input, pattern, value, -1)
}

func (h *Publisher) Discover() {
	log.Println(h.name + ": mqtt does not support discovery")
}
