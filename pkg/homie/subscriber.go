package homie

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

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
	qos       byte
	mux       sync.RWMutex
	props     *PropertySet
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
	return homieSubscriber
}

// NewSubscriber creates Homie/MQTT data source
func NewSubscriber(name string, rootTopic string, mqttOptions *mqtt.ClientOptions) *Subscriber {
	h := &Subscriber{
		MqttConnector: &mq.MqttConnector{},
		name:          name,
		rootTopic:     mq.StripTrailingSlash(rootTopic),
		qos:           1,
		props:         NewPropertySet(),
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
	// start publishing
	h.receiver = out

	// discover homie devices
	topic := fmt.Sprintf("%s/+/+/%s", h.rootTopic, propProperties)
	h.MqttClient.Subscribe(topic, h.qos, func(c mqtt.Client, msg mqtt.Message) {
		// strip $properties
		segments := strings.Split(msg.Topic(), "/")
		topic = strings.Join(segments[:len(segments)-1], "/")
		properties := strings.Split(string(msg.Payload()), ",")
		go h.propertyChangeHandler(topic, properties)
	})
}

// propertyChangeHandler handles changes to node's property definition
func (h *Subscriber) propertyChangeHandler(topic string, properties []string) {
	var wg sync.WaitGroup
	wg.Add(len(properties))

	// add properties
	for _, property := range properties {
		go func(property string) {
			propertyTopic := fmt.Sprintf("%s/%s", topic, property)
			if h.validateProperty(propertyTopic) {
				if h.props.Add(propertyTopic) {
					// print only if not already subscribed
					log.Printf(h.name+": discovered %s", propertyTopic)
					h.subscribeToProperty(propertyTopic)
				}
			}
			wg.Done()
		}(property)
	}

	// wait until properties are merged to remove remaining ones
	wg.Wait()

	// remove obsolete properties
	newProps := NewPropertySet()
	for _, property := range properties {
		newProps.Add(fmt.Sprintf("%s/%s", topic, property))
	}

	nodeProps := h.props.Match(topic + "/")
	for _, old := range nodeProps {
		if !newProps.Contains(old) {
			if h.props.Remove(old) {
				log.Printf(h.name+": removed %s", old)
			}
			h.MqttClient.Unsubscribe(old)
		}
	}
}

// validateProperty collects property definition from $ subtopics
func (h *Subscriber) validateProperty(topic string) bool {
	var mux sync.Mutex
	def := make(map[string][]byte)

	// listen to property definition
	propertyDefinition := topic + "/+"
	h.MqttClient.Subscribe(propertyDefinition, h.qos, func(c mqtt.Client, msg mqtt.Message) {
		mux.Lock()
		defer mux.Unlock()
		def[msg.Topic()] = msg.Payload()
	})

	// wait for timeout according to specification
	select {
	case <-time.After(timeout):
		mux.Lock()
		defer mux.Unlock()
		h.MqttClient.Unsubscribe(propertyDefinition)
	}

	// parse property definition
	if datatype, ok := def[fmt.Sprintf("%s/%s", topic, propDatatype)]; ok {
		return string(datatype) == "float"
	}

	return false
}

func (h *Subscriber) subscribeToProperty(topic string) {
	h.MqttClient.Subscribe(topic, h.qos, func(c mqtt.Client, msg mqtt.Message) {
		log.Printf(h.name+": recv (%s=%s)", msg.Topic(), msg.Payload())

		segments := strings.Split(msg.Topic(), "/")
		name := segments[len(segments)-1]

		payload := string(msg.Payload())
		value, err := strconv.ParseFloat(payload, 64)
		if err != nil {
			log.Printf(h.name+": float conversion error, skipping (%s=%s)", msg.Topic(), payload)
			return
		}

		d := data.Data{
			Name:  name,
			Value: value,
		}

		h.receiver <- d
	})
}
