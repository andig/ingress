package mqtt

import (
	"log"
	"strconv"
	"sync"

	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/data"
	"github.com/eclipse/paho.mqtt.golang"
)

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

	// r := rand.New(rand.NewSource(time.Now().Unix()))
	// for {
	// 	time.Sleep(time.Duration(r.Int31n(1000)) * time.Millisecond)
	// 	data := data.Data{
	// 		Name:  "mqttSample",
	// 		Value: r.Float64(),
	// 	}
	// 	out <- data
	// }
	// panic("not implemented")

	// topic := fmt.Sprintf("%s/+/+/+", h.rootTopic)
	h.MqttClient.Subscribe(h.rootTopic, 1, func(c mqtt.Client, msg mqtt.Message) {
		log.Printf(h.name+": received (%s=%s)", msg.Topic, msg.Payload())

		payload := string(msg.Payload())
		value, err := strconv.ParseFloat(payload, 64)
		if err != nil {
			log.Printf(h.name+": float convesion error, skipping (%s=%s)", msg.Topic, payload)
		}

		data := data.Data{
			Name:  msg.Topic(),
			Value: value,
		}
		out <- data
	})
}
