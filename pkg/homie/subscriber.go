package homie

import (
	"fmt"
	"log"
	"regexp"
	"sync"

	"github.com/andig/ingress/pkg/config"
	mq "github.com/andig/ingress/pkg/mqtt"
	"github.com/eclipse/paho.mqtt.golang"
)

type Subscriber struct {
	*mq.MqttConnector
	rootTopic string
	mux       sync.Mutex
	Devices   []*Device
}

func NewFromInputConfig(c config.Input) *Subscriber {
	mqttOptions := mq.NewMqttClientOptions(c.URL, c.User, c.Password)
	homieSubscriber := NewSubscriber("homie", mqttOptions)
	mqttClient := mqtt.NewClient(mqttOptions)
	homieSubscriber.Connect(mqttClient)
	homieSubscriber.Discover()
	return homieSubscriber
}

func NewSubscriber(rootTopic string, mqttOptions *mqtt.ClientOptions) *Subscriber {
	h := &Subscriber{
		MqttConnector: &mq.MqttConnector{},
		rootTopic:     mq.StripTrailingSlash(rootTopic),
		Devices:       []*Device{},
	}

	// connection lost handler
	mqttOptions.SetOnConnectHandler(h.connectionHandler)
	mqttOptions.SetConnectionLostHandler(h.connectionLostHandler)

	return h
}

func (h *Subscriber) connectionHandler(client mqtt.Client) {
	log.Println("mqtt: connected")
}

func (h *Subscriber) connectionLostHandler(client mqtt.Client, err error) {
	log.Println("mqtt: disconnected")
}

func (h *Subscriber) Run() {
	topic := fmt.Sprintf("%s/+/+/+", h.rootTopic)
	h.MqttClient.Subscribe(topic, 1, func(c mqtt.Client, msg mqtt.Message) {
		log.Printf("homie: received payload %s", msg.Payload())
	})
}

func (h *Subscriber) Discover() {
	topic := fmt.Sprintf("%s/+/+/+/%s", h.rootTopic, propDatatype)
	h.MqttClient.Subscribe(topic, 1, func(c mqtt.Client, msg mqtt.Message) {
		topic := msg.Topic()
		if string(msg.Payload()) == "float" {
			h.discoverDevice(topic)
		} else {
			log.Printf("homie: unsupported datatype - ignoring %s", topic)
		}
	})
}

func (h *Subscriber) discoverDevice(topic string) {
	re, _ := regexp.Compile(`^[a-z0-9]+/([a-z0-9]+)/([a-z0-9]+)/([a-z0-9]+)/?`)
	matches := re.FindStringSubmatch(topic)

	if len(matches) == 4 {
		log.Printf("homie: discovered %s/%s/%s", matches[1], matches[2], matches[3])
		h.mergeDevice(matches[1], matches[2], matches[3])
	} else {
		log.Printf("homie: discovered unexpected device %s", topic)
	}
}

func (h *Subscriber) mergeDevice(deviceName string, nodeName string, propertyName string) {
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
