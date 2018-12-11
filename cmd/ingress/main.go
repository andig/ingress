package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/andig/ingress/pkg/config"
	. "github.com/andig/ingress/pkg/mqtt"
	"github.com/andig/ingress/pkg/wiring"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"gopkg.in/urfave/cli.v1"
)

func inject() {
	mqttOptions := NewMqttClientOptions("tcp://localhost:1883", "", "")
	mqttClient := mqtt.NewClient(mqttOptions)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("mqtt: error connecting: ", token.Error())
	}

	time.Sleep(200 * time.Millisecond)
	token := mqttClient.Publish("input/inject", 0, false, "3.14")
	if token.WaitTimeout(100 * time.Millisecond) {
		log.Println("--> inject done")
	}

	token = mqttClient.Publish("homie/meter1/zaehlwerk1/power", 0, false, "4711")
	if token.WaitTimeout(100 * time.Millisecond) {
		log.Println("--> inject done")
	}
}

func WaitForCtrlC() {
	var wg sync.WaitGroup
	wg.Add(1)
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt, os.Kill)
	go func() {
		<-channel
		wg.Done()
	}()
	wg.Wait()
}

func main() {
	app := cli.NewApp()
	app.Name = "ingress"
	app.Usage = "ingress data mapper daemon"
	app.HideVersion = true
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "config.yml",
			Usage: "Config file",
		},
		cli.BoolFlag{
			Name:  "dump, d",
			Usage: "Dump parsed config",
		},
		cli.BoolFlag{
			Name:  "test, t",
			Usage: "Inject test data",
		},
	}

	app.Action = func(c *cli.Context) {
		if c.NArg() > 0 {
			log.Fatalf("Unexpected arguments: %v", c.Args())
		}

		var conf config.Config
		conf.Load(c.String("config"))

		if c.Bool("dump") {
			conf.Dump()
		}

		connectors := wiring.NewConnectors(conf.Sources, conf.Targets)
		mappings := wiring.NewMappings(conf.Mappings, connectors)
		wires := wiring.NewWiring(conf.Wires, mappings, connectors)
		mapper := wiring.NewMapper(wires, connectors)
		go connectors.Run(mapper)

		if c.Bool("test") {
			inject()
		}

		// time.Sleep(3 * time.Second)
		WaitForCtrlC()
	}

	app.Run(os.Args)
}
