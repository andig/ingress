package mqtt

import (
	"log"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
)

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

type MqttConnector struct {
	MqttClient mqtt.Client
}

func (m *MqttConnector) Connect(mqttClient mqtt.Client) {
	m.MqttClient = mqttClient

	// connect
	if token := m.MqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("mqtt: error connecting: ", token.Error())
	}
}

func (m *MqttConnector) WaitForToken(token mqtt.Token) {
	if token.WaitTimeout(2000 * time.Millisecond) {
		if token.Error() != nil {
			log.Printf("mqtt: error: %s", token.Error())
		}
	} else {
		// if m.verbose {
		log.Printf("mqtt: timeout")
		// }
	}
}
