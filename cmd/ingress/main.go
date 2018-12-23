package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/andig/ingress/pkg/config"
	. "github.com/andig/ingress/pkg/mqtt"
	"github.com/andig/ingress/pkg/wiring"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/tcnksm/go-latest"
	"github.com/sirupsen/logrus"
	"gopkg.in/urfave/cli.v1"
)

var log *logrus.Entry

func inject() {
	mqttOptions := NewMqttClientOptions("tcp://localhost:1883", "", "")
	mqttClient := mqtt.NewClient(mqttOptions)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("mqtt: error connecting: %s", token.Error())
	}

	time.Sleep(200 * time.Millisecond)
	token := mqttClient.Publish("input/inject", 0, false, "3.14")
	if token.WaitTimeout(100 * time.Millisecond) {
		log.Println("--> inject done")
	}

	mqttClient.Publish("homie/meter1/$nodes", 1, true, "zaehlwerk1")
	mqttClient.Publish("homie/meter1/zaehlwerk1/$properties", 1, true, "power")
	mqttClient.Publish("homie/meter1/zaehlwerk1/power/$name", 1, true, "Leistung")
	mqttClient.Publish("homie/meter1/zaehlwerk1/power/$unit", 1, true, "W")
	mqttClient.Publish("homie/meter1/zaehlwerk1/power/$datatype", 1, true, "float")
	mqttClient.Publish("homie/meter1/zaehlwerk1/power", 1, false, "3048")
}

func checkVersion() {
	githubTag := &latest.GithubTag{
		Owner:      "andig",
		Repository: "ingress",
	}

	if res, err := latest.Check(githubTag, tag); err == nil {
		if res.Outdated {
			log.Printf("updates available - please upgrade to ingress %s", res.Current)
		}
	}
}

func setLogLevel(level string) {
	switch strings.ToLower(level) {
	case "error": logrus.SetLevel(logrus.ErrorLevel)
	case "info": logrus.SetLevel(logrus.InfoLevel)
	case "debug": logrus.SetLevel(logrus.DebugLevel)
	case "trace": logrus.SetLevel(logrus.TraceLevel)
	default: logrus.SetLevel(logrus.DebugLevel)
	}
}

func waitForCtrlC() {
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
	app.Version = fmt.Sprintf("%s (https://github.com/andig/ingress/commit/%s)", tag, hash)
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
			Name:  "diagnose",
			Usage: "Memory diagnostics",
		},
		cli.StringFlag{
			Name:  "log",
			Value: "debug",
			Usage: "Log level (error, info, debug, trace)",
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

		setLogLevel(c.String("log"))

		go checkVersion()

		connectors := wiring.NewConnectors(conf.Sources, conf.Targets)
		actions := wiring.NewActions(conf.Actions)
		mappings := wiring.NewMappings(conf.Mappings, connectors)
		wires := wiring.NewWiring(conf.Wires, mappings, connectors)
		mapper := wiring.NewMapper(wires, connectors)
		_ = actions
		_ = mapper
		go connectors.Run(mapper)

		if c.Bool("test") {
			inject()
		}

		if c.Bool("diagnose") {
			go func() {
				for {
					time.Sleep(time.Second)
					var memstats runtime.MemStats
					runtime.ReadMemStats(&memstats)
					fmt.Printf("%db\n", memstats.Alloc)
				}
			}()
		}

		waitForCtrlC()
	}

	app.Run(os.Args)
}
