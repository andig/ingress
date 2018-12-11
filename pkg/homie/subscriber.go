package homie

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/data"
	mq "github.com/andig/ingress/pkg/mqtt"
	"github.com/eclipse/paho.mqtt.golang"
)

type Subscriber struct {
	*mq.MqttConnector
	name      string
	rootTopic string
	mux       sync.Mutex
	devices   []string
	receiver  chan data.Data
}

func NewFromSourceConfig(c config.Source) *Subscriber {
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

func (h *Subscriber) Run(out chan data.Data) {
	log.Printf(h.name+": subscribed to topic %s", h.rootTopic)

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

		if h.receiver != nil {
			d := data.Data{
				Name:  name,
				Value: value,
			}

			h.receiver <- d
		}
	})
}

func (h *Subscriber) Discover() {
	topic := fmt.Sprintf("%s/+/+/+/%s", h.rootTopic, propDatatype)
	h.MqttClient.Subscribe(topic, 1, func(c mqtt.Client, msg mqtt.Message) {
		topic := msg.Topic()
		datatype := msg.Payload()

		// strip datatype
		segments := strings.Split(topic, "/")
		topic = strings.Join(segments[:len(segments)-1], "/")

		if string(datatype) == "float" {
			h.mux.Lock()
			defer h.mux.Unlock()
			h.devices = append(h.devices, topic)
			h.listen(topic)
		} else {
			log.Printf(h.name+": unsupported datatype %s - ignoring %s", datatype, topic)
		}
	})
}

/*
func (h *Subscriber) discoverDevice(topic string) {
	segments := strings.Split(topic, "/")

	if len(segments) == 4 {
		log.Printf(h.name+": discovered %s/%s/%s", segments[1], segments[2], segments[3])
		h.mergeDevice(topic, segments[1], segments[2], segments[3])
	} else {
		log.Printf(h.name+": discovered unexpected device %s", topic)
	}
}

func (h *Subscriber) mergeDevice(topic string, deviceName string, nodeName string, propertyName string) {
	h.mux.Lock()
	defer h.mux.Unlock()

	// find or create device
	var device *Device
	for _, d := range h.Devices {
		if d.Name == deviceName {
			device = d
			break
		}
	}

	if device == nil {
		device = &Device{
			Name: deviceName,
		}
		h.Devices = append(h.Devices, device)
	}

	// find or create node
	var node *Node
	for _, n := range device.Nodes {
		if n.Name == nodeName {
			node = n
			break
		}
	}

	if node == nil {
		node = &Node{
			Name: nodeName,
		}
		device.Nodes = append(device.Nodes, node)
	}

	// find or create property
	var property *Property
	for _, p := range node.Properties {
		if p.Name == propertyName {
			property = p
			break
		}
	}

	if property == nil {
		property = &Property{
			Name: propertyName,
		}
		node.Properties = append(node.Properties, property)
	}
}
*/
