package mqtt

import (
	"time"

	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const defaultTimeout = 2000 * time.Millisecond

type mqttConfig = struct {
	config.Target `yaml:",squash"`
	URL           string
	Topic         string
}

func NewMqttClientOptions(url string, user string, password string) *mqtt.ClientOptions {
	if url == "" {
		url = "tcp://localhost:1883"
	}
	mqttOptions := mqtt.NewClientOptions()
	mqttOptions.AddBroker(url)
	mqttOptions.SetUsername(user)
	mqttOptions.SetPassword(password)
	// mqttOptions.SetClientID(mqttClientID)
	// mqttOptions.SetCleanSession(mqttCleanSession)
	mqttOptions.SetAutoReconnect(true)
	return mqttOptions
}

func StripTrailingSlash(s string) string {
	if s[len(s)-1:] == "/" {
		s = s[:len(s)-1]
	}
	return s
}

func ServerFromClient(client mqtt.Client) string {
	options := client.OptionsReader()
	server := options.Servers()[0].String()
	return server
}

type Connector struct {
	MqttClient mqtt.Client
}

func (m *Connector) Connect(mqttClient mqtt.Client) {
	m.MqttClient = mqttClient

	// connect
	if token := m.MqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("mqtt: error connecting: ", token.Error())
	}
}

// WaitForToken returns if an mqtt operation finished within timespan
func (m *Connector) WaitForToken(token mqtt.Token, timeout time.Duration) bool {
	if token.WaitTimeout(timeout) {
		if token.Error() == nil {
			return true
		}
		log.Printf("mqtt: error: %s", token.Error())
	} else {
		log.Printf("mqtt: timeout")
	}
	return false
}
