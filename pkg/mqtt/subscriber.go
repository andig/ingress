package mqtt

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/data"
	"github.com/eclipse/paho.mqtt.golang"
)

type Subscriber struct {
	*MqttConnector
	rootTopic string
	mux       sync.Mutex
	// Devices   []*Device
}

func NewFromInputConfig(c config.Input) *Subscriber {
	topic := c.Topic
	if topic == "" {
		topic = "#"
	}

	mqttOptions := NewMqttClientOptions(c.URL, c.User, c.Password)
	mqttSubscriber := NewSubscriber(topic, mqttOptions)
	mqttClient := mqtt.NewClient(mqttOptions)
	mqttSubscriber.Connect(mqttClient)

	return mqttSubscriber
}

func NewSubscriber(rootTopic string, mqttOptions *mqtt.ClientOptions) *Subscriber {
	h := &Subscriber{
		MqttConnector: &MqttConnector{},
		rootTopic:     StripTrailingSlash(rootTopic),
		// Devices:       []*Device{},
	}

	// connection lost handler
	mqttOptions.SetOnConnectHandler(h.connectionHandler)
	mqttOptions.SetConnectionLostHandler(h.connectionLostHandler)

	return h
}

func (h *Subscriber) connectionHandler(client mqtt.Client) {
	log.Println("mqtt: connected to " + ServerFromClient(client))
}

func (h *Subscriber) connectionLostHandler(client mqtt.Client, err error) {
	log.Println("mqtt: disconnected from " + ServerFromClient(client))
}

func (h *Subscriber) Run(out chan data.Data) {
	log.Printf("mqtt: subscribed to topic %s", h.rootTopic)

	r := rand.New(rand.NewSource(time.Now().Unix()))
	for {
		time.Sleep(time.Duration(r.Int31n(1000)) * time.Millisecond)
		data := data.Data{
			Name:  "mqttSample",
			Value: r.Float64(),
		}
		out <- data
	}
	panic("not implemented")

	topic := fmt.Sprintf("%s/+/+/+", h.rootTopic)
	h.MqttClient.Subscribe(topic, 1, func(c mqtt.Client, msg mqtt.Message) {
		log.Printf("mqtt: received payload %s", msg.Payload())
	})
}
