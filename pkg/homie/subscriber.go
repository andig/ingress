package homie

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/data"
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
	topic := c.Topic
	if topic == "" {
		topic = "homie"
	}

	mqttOptions := mq.NewMqttClientOptions(c.URL, c.User, c.Password)
	homieSubscriber := NewSubscriber(topic, mqttOptions)
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
	log.Println("mqtt: connected to " + mq.ServerFromClient(client))
}

func (h *Subscriber) connectionLostHandler(client mqtt.Client, err error) {
	log.Println("mqtt: disconnected from " + mq.ServerFromClient(client))
}

func (h *Subscriber) Run(out chan data.Data) {
	log.Printf("homie: subscribed to topic %s", h.rootTopic)

	r := rand.New(rand.NewSource(time.Now().Unix()))
	for {
		time.Sleep(time.Duration(r.Int31n(1000)) * time.Millisecond)
		data := data.Data{
			Name:  "homieSample",
			Value: r.Float64(),
		}
		out <- data
	}
	panic("not implemented")

	topic := fmt.Sprintf("%s/+/+/+", h.rootTopic)
	h.MqttClient.Subscribe(topic, 1, func(c mqtt.Client, msg mqtt.Message) {
		log.Printf("homie: received payload %s", msg.Payload())
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
			h.discoverDevice(topic)
		} else {
			log.Printf("homie: unsupported datatype %s - ignoring %s", datatype, topic)
		}
	})
}

func (h *Subscriber) discoverDevice(topic string) {
	segments := strings.Split(topic, "/")

	if len(segments) == 4 {
		log.Printf("homie: discovered %s/%s/%s", segments[1], segments[2], segments[3])
		h.mergeDevice(segments[1], segments[2], segments[3])
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
