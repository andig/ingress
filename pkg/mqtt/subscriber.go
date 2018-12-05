package mqtt

import (
	"fmt"
	"log"
	"sync"

	"github.com/andig/ingress/pkg/config"
	"github.com/eclipse/paho.mqtt.golang"
)

type Subscriber struct {
	*MqttConnector
	rootTopic string
	mux       sync.Mutex
	// Devices   []*Device
}

func NewFromInputConfig(c config.Input) *Subscriber {
	mqttOptions := NewMqttClientOptions(c.URL, c.User, c.Password)
	mqttSubscriber := NewSubscriber("#", mqttOptions)
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
	log.Println("mqtt: connected")
}

func (h *Subscriber) connectionLostHandler(client mqtt.Client, err error) {
	log.Println("mqtt: disconnected")
}

func (h *Subscriber) Run() {
	panic("not implemented")

	topic := fmt.Sprintf("%s/+/+/+", h.rootTopic)
	h.MqttClient.Subscribe(topic, 1, func(c mqtt.Client, msg mqtt.Message) {
		log.Printf("homie: received payload %s", msg.Payload())
	})
}
