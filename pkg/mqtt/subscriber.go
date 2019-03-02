package mqtt

import (
	"net/url"
	"regexp"
	"strconv"
	"sync"

	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/data"
	"github.com/andig/ingress/pkg/log"
	"github.com/andig/ingress/pkg/registry"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const topicPattern = "([^/]+$)"

func init() {
	registry.RegisterSource("mqtt", NewFromSourceConfig)
}

// Subscriber is the MQTT data source
type Subscriber struct {
	*MqttConnector
	name      string
	rootTopic string
	mux       sync.Mutex
}

// NewFromSourceConfig creates MQTT data source
func NewFromSourceConfig(c config.Source) (s api.Source, err error) {
	topic := c.Topic
	if topic == "" {
		topic = "#"
	}

	if _, err = url.ParseRequestURI(c.URL); err != nil {
		return s, err
	}

	mqttOptions := NewMqttClientOptions(c.URL, c.User, c.Password)
	mqttSubscriber := NewSubscriber(c.Name, topic, mqttOptions)
	mqttClient := mqtt.NewClient(mqttOptions)
	mqttSubscriber.Connect(mqttClient)

	return mqttSubscriber, nil
}

// NewSubscriber creates MQTT data source
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
	log.Context(log.SRC, h.name).Println("connected to " + ServerFromClient(client))
}

func (h *Subscriber) connectionLostHandler(client mqtt.Client, err error) {
	log.Context(log.SRC, h.name).Warnf("disconnected from " + ServerFromClient(client))
}

// Run implements api.Source
func (h *Subscriber) Run(out chan api.Data) {
	log.Context(log.SRC, h.name).Printf(h.name+": subscribed to topic %s", h.rootTopic)

	h.MqttClient.Subscribe(h.rootTopic, 1, func(c mqtt.Client, msg mqtt.Message) {
		log.Context(log.SRC, h.name).Printf(h.name+": recv (%s=%s)", msg.Topic(), msg.Payload())

		payload := string(msg.Payload())
		value, err := strconv.ParseFloat(payload, 64)
		if err != nil {
			log.Context(log.SRC, h.name).Printf(h.name+": float conversion error, skipping (%s=%s)", msg.Topic(), payload)
			return
		}

		name := h.matchString(msg.Topic(), topicPattern)
		log.Context(log.SRC, h.name).Printf(h.name+": matched topic (id=%s,name=%s)", name, name)

		data := data.New(name, value)
		out <- data
	})
}

func (h *Subscriber) matchString(s string, pattern string) string {
	re, err := regexp.Compile(topicPattern)
	if err != nil {
		log.Context(log.SRC, h.name).Fatal("invalid regex pattern " + pattern)
	}

	matches := re.FindStringSubmatch(s)
	if matches != nil && len(matches) == 2 {
		return matches[1]
	}

	return ""
}
