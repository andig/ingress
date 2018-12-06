package mqtt

import (
	"log"
	"regexp"
	"strconv"
	"sync"

	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/data"
	"github.com/eclipse/paho.mqtt.golang"
)

const topicPattern = "([^/]+$)"

type Subscriber struct {
	*MqttConnector
	name      string
	rootTopic string
	mux       sync.Mutex
}

func NewFromInputConfig(c config.Input) *Subscriber {
	topic := c.Topic
	if topic == "" {
		topic = "#"
	}

	mqttOptions := NewMqttClientOptions(c.URL, c.User, c.Password)
	mqttSubscriber := NewSubscriber(c.Name, topic, mqttOptions)
	mqttClient := mqtt.NewClient(mqttOptions)
	mqttSubscriber.Connect(mqttClient)

	return mqttSubscriber
}

func NewSubscriber(name string, rootTopic string, mqttOptions *mqtt.ClientOptions) *Subscriber {
	h := &Subscriber{
		MqttConnector: &MqttConnector{},
		name:          name,
		rootTopic:     StripTrailingSlash(rootTopic),
	}

	// connection lost handler
	mqttOptions.SetOnConnectHandler(h.connectionHandler)
	mqttOptions.SetConnectionLostHandler(h.connectionLostHandler)

	return h
}

func (h *Subscriber) connectionHandler(client mqtt.Client) {
	log.Println(h.name + ": connected to " + ServerFromClient(client))
}

func (h *Subscriber) connectionLostHandler(client mqtt.Client, err error) {
	log.Println(h.name + ": disconnected from " + ServerFromClient(client))
}

func (h *Subscriber) Run(out chan data.Data) {
	log.Printf(h.name+": subscribed to topic %s", h.rootTopic)

	h.MqttClient.Subscribe(h.rootTopic, 1, func(c mqtt.Client, msg mqtt.Message) {
		log.Printf(h.name+": recv (%s=%s)", msg.Topic(), msg.Payload())

		payload := string(msg.Payload())
		value, err := strconv.ParseFloat(payload, 64)
		if err != nil {
			log.Printf(h.name+": float conversion error, skipping (%s=%s)", msg.Topic(), payload)
		}

		name := h.matchString(msg.Topic(), topicPattern)
		log.Printf(h.name+": matched topic (id=%s,name=%s)", name, name)

		data := data.Data{
			ID:    name,
			Name:  name,
			Value: value,
		}
		out <- data
	})
}

func (h *Subscriber) matchString(s string, pattern string) string {
	re, err := regexp.Compile(topicPattern)
	if err != nil {
		panic(h.name + ": invalid regex pattern " + pattern)
	}

	matches := re.FindStringSubmatch(s)
	if matches != nil && len(matches) == 2 {
		return matches[1]
	}

	return ""
}
