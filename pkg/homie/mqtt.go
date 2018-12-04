package homie

import (
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func stripTrailingSlash(s string) string {
	if s[len(s)-1:] == "/" {
		s = s[:len(s)-1]
	}
	return s
}

type MqttConnector struct {
	mqttClient mqtt.Client
}

func (m *MqttConnector) Connect(mqttClient mqtt.Client) {
	m.mqttClient = mqttClient

	// connect
	if token := m.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("MQTT: error connecting: ", token.Error())
		panic(token.Error())
	}
}

func (m *MqttConnector) WaitForToken(token mqtt.Token) {
	if token.WaitTimeout(2000 * time.Millisecond) {
		if token.Error() != nil {
			log.Printf("MQTT: error: %s", token.Error())
		}
	} else {
		// if m.verbose {
		log.Printf("MQTT: timeout")
		// }
	}
}
