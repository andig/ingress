package homie

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/data"
	mq "github.com/andig/ingress/pkg/mqtt"
	"github.com/eclipse/paho.mqtt.golang"
)

// Subscriber Homie/MQTT data source
type Subscriber struct {
	*mq.MqttConnector
	name      string
	rootTopic string
	mux       sync.RWMutex
	devices   []string
	receiver  chan data.Data
}

// NewFromSourceConfig creates Homie/MQTT data source
func NewFromSourceConfig(c config.Source) api.Source {
	topic := c.Topic
	if topic == "" {
		topic = "homie"
	}

	mqttOptions := mq.NewMqttClientOptions(c.URL, c.User, c.Password)
	homieSubscriber := NewSubscriber(c.Name, topic, mqttOptions)
	mqttClient := mqtt.NewClient(mqttOptions)
	homieSubscriber.Connect(mqttClient)
	homieSubscriber.Discover()
	return homieSubscriber
}

// NewSubscriber creates Homie/MQTT data source
func NewSubscriber(name string, rootTopic string, mqttOptions *mqtt.ClientOptions) *Subscriber {
	h := &Subscriber{
		MqttConnector: &mq.MqttConnector{},
		name:          name,
		rootTopic:     mq.StripTrailingSlash(rootTopic),
		devices:       make([]string, 0),
	}

	// connection lost handler
	mqttOptions.SetOnConnectHandler(h.connectionHandler)
	mqttOptions.SetConnectionLostHandler(h.connectionLostHandler)

	return h
}

func (h *Subscriber) connectionHandler(client mqtt.Client) {
	log.Println(h.name + ": connected to " + mq.ServerFromClient(client))
}

func (h *Subscriber) connectionLostHandler(client mqtt.Client, err error) {
	log.Println(h.name + ": disconnected from " + mq.ServerFromClient(client))
}

// Run implements api.Source
func (h *Subscriber) Run(out chan data.Data) {
	h.mux.Lock()
	defer h.mux.Unlock()
	h.receiver = out
}

func (h *Subscriber) listen(topic string) {
	h.MqttClient.Subscribe(topic, 1, func(c mqtt.Client, msg mqtt.Message) {
		log.Printf(h.name+": recv (%s=%s)", msg.Topic(), msg.Payload())

		segments := strings.Split(msg.Topic(), "/")
		name := segments[len(segments)-1]

		payload := string(msg.Payload())
		value, err := strconv.ParseFloat(payload, 64)
		if err != nil {
			log.Printf(h.name+": float conversion error, skipping (%s=%s)", msg.Topic(), payload)
			return
		}

		h.mux.RLock()
		defer h.mux.RUnlock()

		if h.receiver != nil {
			d := data.Data{
				Name:  name,
				Value: value,
			}

			h.receiver <- d
		}
	})
}

// Discover implements api.Source
func (h *Subscriber) Discover() {
	topic := fmt.Sprintf("%s/+/+/+/%s", h.rootTopic, propDatatype)
	h.MqttClient.Subscribe(topic, 1, func(c mqtt.Client, msg mqtt.Message) {
		topic := msg.Topic()
		datatype := msg.Payload()

		// strip datatype
		segments := strings.Split(topic, "/")
		topic = strings.Join(segments[:len(segments)-1], "/")

		if string(datatype) == "float" {
			if h.addDevice(topic) {
				// print only if not already subscribed
				log.Printf(h.name+": discovered %s", topic)
				h.listen(topic)
			}
		} else if len(datatype) == 0 {
			if h.removeDevice(topic) {
				// print only if already subscribed
				log.Printf(h.name+": removed %s", topic)
				h.MqttClient.Unsubscribe(topic)
			}
		} else {
			log.Printf(h.name+": unsupported datatype %s - ignoring %s", datatype, topic)
		}
	})
}

func (h *Subscriber) deviceIndex(topic string) int {
	h.mux.RLock()
	defer h.mux.RUnlock()

	for i, dev := range h.devices {
		if dev == topic {
			return i
		}
	}
	return -1
}

func (h *Subscriber) addDevice(topic string) bool {
	if i := h.deviceIndex(topic); i < 0 {
		h.mux.Lock()
		defer h.mux.Unlock()

		h.devices = append(h.devices, topic)
		return true
	}
	return false
}

func (h *Subscriber) removeDevice(topic string) bool {
	if i := h.deviceIndex(topic); i >= 0 {
		h.mux.Lock()
		defer h.mux.Unlock()

		// remove element i by moving last element to its position
		h.devices[i] = h.devices[len(h.devices)-1]
		h.devices = h.devices[:len(h.devices)-1]
		return true
	}
	return false
}
